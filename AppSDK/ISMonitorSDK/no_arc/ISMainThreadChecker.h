//
//  ISMainThreadChecker.h
//  MainThreadWatcher
//
//  Created by 舒彪 on 2018/1/15.
//  Copyright © 2018年 intsig. All rights reserved.
//

#import <Foundation/Foundation.h>
#import <mach/mach_time.h>
#include "ISStacktraceRecorder.hpp"
#include <memory>

namespace ISMainThreadChecker {
    int logCount();
    
    struct CheckerResult {
        double runloopId;
        double runloopDuration;
        ISBSRecorder::Stacks stacks;
        
        CheckerResult(double loopId)
        :runloopId(loopId)
        ,runloopDuration(0)
        ,stacks()
        {
            
        }
    };
    
    typedef std::shared_ptr<CheckerResult> CheckerResultPtr;
    
    struct Checker {
        mach_timebase_info_data_t timebase_info;
        pthread_t thread;
        pthread_t mainThread;
        dispatch_semaphore_t loopSem;
        double currentRunloopId;
        uint64_t waitTime;
        bool isScheduing;
        CFRunLoopObserverRef wakeObserver;
        CFRunLoopObserverRef sleepObserver;
        bool isWatching;
        
        Checker();
        ~Checker();
        
        bool startWatch(uint64_t runloopThreshold);
        void stopWatch();
        
        void beginSchedule();
        void finishSchedule();
        
        CheckerResultPtr getResult();
        
    private:
        void setResult(CheckerResultPtr newPtr);
        CheckerResultPtr result;
    };
    
    void startWatch(uint64_t runloopThreshold);
    void stopWatch();
}
