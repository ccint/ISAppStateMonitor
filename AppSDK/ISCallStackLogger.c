//
//  ISCallStackLogger.cpp
//  CamCard
//
//  Created by Brent Shu on 2018/3/16.
//  Copyright © 2018年 舒彪. All rights reserved.
//

#include "ISCallStackLogger.h"
#include <dlfcn.h>
#include <pthread.h>
#include <sys/types.h>
#include <limits.h>
#include <string.h>
#include <mach-o/dyld.h>
#include <mach-o/nlist.h>

#pragma -mark DEFINE MACRO FOR DIFFERENT CPU ARCHITECTURE
#if defined(__arm64__)
#define DETAG_INSTRUCTION_ADDRESS(A) ((A) & ~(3UL))
#define BS_THREAD_STATE_COUNT ARM_THREAD_STATE64_COUNT
#define BS_THREAD_STATE ARM_THREAD_STATE64
#define BS_FRAME_POINTER __fp
#define BS_STACK_POINTER __sp
#define BS_INSTRUCTION_ADDRESS __pc

#elif defined(__arm__)
#define DETAG_INSTRUCTION_ADDRESS(A) ((A) & ~(1UL))
#define BS_THREAD_STATE_COUNT ARM_THREAD_STATE_COUNT
#define BS_THREAD_STATE ARM_THREAD_STATE
#define BS_FRAME_POINTER __r[7]
#define BS_STACK_POINTER __sp
#define BS_INSTRUCTION_ADDRESS __pc

#elif defined(__x86_64__)
#define DETAG_INSTRUCTION_ADDRESS(A) (A)
#define BS_THREAD_STATE_COUNT x86_THREAD_STATE64_COUNT
#define BS_THREAD_STATE x86_THREAD_STATE64
#define BS_FRAME_POINTER __rbp
#define BS_STACK_POINTER __rsp
#define BS_INSTRUCTION_ADDRESS __rip

#elif defined(__i386__)
#define DETAG_INSTRUCTION_ADDRESS(A) (A)
#define BS_THREAD_STATE_COUNT x86_THREAD_STATE32_COUNT
#define BS_THREAD_STATE x86_THREAD_STATE32
#define BS_FRAME_POINTER __ebp
#define BS_STACK_POINTER __esp
#define BS_INSTRUCTION_ADDRESS __eip

#endif

#define CALL_INSTRUCTION_FROM_RETURN_ADDRESS(A) (DETAG_INSTRUCTION_ADDRESS((A)) - 1)

#if defined(__LP64__)
#define TRACE_FMT         "%-4d%-31s 0x%016lx %s + %lu"
#define POINTER_FMT       "0x%016lx"
#define POINTER_SHORT_FMT "0x%lx"
#define BS_NLIST struct nlist_64
#else
#define TRACE_FMT         "%-4d%-31s 0x%08lx %s + %lu"
#define POINTER_FMT       "0x%08lx"
#define POINTER_SHORT_FMT "0x%lx"
#define BS_NLIST struct nlist
#endif

typedef struct BSStackFrameEntry{
  const struct BSStackFrameEntry *const previous;
  const uintptr_t return_address;
} BSStackFrameEntry;

bool bs_fillThreadStateIntoMachineContext(thread_t thread, _STRUCT_MCONTEXT *machineContext);

bool bs_dladdr(const uintptr_t address, Dl_info* const info);

uintptr_t bs_mach_framePointer(mcontext_t const machineContext);

uintptr_t bs_mach_stackPointer(mcontext_t const machineContext);

uintptr_t bs_mach_instructionAddress(mcontext_t const machineContext);

uintptr_t bs_mach_linkRegister(mcontext_t const machineContext);

kern_return_t bs_mach_copyMem(const void *const src, void *const dst, const size_t numBytes);

void bs_symbolicate(const uintptr_t* const backtraceBuffer,
                    Dl_info* const symbolsBuffer,
                    const int numEntries,
                    const int skippedEntries);

uintptr_t bs_firstCmdAfterHeader(const struct mach_header* const header);

uint32_t bs_imageIndexContainingAddress(const uintptr_t address);

uintptr_t bs_segmentBaseOfImageIndex(const uint32_t idx);

#pragma -mark HandleMachineContext
bool bs_fillThreadStateIntoMachineContext(thread_t thread, _STRUCT_MCONTEXT *machineContext) {
  mach_msg_type_number_t state_count = BS_THREAD_STATE_COUNT;
  kern_return_t kr = thread_get_state(thread, BS_THREAD_STATE, (thread_state_t)&machineContext->__ss, &state_count);
  return (kr == KERN_SUCCESS);
}

uintptr_t bs_mach_framePointer(mcontext_t const machineContext) {
  return machineContext->__ss.BS_FRAME_POINTER;
}

uintptr_t bs_mach_stackPointer(mcontext_t const machineContext) {
  return machineContext->__ss.BS_STACK_POINTER;
}

uintptr_t bs_mach_instructionAddress(mcontext_t const machineContext) {
  return machineContext->__ss.BS_INSTRUCTION_ADDRESS;
}

