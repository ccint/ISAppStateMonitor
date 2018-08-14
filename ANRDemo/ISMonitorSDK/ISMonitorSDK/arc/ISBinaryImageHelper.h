//
//  ISBinaryImageHelper.h
//  Sample
//
//  Created by 舒彪 on 2018/4/30.
//  Copyright © 2018年 intsig. All rights reserved.
//

#ifndef ISBinaryImageHelper_h
#define ISBinaryImageHelper_h

#include <stdio.h>
#include <vector>

namespace ISBinaryImage {
    struct ISBinaryImageInfo {
        uintptr_t baseAddress;
        std::vector<std::pair<uint64_t, uint64_t>> segmentRanges;
        uintptr_t slide;
        const char *imageName;
        const char *uuid;
        
        ISBinaryImageInfo();
        ~ISBinaryImageInfo();
        
        bool containesAddress(uintptr_t address);
    };
    
    ISBinaryImageInfo *imageContainesAddress(uintptr_t address);
    
    ISBinaryImageInfo *imageOfName(const char *name);
    
    void initBinaryImagesInfo();
}
#endif /* ISBinaryImageHelper_h */
