//
//  ISCallStackLogger.h
//  CamCard
//
//  Created by Brent Shu on 2018/3/16.
//  Copyright © 2018年 舒彪. All rights reserved.
//

#ifndef ISCallStackLogger_h
#define ISCallStackLogger_h

#include <stdio.h>
#include <mach/mach.h>

bool bs_backtraceOfCurrentThread(uintptr_t *buffer,
                                 int32_t bufferMaxSize,
                                 int32_t *bufferSize);

bool bs_backtraceOfThread(thread_t thread,
                          uintptr_t *buffer,
                          int32_t bufferMaxSize,
                          int32_t *bufferSize);

bool bs_backtraceOfAllThread(uintptr_t **allThreadBuffer,
                             int32_t bufferMaxSize,
                             int32_t maxThreadCount,
                             int32_t *buffersSize,
                             int32_t *threadCount,
                             char **threadsName);

#endif /* ISCallStackLogger_h */