uintptr_t bs_mach_linkRegister(mcontext_t const machineContext) {
#if defined(__i386__) || defined(__x86_64__)
  return 0;
#else
  return machineContext->__ss.__lr;
#endif
}

kern_return_t bs_mach_copyMem(const void *const src, void *const dst, const size_t numBytes) {
  vm_size_t bytesCopied = 0;
  return vm_read_overwrite(mach_task_self(), (vm_address_t)src, (vm_size_t)numBytes, (vm_address_t)dst, &bytesCopied);
}

#pragma -mark Symbolicate
void bs_symbolicate(const uintptr_t* const backtraceBuffer,
                    Dl_info* const symbolsBuffer,
                    const int numEntries,
                    const int skippedEntries) {
  int i = 0;
  
  if(!skippedEntries && i < numEntries) {
    bs_dladdr(backtraceBuffer[i], &symbolsBuffer[i]);
    i++;
  }
  
  for(; i < numEntries; i++) {
    bs_dladdr(CALL_INSTRUCTION_FROM_RETURN_ADDRESS(backtraceBuffer[i]), &symbolsBuffer[i]);
  }
}

bool bs_dladdr(const uintptr_t address, Dl_info* const info) {
  info->dli_fname = NULL;
  info->dli_fbase = NULL;
  info->dli_sname = NULL;
  info->dli_saddr = NULL;
  
  const uint32_t idx = bs_imageIndexContainingAddress(address);
  if(idx == UINT_MAX) {
    return false;
  }
  const struct mach_header* header = _dyld_get_image_header(idx);
  const uintptr_t imageVMAddrSlide = (uintptr_t)_dyld_get_image_vmaddr_slide(idx);
  const uintptr_t addressWithSlide = address - imageVMAddrSlide;
  const uintptr_t segmentBase = bs_segmentBaseOfImageIndex(idx) + imageVMAddrSlide;
  if(segmentBase == 0) {
    return false;
  }
  
  info->dli_fname = _dyld_get_image_name(idx);
  info->dli_fbase = (void*)header;
  
  // Find symbol tables and get whichever symbol is closest to the address.
  const BS_NLIST* bestMatch = NULL;
  uintptr_t bestDistance = ULONG_MAX;
  uintptr_t cmdPtr = bs_firstCmdAfterHeader(header);
  if(cmdPtr == 0) {
    return false;
  }
  for(uint32_t iCmd = 0; iCmd < header->ncmds; iCmd++) {
    const struct load_command* loadCmd = (struct load_command*)cmdPtr;
    if(loadCmd->cmd == LC_SYMTAB) {
      const struct symtab_command* symtabCmd = (struct symtab_command*)cmdPtr;
      const BS_NLIST* symbolTable = (BS_NLIST*)(segmentBase + symtabCmd->symoff);
      const uintptr_t stringTable = segmentBase + symtabCmd->stroff;
      
      for(uint32_t iSym = 0; iSym < symtabCmd->nsyms; iSym++) {
        // If n_value is 0, the symbol refers to an external object.
        if(symbolTable[iSym].n_value != 0) {
          uintptr_t symbolBase = symbolTable[iSym].n_value;
          uintptr_t currentDistance = addressWithSlide - symbolBase;
          if((addressWithSlide >= symbolBase) &&
             (currentDistance <= bestDistance)) {
            bestMatch = symbolTable + iSym;
            bestDistance = currentDistance;
          }
        }
      }
      if(bestMatch != NULL) {
        info->dli_saddr = (void*)(bestMatch->n_value + imageVMAddrSlide);
        info->dli_sname = (char*)((intptr_t)stringTable + (intptr_t)bestMatch->n_un.n_strx);
        if(*info->dli_sname == '_') {
          info->dli_sname++;
        }
        // This happens if all symbols have been stripped.
        if(info->dli_saddr == info->dli_fbase && bestMatch->n_type == 3) {
          info->dli_sname = NULL;
        }
        break;
      }
    }
    cmdPtr += loadCmd->cmdsize;
  }
  return true;
}

uintptr_t bs_firstCmdAfterHeader(const struct mach_header* const header) {
  switch(header->magic) {
    case MH_MAGIC:
    case MH_CIGAM:
      return (uintptr_t)(header + 1);
    case MH_MAGIC_64:
    case MH_CIGAM_64:
      return (uintptr_t)(((struct mach_header_64*)header) + 1);
    default:
      return 0;  // Header is corrupt
  }
}

