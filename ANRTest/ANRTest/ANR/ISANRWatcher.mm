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
+ (void)setLogBaseInfoWithAppVersion:(NSString *)appVersion
                               appId:(NSString *)appId
                     binaryImageName:(NSString *)binaryImageName
                          deviceUUID:(NSString *)deviceUUID {
    [ISMonitorCenter setLogBaseInfoWithAppVersion:appVersion
                                            appId:appId
                                  binaryImageName:binaryImageName
                                       deviceUUID:deviceUUID];
}

+ (void)startWatch:(uint64_t)runloopThreshold {
    ISMainThreadChecker::startWatch(runloopThreshold);
}

+ (void)stopWatch {
    ISMainThreadChecker::stopWatch();
}

@end
