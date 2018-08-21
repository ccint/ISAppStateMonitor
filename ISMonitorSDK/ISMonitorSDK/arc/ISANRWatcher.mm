//
//  ISANRWatcher.mm
//  ANRTest
//
//  Created by Brent Shu on 2018/8/1.
//  Copyright © 2018年 intsig. All rights reserved.
//

#import "ISANRWatcher.h"
#import "ISMainThreadChecker.h"
#import "ISMonitorCenter.h"

@implementation ISANRWatcher
+ (void)setUploadHost:(NSString *)host isDebug:(BOOL)isDebug {
    [ISMonitorCenter setUploadHost:host isDebug:isDebug];
}

+ (void)startWatch:(uint64_t)runloopThreshold {
    ISMainThreadChecker::startWatch(runloopThreshold);
}

+ (void)stopWatch {
    ISMainThreadChecker::stopWatch();
}

+ (int)logCount {
    return ISMainThreadChecker::logCount();
}

@end
