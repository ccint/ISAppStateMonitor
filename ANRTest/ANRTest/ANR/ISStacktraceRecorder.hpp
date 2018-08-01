//
//  ISStacktraceRecorder.hpp
//  ANRTest
//
//  Created by Brent Shu on 2018/8/1.
//  Copyright © 2018年 intsig. All rights reserved.
//

#ifndef ISStacktraceRecorder_hpp
#define ISStacktraceRecorder_hpp

#include <stdio.h>
#include <vector>
#include <string>

namespace ISBSRecorder {
    typedef std::vector<uintptr_t> Frames;
    
    struct Stack {
        std::string threadName;
        Frames frames;
    };
    
    enum status {
        status_ok       = 0,
        status_error    = 1
    };
    
    typedef std::vector<Stack> Stacks;
    
    status backtraceOfCurrentThread(Stack & stack);
    
    void backtraceOfAllThread(Stacks & Stacks);
}


#endif /* ISStacktraceRecorder_hpp */
