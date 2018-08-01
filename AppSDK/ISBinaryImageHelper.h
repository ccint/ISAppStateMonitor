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

typedef struct ISBinaryImageInfo {
  uintptr_t baseAddress;
  uintptr_t maxAddress;
  const char *imageName;
  const char *uuid;
} ISBinaryImageInfo, *ISBinaryImageInfoRef;

void initBinaryImagesInfo();

ISBinaryImageInfoRef imageContainesAddress(uintptr_t address);

ISBinaryImageInfoRef imageOfName(const char *name);

#endif /* ISBinaryImageHelper_h */
