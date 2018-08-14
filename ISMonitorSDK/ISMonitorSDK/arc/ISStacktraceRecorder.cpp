//
//  ISStacktraceRecorder.cpp
//  ANRTest
//
//  Created by Brent Shu on 2018/8/1.
//  Copyright © 2018年 intsig. All rights reserved.
//

#include "ISStacktraceRecorder.hpp"
#include <dlfcn.h>
#include <pthread.h>
#include <sys/types.h>
#include <limits.h>
#include <mach-o/dyld.h>
#include <mach-o/nlist.h>
#include <mach/mach.h>
#include <dispatch/dispatch.h>

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

namespace ISBSRecorder {
    
    static const int maxRecordThreadCount = 100;
    static const int maxFramesCountPerStack = 100;
    static const int maxThreadNamelength = 50;
    
    struct FrameEntry {
        const struct FrameEntry *const previous;
        const uintptr_t return_address;
    };
    
#pragma -mark HandleMachineContext
    bool fillThreadStateIntoMachineContext(thread_t thread, _STRUCT_MCONTEXT *machineContext) {
        mach_msg_type_number_t state_count = BS_THREAD_STATE_COUNT;
        kern_return_t kr = thread_get_state(thread, BS_THREAD_STATE,
                                            (thread_state_t)&machineContext->__ss,
                                            &state_count);
        return (kr == KERN_SUCCESS);
    }
    
    uintptr_t mach_framePointer(mcontext_t const machineContext) {
        return machineContext->__ss.BS_FRAME_POINTER;
    }
    
    uintptr_t mach_stackPointer(mcontext_t const machineContext) {
        return machineContext->__ss.BS_STACK_POINTER;
    }
    
    uintptr_t mach_instructionAddress(mcontext_t const machineContext) {
        return machineContext->__ss.BS_INSTRUCTION_ADDRESS;
    }
    
    uintptr_t mach_linkRegister(mcontext_t const machineContext) {
#if defined(__i386__) || defined(__x86_64__)
        return 0;
#else
        return machineContext->__ss.__lr;
#endif
    }
    
    kern_return_t mach_copyMem(const void *const src, void *const dst, const size_t numBytes) {
        vm_size_t bytesCopied = 0;
        return vm_read_overwrite(mach_task_self(),
                                 (vm_address_t)src,
                                 (vm_size_t)numBytes,
                                 (vm_address_t)dst,
                                 &bytesCopied);
    }
    
    bool getQueueName(const thread_t thread, char* const buffer, int bufLength) {
        integer_t infoBuffer[THREAD_IDENTIFIER_INFO_COUNT] = {0};
        thread_info_t info = infoBuffer;
        mach_msg_type_number_t inOutSize = THREAD_IDENTIFIER_INFO_COUNT;
        kern_return_t kr = 0;
        
        kr = thread_info((thread_t)thread, THREAD_IDENTIFIER_INFO, info, &inOutSize);
        if(kr != KERN_SUCCESS)
        {
            return false;
        }
        
        thread_identifier_info_t idInfo = (thread_identifier_info_t)info;
        
        dispatch_queue_t* dispatch_queue_ptr = (dispatch_queue_t*)idInfo->dispatch_qaddr;
        
        if(dispatch_queue_ptr == NULL || idInfo->thread_handle == 0 || *dispatch_queue_ptr == NULL)
        {
            return false;
        }
        
        dispatch_queue_t dispatch_queue = *dispatch_queue_ptr;
        const char* queue_name = dispatch_queue_get_label(dispatch_queue);
        if(queue_name == NULL)
        {
            return false;
        }
        
        int length = (int)strlen(queue_name);
        
        int iLabel;
        for(iLabel = 0; iLabel < length + 1; iLabel++)
        {
            if(queue_name[iLabel] < ' ' || queue_name[iLabel] > '~')
            {
                break;
            }
        }
        if(queue_name[iLabel] != 0)
        {
            return false;
        }
        bufLength = std::min(length, bufLength - 1);
        strncpy(buffer, queue_name, bufLength);
        buffer[bufLength] = 0;
        return true;
    }
    
    inline static thread_t thread_self() {
        thread_t thread_self = mach_thread_self();
        mach_port_deallocate(mach_task_self(), thread_self);
        return thread_self;
    }
    
