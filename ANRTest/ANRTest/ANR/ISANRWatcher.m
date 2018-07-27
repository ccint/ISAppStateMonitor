//
//  ISANRWatcher.m
//  MainThreadWatcher
//
//  Created by 舒彪 on 2018/1/15.
//  Copyright © 2018年 舒彪. All rights reserved.
//

#import "ISANRWatcher.h"
#import <QuartzCore/QuartzCore.h>
#import <mach/mach.h>
#import <mach/mach_time.h>
#import <pthread.h>
#import <libkern/OSAtomic.h>

#import "ISMonitorCenter.h"
#import "ISCallStackLogger.h"

#define SIGBSR SIGUSR1
#define WaitTime 1000000

typedef NS_ENUM(NSUInteger, ISMainThreadWatcherMode) {
    ISMainThreadWatcherModeProfile,
    ISMainThreadWatcherModeProd,
};

static const uint64_t NANOS_PER_USEC = 1000ULL;
static ISMainThreadWatcherMode watcherMode = ISMainThreadWatcherModeProd;

typedef struct ISMainThreadCheckerResult {
    double runloopId;
    double runloopDuration;
    uintptr_t **allThreadBuffer;
    int32_t *buffersSize;
    int32_t threadCount;
    char  **threadsName;
    int32_t bufferMaxSize;
    int32_t maxThreadCount;
} ISMainThreadCheckerResult, *ISMainThreadCheckerResultRef;

typedef struct ISMainThreadChecker {
    mach_timebase_info_data_t timebase_info;
    pthread_t thread;
    pthread_t mainThread;
    dispatch_semaphore_t loopSem;
    double currentRunloopId;
    unsigned long long waitTime;
    BOOL isScheduing;
    BOOL isResetBySignal;
    ISMainThreadCheckerResultRef result;
} ISMainThreadChecker, *ISMainThreadCheckerRef;

static ISMainThreadCheckerRef checkerInstance;

static CFRunLoopObserverRef wakeObserver;
static CFRunLoopObserverRef sleepObserver;

@implementation ISANRWatcher

#pragma mark - Interface
+ (void)startWatch {
    if (!checkerInstance) {
        checkerInstance = ISMainThreadCheckerCreate();
        ISMainThreadCheckerStartWatch(checkerInstance);
        wakeObserver = CFRunLoopObserverCreate(NULL,
                                               kCFRunLoopAfterWaiting,
                                               true,
                                               0,
                                               runloopObserverCallBack,
                                               NULL);
        sleepObserver = CFRunLoopObserverCreate(NULL,
                                                kCFRunLoopBeforeWaiting,
                                                true,
                                                INT_MAX,
                                                runloopObserverCallBack,
                                                NULL);
        CFRunLoopAddObserver(CFRunLoopGetMain(), wakeObserver, kCFRunLoopCommonModes);
        CFRunLoopAddObserver(CFRunLoopGetMain(), sleepObserver, kCFRunLoopCommonModes);
    }
}

+ (void)stopWatch {
    if (checkerInstance) {
        CFRunLoopRemoveObserver(CFRunLoopGetMain(), wakeObserver, kCFRunLoopCommonModes);
        CFRunLoopRemoveObserver(CFRunLoopGetMain(), sleepObserver, kCFRunLoopCommonModes);
        ISMainThreadCheckerFree(checkerInstance);
        checkerInstance = NULL;
        wakeObserver = NULL;
        sleepObserver = NULL;
    }
}

#pragma mark - MainThread
static void runloopObserverCallBack(CFRunLoopObserverRef observer, CFRunLoopActivity activity, void *info) {
    if (activity == kCFRunLoopAfterWaiting) {
        ISMainThreadCheckerResultRef newResult = ISMainThreadCheckerResultCreate(CACurrentMediaTime());
        ISMainThreadCheckerBeginSchedule(checkerInstance, newResult);
    } else if (activity == kCFRunLoopBeforeWaiting) {
        ISMainThreadCheckerResultRef result = checkerInstance->result;
        if (result && result->allThreadBuffer) {
            result->runloopDuration = (CACurrentMediaTime() - result->runloopId) * 1000;
            [ISMonitorCenter logMainTreadTimeoutWithDuration:result->runloopDuration
                                                       stack:result->allThreadBuffer
                                                 threadCount:result->threadCount
                                                 threadsName:result->threadsName
                                                 buffersSize:result->buffersSize];
            result->allThreadBuffer = nil;
            result->threadsName = nil;
            result->buffersSize = nil;
        }
        ISMainThreadCheckerFinishSchedule(checkerInstance);
        ISMainThreadCheckerResultFree(result);
    }
}

