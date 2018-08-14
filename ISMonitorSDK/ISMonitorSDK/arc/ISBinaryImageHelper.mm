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

namespace ISBinaryImage {
    static std::vector<ISBinaryImageInfo *> g_imagesInfo;
    
    ISBinaryImageInfo::ISBinaryImageInfo()
    :imageName(nullptr)
    ,uuid(nullptr)
    ,baseAddress(0)
    ,segmentRanges()
    ,slide(0)
    {
    }
    
    ISBinaryImageInfo::~ISBinaryImageInfo() {
        free((char *)imageName);
        free((char *)uuid);
    }
    
    bool ISBinaryImageInfo::containesAddress(uintptr_t address) {
        if (address < this->baseAddress) {
            return false;
        }
        
        for (auto it = this->segmentRanges.begin();
             it != this->segmentRanges.end();
             ++it) {
            auto pair = *it;
            uintptr_t addressWSlide = address - this->slide;
            if (addressWSlide >= pair.first && addressWSlide < pair.second) {
                return true;
            }
        }
        return false;
    }
    
    uintptr_t firstCmdAfterHeader(const struct mach_header* const header) {
        switch(header->magic)
        {
            case MH_MAGIC:
            case MH_CIGAM:
                return (uintptr_t)(header + 1);
            case MH_MAGIC_64:
            case MH_CIGAM_64:
                return (uintptr_t)(((struct mach_header_64*)header) + 1);
            default:
                // Header is corrupt
                return 0;
        }
    }
    
    void initBinaryImagesInfo() {
        if (g_imagesInfo.size() > 0) {
            return ;
        }
        
        const uint32_t imageCount = _dyld_image_count();
        const struct mach_header* header = 0;
        
        for(uint32_t iImg = 0; iImg < imageCount; iImg++) {
            header = _dyld_get_image_header(iImg);
            if(header != NULL) {
                uintptr_t cmdPtr = firstCmdAfterHeader(header);
                if(cmdPtr == 0) {
                    continue;
                }
                
                ISBinaryImageInfo *newImageInfo = new ISBinaryImageInfo();
                
                for(uint32_t iCmd = 0; iCmd < header->ncmds; iCmd++) {
                    const struct load_command *loadCmd = (struct load_command *)cmdPtr;
                    switch(loadCmd->cmd)
                    {
                        case LC_SEGMENT:
                        {
                            const struct segment_command *segCmd = (struct segment_command *)cmdPtr;
                            newImageInfo->segmentRanges.push_back({segCmd->vmaddr, segCmd->vmaddr + segCmd->vmsize});
                            break;
                        }
                        case LC_SEGMENT_64:
                        {
                            const struct segment_command_64 *segCmd = (struct segment_command_64 *)cmdPtr;
                            newImageInfo->segmentRanges.push_back({segCmd->vmaddr, segCmd->vmaddr + segCmd->vmsize});
                            break;
                        }
                        case LC_UUID:
                        {
                            struct uuid_command *uuidCmd = (struct uuid_command *)cmdPtr;
                            CFUUIDRef uuidRef = CFUUIDCreateFromUUIDBytes(NULL, *((CFUUIDBytes*)uuidCmd->uuid));
                            CFStringRef str = CFUUIDCreateString(NULL, uuidRef);
                            CFRelease(uuidRef);
                            auto strLen = CFStringGetLength(str);
                            char *uuidBuffer = (char *)calloc(strLen + 1, sizeof(char));
                            CFStringGetCString(str, uuidBuffer, strLen + 1, kCFStringEncodingUTF8);
                            CFRelease(str);
                            newImageInfo->uuid = uuidBuffer;
                            break;
                        }
                        default:
                            break;
                    }
                    cmdPtr += loadCmd->cmdsize;
                }
                if (newImageInfo->uuid) {
                    newImageInfo->baseAddress = (uintptr_t)header;
                    newImageInfo->slide = (uintptr_t)_dyld_get_image_vmaddr_slide(iImg);
                    NSString *imageName =
                    [NSString stringWithCString:_dyld_get_image_name((unsigned)iImg)
                                       encoding:NSUTF8StringEncoding].lastPathComponent;
                    char *imageNameBuffer = (char *)calloc(imageName.length + 1, sizeof(char));
                    [imageName getCString:imageNameBuffer
                                maxLength:imageName.length + 1
                                 encoding:NSUTF8StringEncoding];
                    newImageInfo->imageName = imageNameBuffer;
                }
                g_imagesInfo.push_back(newImageInfo);
            }
        }
    }
    
    ISBinaryImageInfo * imageContainesAddress(uintptr_t address) {
        if (g_imagesInfo.size() == 0) {
            initBinaryImagesInfo();
        }
        
        for (auto it = g_imagesInfo.begin() ;
             it != g_imagesInfo.end();
             ++it) {
            ISBinaryImageInfo *image = (ISBinaryImageInfo *)*it;
            if (image->containesAddress(address)) {
                return image;
            }
        }
        return NULL;
    }
    
    ISBinaryImageInfo *imageOfName(const char *name) {
        if (g_imagesInfo.size() == 0) {
            initBinaryImagesInfo();
        }
        
        for (auto it = g_imagesInfo.begin() ;
             it != g_imagesInfo.end();
             ++it) {
            ISBinaryImageInfo * image = (ISBinaryImageInfo *)*it;
            if (strcmp(name, image->imageName) == 0) {
                return image;
            }
        }
        return NULL;
    }
}
