//
//  ISANRWatcher.h
//  ANRTest
//
//  Created by Brent Shu on 2018/8/1.
//  Copyright © 2018年 intsig. All rights reserved.
//

#import <Foundation/Foundation.h>

@interface ISANRWatcher : NSObject
+ (void)setLogBaseInfoWithAppVersion:(NSString *)appVersion
                               appId:(NSString *)appId
                     binaryImageName:(NSString *)binaryImageName
                          deviceUUID:(NSString *)deviceUUID;
+ (void)startWatch:(uint64_t)runloopThreshold;
+ (void)stopWatch;
+ (int)logCount;
@end
