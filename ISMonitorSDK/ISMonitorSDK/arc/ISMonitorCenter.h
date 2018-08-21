//
//  ISMonitorCenter.h
//  Sample
//
//  Created by 舒彪 on 2018/4/22.
//  Copyright © 2018年 intsig. All rights reserved.
//

#import <Foundation/Foundation.h>
#include "ISMainThreadChecker.h"

@interface ISMonitorCenter : NSObject

+ (void)setUploadHost:(NSString *)host  isDebug:(BOOL)isDebug;

+ (void)logMainTreadTimeoutWithResult:(ISMainThreadChecker::CheckerResultPtr)checkerResultPtr;

@end