uint32_t bs_imageIndexContainingAddress(const uintptr_t address) {
  const uint32_t imageCount = _dyld_image_count();
  const struct mach_header* header = 0;
  
  for(uint32_t iImg = 0; iImg < imageCount; iImg++) {
    header = _dyld_get_image_header(iImg);
    if(header != NULL) {
      // Look for a segment command with this address within its range.
      uintptr_t addressWSlide = address - (uintptr_t)_dyld_get_image_vmaddr_slide(iImg);
      uintptr_t cmdPtr = bs_firstCmdAfterHeader(header);
      if(cmdPtr == 0) {
        continue;
      }
      for(uint32_t iCmd = 0; iCmd < header->ncmds; iCmd++) {
        const struct load_command* loadCmd = (struct load_command*)cmdPtr;
        if(loadCmd->cmd == LC_SEGMENT) {
          const struct segment_command* segCmd = (struct segment_command*)cmdPtr;
          if(addressWSlide >= segCmd->vmaddr &&
             addressWSlide < segCmd->vmaddr + segCmd->vmsize) {
            return iImg;
          }
        }
        else if(loadCmd->cmd == LC_SEGMENT_64) {
          const struct segment_command_64* segCmd = (struct segment_command_64*)cmdPtr;
          if(addressWSlide >= segCmd->vmaddr &&
             addressWSlide < segCmd->vmaddr + segCmd->vmsize) {
            return iImg;
          }
        }
        cmdPtr += loadCmd->cmdsize;
      }
    }
  }
  return UINT_MAX;
}

uintptr_t bs_segmentBaseOfImageIndex(const uint32_t idx) {
  const struct mach_header* header = _dyld_get_image_header(idx);
  
  // Look for a segment command and return the file image address.
  uintptr_t cmdPtr = bs_firstCmdAfterHeader(header);
  if(cmdPtr == 0) {
    return 0;
  }
  for(uint32_t i = 0;i < header->ncmds; i++) {
    const struct load_command* loadCmd = (struct load_command*)cmdPtr;
    if(loadCmd->cmd == LC_SEGMENT) {
      const struct segment_command* segmentCmd = (struct segment_command*)cmdPtr;
      if(strcmp(segmentCmd->segname, SEG_LINKEDIT) == 0) {
        return segmentCmd->vmaddr - segmentCmd->fileoff;
      }
    }
    else if(loadCmd->cmd == LC_SEGMENT_64) {
      const struct segment_command_64* segmentCmd = (struct segment_command_64*)cmdPtr;
      if(strcmp(segmentCmd->segname, SEG_LINKEDIT) == 0) {
        return (uintptr_t)(segmentCmd->vmaddr - segmentCmd->fileoff);
      }
    }
    cmdPtr += loadCmd->cmdsize;
  }
  return 0;
}

#pragma -mark Interface
bool bs_backtraceOfThread(thread_t thread,
                          uintptr_t *buffer,
                          int32_t bufferMaxSize,
                          int32_t *bufferSize) {
  if (bufferMaxSize < 2) {
    return false;
  }
  
  int i = 0;
  
  _STRUCT_MCONTEXT machineContext;
  if(!bs_fillThreadStateIntoMachineContext(thread, &machineContext)) {
    return false;
  }
  
  const uintptr_t instructionAddress = bs_mach_instructionAddress(&machineContext);
  buffer[i] = instructionAddress;
  ++i;
  
  uintptr_t linkRegister = bs_mach_linkRegister(&machineContext);
  if (linkRegister) {
    buffer[i] = linkRegister;
    i++;
  }
  
  if(instructionAddress == 0) {
    return false;
  }
  
  BSStackFrameEntry frame = {0};
  const uintptr_t framePtr = bs_mach_framePointer(&machineContext);
  if(framePtr == 0 ||
     bs_mach_copyMem((void *)framePtr, &frame, sizeof(frame)) != KERN_SUCCESS) {
    return false;
  }
  
  int32_t bufferLen = i;
  for(; i < bufferMaxSize; ++i) {
    buffer[i] = frame.return_address;
    bufferLen = i + 1;
    if(buffer[i] == 0 ||
       frame.previous == 0 ||
       bs_mach_copyMem(frame.previous, &frame, sizeof(frame)) != KERN_SUCCESS) {
      bufferLen = buffer[i] == 0 ? i : bufferLen;
      break;
    }
  }
  
  *bufferSize = bufferLen;
  return true;
}

bool bs_backtraceOfCurrentThread(uintptr_t *buffer,
                                 int32_t bufferMaxSize,
                                 int32_t *bufferSize) {
  return bs_backtraceOfThread(mach_thread_self(),
                              buffer,
                              bufferMaxSize,
                              bufferSize);
}

bool bs_backtraceOfAllThread(uintptr_t **allThreadBuffer,
                             int32_t bufferMaxSize,
                             int32_t maxThreadCount,
                             int32_t *buffersSize,
                             int32_t *threadCount,
                             char **threadsName) {
  thread_act_array_t threads;
  mach_msg_type_number_t thread_count = 0;
  const task_t this_task = mach_task_self();
  
  kern_return_t kr = task_threads(this_task, &threads, &thread_count);
  if(kr != KERN_SUCCESS) {
    return false;
  }
  
  *threadCount = thread_count;
  
  for(int i = 0; i < MIN(thread_count, maxThreadCount); i++) {
    pthread_t pThread = pthread_from_mach_thread_np(threads[i]);
    pthread_getname_np(pThread, threadsName[i], 100);
    bs_backtraceOfThread(threads[i], allThreadBuffer[i], bufferMaxSize, buffersSize + i);
  }
  return true;
}