#pragma mark - WatcherThread
void alarmSignalHandler(int sig) {
    // do nothing
}

void *checkerThreadLoop(void* argument) {
    signal(SIGALRM, alarmSignalHandler);
    ISMainThreadCheckerRef checker = (ISMainThreadCheckerRef)argument;
    while (YES) {
        dispatch_semaphore_wait(checker->loopSem, DISPATCH_TIME_FOREVER);
        ISMainThreadCheckerResultRef result = checker->result;
        if (result) {
            double semDuration = (CACurrentMediaTime() - result->runloopId) * 1000000;
            checker->isScheduing = YES;
            wait_until(checker->waitTime - semDuration, checker->timebase_info);
            checker->isScheduing = NO;
            if (checker->isResetBySignal) {
                checker->isResetBySignal = NO;
                continue;
            }
            ISMainThreadCheckerResultRef resultAtMoment = checker->result;
            if (resultAtMoment && result->runloopId == resultAtMoment->runloopId) {
                if (!resultAtMoment->allThreadBuffer) {
                    ISMainThreadCheckerResultCreateThreadBuffer(resultAtMoment);
                }
                if (watcherMode == ISMainThreadWatcherModeProd) {
                    ismc_suspendEnvironment();
                    bs_backtraceOfAllThread(resultAtMoment->allThreadBuffer,
                                            resultAtMoment->bufferMaxSize,
                                            resultAtMoment->maxThreadCount,
                                            resultAtMoment->buffersSize,
                                            &resultAtMoment->threadCount,
                                            resultAtMoment->threadsName);
                    ismc_resumeEnvironment();
                } else {
                    pthread_kill(checker->mainThread, SIGBSR);
                }
            }
        }
    }
}

inline static void ISAtomicSetPointer(void *newValue, void **target) {
    while (YES) {
        void *ptr = *target;
        if (OSAtomicCompareAndSwapPtrBarrier(ptr, newValue, target)) {
            break;
        }
    }
}

inline static void ISMainThreadCheckerBeginSchedule(ISMainThreadCheckerRef checker,
                                                    ISMainThreadCheckerResultRef result) {
    ISAtomicSetPointer(result, (void **)&checker->result);
    if (checker->isScheduing) {
        checker->isResetBySignal = YES;
        pthread_kill(checker->thread, SIGALRM);
    }
    dispatch_semaphore_signal(checker->loopSem);
}

inline static void ISMainThreadCheckerFinishSchedule(ISMainThreadCheckerRef checker) {
    ISAtomicSetPointer(NULL, (void **)&checker->result);
}

inline static ISMainThreadCheckerRef ISMainThreadCheckerCreate() {
    assert([NSThread isMainThread]);
    ISMainThreadCheckerRef newChecker = (ISMainThreadCheckerRef)calloc(1, sizeof(ISMainThreadChecker));
    mach_timebase_info(&newChecker->timebase_info);
    newChecker->waitTime = WaitTime;
    return newChecker;
}

static BOOL ISMainThreadCheckerStartWatch(ISMainThreadCheckerRef checker) {
    assert([NSThread isMainThread]);
    checker->mainThread = pthread_self();
    checker->loopSem = dispatch_semaphore_create(0);
    if (pthread_create(&checker->thread, NULL, checkerThreadLoop, checker) == 0) {
        move_pthread_to_realtime_scheduling_class(checker->thread);
        signal(SIGBSR, backTraceRecordSignalHandler);
        return YES;
    }
    return NO;
}

static void backTraceRecordSignalHandler(int sig) {
    if (sig == SIGBSR) {
        ISMainThreadCheckerResultRef result = checkerInstance->result;
        if (result && result->allThreadBuffer) {
            bs_backtraceOfCurrentThread(result->allThreadBuffer[0], result->bufferMaxSize, result->buffersSize);
        }
    }
}

inline static ISMainThreadCheckerResultRef ISMainThreadCheckerResultCreate(double identifier) {
    ISMainThreadCheckerResultRef newResult = (ISMainThreadCheckerResultRef)calloc(1, sizeof(ISMainThreadCheckerResult));
    if (newResult) {
        newResult->runloopId = identifier;
        return newResult;
    }
    return NULL;
}

