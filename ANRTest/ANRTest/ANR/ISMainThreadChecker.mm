//
//  ISMainThreadChecker.mm
//  MainThreadWatcher
//
//  Created by 舒彪 on 2018/1/15.
//  Copyright © 2018年 intsig. All rights reserved.
//

#import "ISMainThreadChecker.h"
#import <QuartzCore/QuartzCore.h>
#import <mach/mach.h>
#import <pthread.h>
#import "ISMonitorCenter.h"

#define SIGBSR SIGUSR1

namespace ISMainThreadChecker {
    
    static const uint64_t MSEC_PER_SEC = 1000000ULL;
    static const uint64_t NANOS_PER_MSEC = 1000000ULL;
    static Checker *checkerInstance;
    
#pragma mark - Tools
    
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
    
    inline static uint64_t nanos_to_abs(uint64_t nanos, mach_timebase_info_data_t timebase_info) {
        return nanos * timebase_info.denom / timebase_info.numer;
    }
    
    inline static void wait_until(uint64_t msec, mach_timebase_info_data_t timebase_info) {
        uint64_t time_to_wait = nanos_to_abs(msec * NANOS_PER_MSEC, timebase_info);
        uint64_t now = mach_absolute_time();
        mach_wait_until(now + time_to_wait);
    }
    
#pragma mark - signalHandler
    
    static void alarmSignalHandler(int sig) {
        // do nothing
    }
    
    static void backTraceRecordSignalHandler(int sig) {
        if (sig == SIGBSR) {
            for (NSString *string in [NSThread callStackSymbols]) {
                printf("%s\n", [string cStringUsingEncoding:NSUTF8StringEncoding]);
            }
        }
    }
    
#pragma mark - Thread Loop
    
    static void *checkerThreadLoop(void* argument) {
        signal(SIGALRM, alarmSignalHandler);
        auto checker = (Checker *)argument;
        while (1) {
            dispatch_semaphore_wait(checker->loopSem, DISPATCH_TIME_FOREVER);
            auto result = checker->result;
            if (result) {
                double semDuration = (CACurrentMediaTime() - result->runloopId) * MSEC_PER_SEC;
                checker->isScheduing = true;
                wait_until(checker->waitTime - semDuration, checker->timebase_info);
                checker->isScheduing = false;
                if (checker->isResetBySignal) {
                    checker->isResetBySignal = false;
                    continue;
                }
                auto resultAtMoment = checker->result;
                if (resultAtMoment &&
                    result->runloopId == resultAtMoment->runloopId &&
                    resultAtMoment->stacks.size() == 0) {
                    suspendOtherThreads();
                    ISBSRecorder::backtraceOfAllThread(resultAtMoment->stacks);
                    resumeOtherThreads();
                }
            }
        }
    }
    
#pragma mark - Runloop callBack
    
    static void runloopObserverCallBack(CFRunLoopObserverRef observer, CFRunLoopActivity activity, void *info) {
        if (activity == kCFRunLoopAfterWaiting) {
            checkerInstance->beginSchedule();
        } else if (activity == kCFRunLoopBeforeWaiting) {
            auto result = checkerInstance->result;
            if (result && result->stacks.size() > 0) {
                result->runloopDuration = (CACurrentMediaTime() - result->runloopId) * 1000;
                [ISMonitorCenter logMainTreadTimeoutWithResult:result];
            }
            checkerInstance->finishSchedule();
        }
    }
    
#pragma mark - Checker Impl
    
    Checker::~Checker(){
        this->mainThread = nullptr;
        dispatch_release(this->loopSem);
        if (this->thread) {
            pthread_cancel(this->thread);
            pthread_join(this->thread, NULL);
        }
    }
    
    Checker::Checker():thread(nullptr)
    ,mainThread(nullptr)
    ,loopSem(nullptr)
    ,currentRunloopId(0)
    ,isScheduing(false)
    ,isResetBySignal(false)
    ,result(nullptr)
    ,wakeObserver(nullptr)
    ,sleepObserver(nullptr)
    ,isWatching(false)
    ,waitTime(0)
    {
        mach_timebase_info(&this->timebase_info);
        this->loopSem = dispatch_semaphore_create(0);
    }
    
    void Checker::beginSchedule() {
        this->result = CheckerResultPtr(new CheckerResult(CACurrentMediaTime()));
        if (this->isScheduing) {
            this->isResetBySignal = YES;
            pthread_kill(this->thread, SIGALRM);
        }
        dispatch_semaphore_signal(this->loopSem);
    }
    
    void Checker::finishSchedule() {
        this->result = nullptr;
    }
    
    bool Checker::startWatch(uint64_t runloopThreshold) {
        assert([NSThread isMainThread]);
        static dispatch_once_t onceToken;
        __block BOOL result = YES;
        dispatch_once(&onceToken, ^{
            this->mainThread = pthread_self();
            if (pthread_create(&this->thread, NULL, checkerThreadLoop, this) == 0) {
                move_pthread_to_realtime_scheduling_class(this->thread);
                signal(SIGBSR, backTraceRecordSignalHandler);
            } else {
                result = NO;
            }
        });
        
        if (!result) {
            return result;
        }
        
        this->waitTime = runloopThreshold;
        this->wakeObserver = CFRunLoopObserverCreate(NULL,
                                                     kCFRunLoopAfterWaiting,
                                                     true,
                                                     0,
                                                     runloopObserverCallBack,
                                                     NULL);
        this->sleepObserver = CFRunLoopObserverCreate(NULL,
                                                      kCFRunLoopBeforeWaiting,
                                                      true,
                                                      INT_MAX,
                                                      runloopObserverCallBack,
                                                      NULL);
        CFRunLoopAddObserver(CFRunLoopGetMain(), this->wakeObserver, kCFRunLoopCommonModes);
        CFRunLoopAddObserver(CFRunLoopGetMain(), this->sleepObserver, kCFRunLoopCommonModes);
        this->isWatching = true;
        
        return result;
    }
    
    void Checker::stopWatch() {
        CFRunLoopRemoveObserver(CFRunLoopGetMain(), wakeObserver, kCFRunLoopCommonModes);
        CFRunLoopRemoveObserver(CFRunLoopGetMain(), sleepObserver, kCFRunLoopCommonModes);
        this->wakeObserver = nullptr;
        this->sleepObserver = nullptr;
        this->isWatching = false;
    }
    
    void startWatch(uint64_t runloopThreshold) {
        if (!checkerInstance) {
            checkerInstance = new Checker();
        }
        if (!checkerInstance->isWatching) {
            checkerInstance->startWatch(runloopThreshold);
        }
    }
    
    void stopWatch() {
        if (checkerInstance && checkerInstance->isWatching) {
            checkerInstance->stopWatch();
        }
    }
}
