//
//  ISBinaryImageHelper.mm
//  Sample
//
//  Created by 舒彪 on 2018/4/30.
//  Copyright © 2018年 intsig. All rights reserved.
//

#include "ISBinaryImageHelper.h"
#include <mach-o/dyld.h>
#import <Foundation/Foundation.h>
#include <vector>

static std::vector<ISBinaryImageInfoRef> g_imagesInfo;

ISBinaryImageInfoRef is_createNewBinaryImageInfo() {
    ISBinaryImageInfoRef newInfo = (ISBinaryImageInfoRef)malloc(sizeof(ISBinaryImageInfo));
    newInfo->imageName = NULL;
    newInfo->uuid = NULL;
    newInfo->baseAddress = NULL;
    newInfo->maxAddress = NULL;
    return newInfo;
}

uintptr_t is_firstCmdAfterHeader(const struct mach_header* const header) {
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

void initBinaryImagesInfo() {
    if (g_imagesInfo.size() > 0) {
        return;
    }
    
    const uint32_t imageCount = _dyld_image_count();
    const struct mach_header* header = 0;
    
    uint64_t maxAddress = 0;
    uint64_t minAddress = 0;
    
    for(uint32_t iImg = 0; iImg < imageCount; iImg++) {
        header = _dyld_get_image_header(iImg);
        if(header != NULL) {
            uintptr_t cmdPtr = is_firstCmdAfterHeader(header);
            if(cmdPtr == 0) {
                continue;
            }
            
            ISBinaryImageInfoRef newImageInfo = is_createNewBinaryImageInfo();
            
            for(uint32_t iCmd = 0; iCmd < header->ncmds; iCmd++) {
                const struct load_command *loadCmd = (struct load_command *)cmdPtr;
                switch(loadCmd->cmd)
                {
                    case LC_SEGMENT:
                    {
                        const struct segment_command *segCmd = (struct segment_command *)cmdPtr;
                        uintptr_t maxSegmentAddress = segCmd->vmaddr + segCmd->vmsize;
                        if (maxSegmentAddress > newImageInfo->maxAddress) {
                            newImageInfo->maxAddress = maxSegmentAddress;
                        }
                        break;
                    }
                    case LC_SEGMENT_64:
                    {
                        const struct segment_command_64 *segCmd = (struct segment_command_64 *)cmdPtr;
                        uintptr_t maxSegmentAddress = segCmd->vmaddr + segCmd->vmsize;
                        if (maxSegmentAddress > newImageInfo->maxAddress) {
                            newImageInfo->maxAddress = maxSegmentAddress;
                        }
                        break;
                    }
                    case LC_UUID:
                    {
                        struct uuid_command *uuidCmd = (struct uuid_command *)cmdPtr;
                        CFUUIDRef uuidRef = CFUUIDCreateFromUUIDBytes(NULL, *((CFUUIDBytes*)uuidCmd->uuid));
                        CFStringRef str = CFUUIDCreateString(NULL, uuidRef);
                        CFRelease(uuidRef);
                        newImageInfo->uuid = CFStringGetCStringPtr(str, kCFStringEncodingUTF8);
                        break;
                    }
                    default:
                        break;
                }
                cmdPtr += loadCmd->cmdsize;
            }
            if (newImageInfo->maxAddress && newImageInfo->uuid) {
                newImageInfo->maxAddress += (uintptr_t)_dyld_get_image_vmaddr_slide(iImg);
                newImageInfo->baseAddress = (uintptr_t)header;
                NSString *imageName =
                [NSString stringWithCString:_dyld_get_image_name((unsigned)iImg) encoding:NSUTF8StringEncoding];
                char *imageNameBuffer = (char *)calloc(50, sizeof(char));
                [imageName.lastPathComponent getCString:imageNameBuffer
                                              maxLength:50
                                               encoding:NSUTF8StringEncoding];
                newImageInfo->imageName = imageNameBuffer;
            }
            g_imagesInfo.push_back(newImageInfo);
            maxAddress = MAX(newImageInfo->maxAddress, maxAddress);
            minAddress = MIN(newImageInfo->baseAddress, minAddress);
        }
    }
    printf("maxAddress: %llu minaddress: %llu\n", maxAddress, minAddress);
}

ISBinaryImageInfoRef imageContainesAddressImp(uintptr_t address) {
    for (std::vector<ISBinaryImageInfoRef>::iterator it = g_imagesInfo.begin() ;
         it != g_imagesInfo.end();
         ++it) {
        ISBinaryImageInfoRef image = (ISBinaryImageInfoRef)*it;
        if (address >= image->baseAddress && address < image->maxAddress) {
            return image;
        }
    }
    return NULL;
}

ISBinaryImageInfoRef imageContainesAddress(uintptr_t address) {
    if (g_imagesInfo.size() == 0) {
        initBinaryImagesInfo();
    }
    
    auto ref = imageContainesAddressImp(address);
    
    if (ref) {
        return ref;
    }
    
    initBinaryImagesInfo();
    
    return imageContainesAddressImp(address);
}

ISBinaryImageInfoRef imageOfName(const char *name) {
    if (g_imagesInfo.size() == 0) {
        initBinaryImagesInfo();
    }
    
    for (std::vector<ISBinaryImageInfoRef>::iterator it = g_imagesInfo.begin() ;
         it != g_imagesInfo.end();
         ++it) {
        ISBinaryImageInfoRef image = (ISBinaryImageInfoRef)*it;
        if (strcmp(name, image->imageName) == 0) {
            return image;
        }
    }
    return NULL;
}
