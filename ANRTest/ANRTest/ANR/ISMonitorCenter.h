//
//  ISMonitorCenter.h
//  Sample
//
//  Created by 舒彪 on 2018/4/22.
//  Copyright © 2018年 舒彪. All rights reserved.
//

#import <Foundation/Foundation.h>

@interface ISMonitorCenter : NSObject

+ (void)setLogBaseInfoWithAppVersion:(NSString *)appVersion
                               appId:(NSString *)appId
                     binaryImageName:(NSString *)binaryImageName
                                deviceUUID:(NSString *)deviceUUID;


+ (void)logMainTreadTimeoutWithDuration:(double)duration
                                  stack:(uintptr_t **)stack
                            threadCount:(int32_t)threadCount
                            threadsName:(char **)threadsName
                             buffersSize:(int32_t *)buffersSize;
@end