static BOOL ISMainThreadCheckerResultCreateThreadBuffer(ISMainThreadCheckerResultRef result) {
    uintptr_t **buffer = (uintptr_t **)calloc(50, sizeof(uintptr_t *));
    char **threadsName = (char **)calloc(50, sizeof(char *));
    for (int i = 0; i < 50; ++i) {
        buffer[i] = (uintptr_t *)calloc(80, sizeof(uintptr_t));
        threadsName[i] = (char *)calloc(50, sizeof(char));
    }
    
    int32_t *buffersSize = (int32_t *)calloc(50, sizeof(int32_t));
    if (buffer) {
        ISAtomicSetPointer(buffer, (void **)&result->allThreadBuffer);
        result->maxThreadCount = 50;
        result->bufferMaxSize = 80;
    }
    
    if (threadsName) {
        ISAtomicSetPointer(threadsName, (void **)&result->threadsName);
    }
    
    if (buffersSize) {
        ISAtomicSetPointer(buffersSize, (void **)&result->buffersSize);
    }
    
    if (buffer && threadsName && buffersSize) {
        return YES;
    }
    
    return NO;
}

static void ISMainThreadCheckerResultFree(ISMainThreadCheckerResultRef result) {
    if (!result) {
        return;
    }
    
    if (result->allThreadBuffer) {
        for (int i = 0; i < result->maxThreadCount; ++i) {
            free(result->allThreadBuffer[i]);
            free(result->threadsName[i]);
        }
        free(result->allThreadBuffer);
        free(result->threadsName);
        
    }
    
    if (result->threadsName) {
        for (int i = 0; i < result->maxThreadCount; ++i) {
            free(result->threadsName[i]);
        }
        free(result->threadsName);
    }
    
    if (result->buffersSize) {
        free(result->buffersSize);
    }
    
    free(result);
}

inline static void ISMainThreadCheckerFree(ISMainThreadCheckerRef checker) {
    if (!checker) {
        return;
    }
    
    ISMainThreadCheckerResultFree(checker->result);
    
    dispatch_release(checker->loopSem);
    pthread_cancel(checker->thread);
    pthread_join(checker->thread, NULL);
    
    free(checker);
}

#pragma mark - Timer-Tools
inline static uint64_t nanos_to_abs(uint64_t nanos, mach_timebase_info_data_t timebase_info) {
    return nanos * timebase_info.denom / timebase_info.numer;
}

inline static void wait_until(unsigned long long usec, mach_timebase_info_data_t timebase_info) {
    uint64_t time_to_wait = nanos_to_abs(usec * NANOS_PER_USEC, timebase_info);
    uint64_t now = mach_absolute_time();
    mach_wait_until(now + time_to_wait);
}

#pragma mark Thread-Tools
static void move_pthread_to_realtime_scheduling_class(pthread_t pthread) {
    mach_timebase_info_data_t timebase_info;
    mach_timebase_info(&timebase_info);
    
    const uint64_t NANOS_PER_MSEC = 1000000ULL;
    double clock2abs = ((double)timebase_info.denom / (double)timebase_info.numer) * NANOS_PER_MSEC;
    
    thread_time_constraint_policy_data_t policy;
    policy.period      = 0;
    policy.computation = (uint32_t)(5 * clock2abs);
    policy.constraint  = (uint32_t)(10 * clock2abs);
    policy.preemptible = FALSE;
    
    int kr = thread_policy_set(pthread_mach_thread_np(pthread),
                               THREAD_TIME_CONSTRAINT_POLICY,
                               (thread_policy_t)&policy,
                               THREAD_TIME_CONSTRAINT_POLICY_COUNT);
    if (kr != KERN_SUCCESS) {
        NSLog(@"setPolicy Failed");
    }
}

inline static thread_t is_thread_self() {
    thread_t thread_self = mach_thread_self();
    mach_port_deallocate(mach_task_self(), thread_self);
    return thread_self;
}

static void ismc_suspendEnvironment() {
    kern_return_t kr;
    const task_t thisTask = mach_task_self();
    const thread_t thisThread = (thread_t)is_thread_self();
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

static void ismc_resumeEnvironment() {
    kern_return_t kr;
    const task_t thisTask = mach_task_self();
    const thread_t thisThread = (thread_t)is_thread_self();
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

@end