    static void suspendOtherThreads() {
        kern_return_t kr;
        const task_t thisTask = mach_task_self();
        const thread_t thisThread = (thread_t)thread_self();
        thread_act_array_t threads;
        mach_msg_type_number_t numThreads;
        
        if((kr = task_threads(thisTask, &threads, &numThreads)) != KERN_SUCCESS)
        {
            return;
        }
        
        for(mach_msg_type_number_t i = 0; i < numThreads; i++)
        {
            thread_t thread = threads[i];
            if(thread != thisThread)
            {
                if((kr = thread_suspend(thread)) != KERN_SUCCESS)
                {
                }
            }
        }
        
        for(mach_msg_type_number_t i = 0; i < numThreads; i++)
        {
            mach_port_deallocate(thisTask, threads[i]);
        }
        vm_deallocate(thisTask, (vm_address_t)threads, sizeof(thread_t) * numThreads);
    }
    
    static void resumeOtherThreads() {
        kern_return_t kr;
        const task_t thisTask = mach_task_self();
        const thread_t thisThread = (thread_t)thread_self();
        thread_act_array_t threads;
        mach_msg_type_number_t numThreads;
        
        if((kr = task_threads(thisTask, &threads, &numThreads)) != KERN_SUCCESS)
        {
            return;
        }
        
        for(mach_msg_type_number_t i = 0; i < numThreads; i++)
        {
            thread_t thread = threads[i];
            if(thread != thisThread)
            {
                if((kr = thread_resume(thread)) != KERN_SUCCESS)
                {
                }
            }
        }
        
        for(mach_msg_type_number_t i = 0; i < numThreads; i++)
        {
            mach_port_deallocate(thisTask, threads[i]);
        }
        vm_deallocate(thisTask, (vm_address_t)threads, sizeof(thread_t) * numThreads);
    }
    
#pragma -mark Interface
    status backtraceOfThread(thread_t thread, Stack & stack) {
        _STRUCT_MCONTEXT machineContext;
        if(!fillThreadStateIntoMachineContext(thread, &machineContext)) {
            return status_error;
        }
        
        const uintptr_t instructionAddress = mach_instructionAddress(&machineContext);
        if(instructionAddress == 0) {
            return status_error;
        }
        
        stack.frames.push_back(instructionAddress);
        
        uintptr_t linkRegister = mach_linkRegister(&machineContext);
        if (linkRegister) {
            stack.frames.push_back(linkRegister);
        }
        
        FrameEntry frame = {0};
        const uintptr_t framePtr = mach_framePointer(&machineContext);
        if(framePtr == 0 ||
           mach_copyMem((void *)framePtr, &frame, sizeof(frame)) != KERN_SUCCESS) {
            return status_error;
        }
        
        while (stack.frames.size() < maxFramesCountPerStack) {
            uintptr_t ret_addr = frame.return_address;
            if(ret_addr == 0 || frame.previous == 0) {
                break;
            }
            stack.frames.push_back(ret_addr);
            
            if (mach_copyMem(frame.previous, &frame, sizeof(frame)) != KERN_SUCCESS) {
                break;
            }
        }
        return status_ok;
    }
    
    status backtraceOfCurrentThread(Stack & stack) {
        return backtraceOfThread(mach_thread_self(), stack);
    }
    
    void backtraceOfAllThread(Stacks & stacks) {
        Stacks tmpStack;
        for (int i = 0; i < maxRecordThreadCount; ++i) {
            Stack newStack;
            tmpStack.push_back(newStack);
            tmpStack[i].threadName.reserve(maxThreadNamelength);
            tmpStack[i].frames.reserve(maxFramesCountPerStack);
        }
        stacks.reserve(maxRecordThreadCount);
        
        suspendOtherThreads();
        thread_act_array_t threads;
        mach_msg_type_number_t thread_count = 0;
        const task_t this_task = mach_task_self();
        
        kern_return_t kr = task_threads(this_task, &threads, &thread_count);
        if(kr != KERN_SUCCESS) {
            resumeOtherThreads();
            return;
        }
        
        for(int i = 0; i < std::min((int)thread_count, maxRecordThreadCount); ++i) {
            char threadName[maxThreadNamelength];
            if (getQueueName(threads[i], threadName, maxThreadNamelength)) {
                tmpStack[i].threadName.assign(threadName);
            }
            
            if (backtraceOfThread(threads[i], tmpStack[i]) == status_ok) {
                stacks.push_back(std::move(tmpStack[i]));
            }
        }
        resumeOtherThreads();
    }
}


