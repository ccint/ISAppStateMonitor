//
//  ISMonitorCenter.m
//  Sample
//
//  Created by 舒彪 on 2018/4/22.
//  Copyright © 2018年 舒彪. All rights reserved.
//

#import "ISMonitorCenter.h"
#import <Objective-LevelDB/LevelDB.h>
#import "ISSerialization.h"
#include "ISBinaryImageHelper.h"

@interface ISMonitorCenter() <NSURLSessionDelegate>
@property (nonatomic, strong) LevelDB *logDB;
@property (nonatomic, strong) dispatch_queue_t logQueue;
@property (nonatomic, strong) NSString *deviceUUID;
@property (nonatomic, strong) NSString *appId;
@property (nonatomic, strong) NSString *appVersion;
@property (nonatomic, strong) NSString *binaryImageName;
@property (nonatomic, strong) NSURLSession *sharedURLSession;
@property (nonatomic, strong) NSOperationQueue *sessionQueue;
@end

@implementation ISMonitorCenter

NSData *getStackData(uintptr_t **stack,
                  int32_t threadCount,
                  char **threadsName,
                  int32_t *buffersSize) {
  ISSerialization *stackInfo = [[ISSerialization alloc] init];
  for (int i = 0; i < threadCount; ++i) {
    // get thread stack
    ISSerialization *threadInfo = [[ISSerialization alloc] init];
    uintptr_t *threadStack = stack[i];
    char *threadName = threadsName[i];
    int32_t stackDeep = buffersSize[i];
    NSString *threadNameStr = [NSString stringWithCString:threadName
                                                 encoding:NSUTF8StringEncoding];
    threadNameStr = threadNameStr.length > 0 ? threadNameStr : [NSString stringWithFormat:@"%d", i];
    [threadInfo setString:threadNameStr
                 forKey:@"thread_name"];
    
    // get one stack line
    for (int bufferIndex = 0; bufferIndex < stackDeep; ++bufferIndex) {
      ISSerialization *bufferInfo = [[ISSerialization alloc] init];
      uintptr_t address = threadStack[bufferIndex];
      ISBinaryImageInfoRef matchedImage = imageContainesAddress(address);
      if (matchedImage) {
        NSString *imageName = [NSString stringWithCString:matchedImage->imageName
                                                 encoding:NSUTF8StringEncoding];
        NSString *uuid = [NSString stringWithCString:matchedImage->uuid
                                          encoding:NSUTF8StringEncoding];
        [bufferInfo setString:imageName
                     forKey:@"mod_name"];
        [bufferInfo setData:[NSData dataWithBytes:&address
                                           length:sizeof(uintptr_t)]
                     forKey:@"ret_adr"];
        [bufferInfo setData:[NSData dataWithBytes:&matchedImage->baseAddress
                                           length:sizeof(uintptr_t)]
                     forKey:@"load_adr"];
        [threadInfo appendData:[bufferInfo generateDataFromDictionary]];
        [stackInfo setString:uuid
                    forKey:imageName];
      } else {
        NSLog(@"no match why??");
      }
    }
    [threadInfo setData:[threadInfo generateDataFromArray]
                 forKey:@"th_stack"];
    [stackInfo appendData:[threadInfo generateDataFromDictionary]];
  }
  [stackInfo setData:[stackInfo generateDataFromArray]
              forKey:@"bs"];
  return [stackInfo generateDataFromDictionary];
}

+ (ISMonitorCenter *)sharedInstance {
  static ISMonitorCenter *sharedInstance = nil;
  static dispatch_once_t onceToken;
  dispatch_once(&onceToken, ^{
    sharedInstance = [[ISMonitorCenter alloc] init];
    NSString *libraryPath = NSSearchPathForDirectoriesInDomains(NSLibraryDirectory,
                                                         NSUserDomainMask,
                                                         YES).firstObject;
    sharedInstance.logQueue = dispatch_queue_create("ISMonitorLogQueue", NULL);
    sharedInstance.logDB = [[LevelDB alloc] initWithPath:[libraryPath
                                                          stringByAppendingPathComponent:@"ISMonitorLog"]
                                                      andName:@"ISMonitorLog"];
    NSOperationQueue *sessionQueue = [[NSOperationQueue alloc] init];
    NSURLSessionConfiguration *configuration = [NSURLSessionConfiguration defaultSessionConfiguration];
    NSURLSession *sharedSession = [NSURLSession sessionWithConfiguration:configuration
                                                                delegate:sharedInstance
                                                           delegateQueue:sessionQueue];
    sharedInstance.sharedURLSession = sharedSession;
    sharedInstance.sessionQueue = sessionQueue;
    if (!sharedInstance.logDB) {
      // some thing wrong
    }
  });
  return sharedInstance;
}

+ (void)setLogBaseInfoWithAppVersion:(NSString *)appVersion
                               appId:(NSString *)appId
                     binaryImageName:(NSString *)binaryImageName
                                deviceUUID:(NSString *)deviceUUID {
  ISMonitorCenter *sharedCenter = [self sharedInstance];
  sharedCenter.appVersion = appVersion;
  sharedCenter.appId = appId;
  sharedCenter.binaryImageName = binaryImageName;
  sharedCenter.deviceUUID = deviceUUID;
}

+ (void)logMainTreadTimeoutWithDuration:(double)duration
                                  stack:(uintptr_t **)stack
                            threadCount:(int32_t)threadCount
                            threadsName:(char **)threadsName
                             buffersSize:(int32_t *)buffersSize {
  static ISMonitorCenter *sharedInstance = [self sharedInstance];
  dispatch_async(sharedInstance.logQueue, ^{
    NSData *stackData = getStackData(stack, threadCount, threadsName, buffersSize);
    ISSerialization *serialization = [[ISSerialization alloc] init];
    [serialization setData:stackData forKey:@"bs"];
    [serialization setDouble:duration forKey:@"dur"];
    [serialization setDouble:[[NSDate date] timeIntervalSince1970] * 1000 forKey:@"time"];
    NSData *logData = [serialization generateDataFromDictionary];
    if (logData) {
      char *logIdBuffer = (char *)calloc(7 + 20, sizeof(char));
      sprintf(logIdBuffer,
              "mt_out_%llu",
              (unsigned long long)(CFAbsoluteTimeGetCurrent() * 1000));
      NSString *logId = [[NSString alloc] initWithCString:logIdBuffer
                                                      encoding:NSUTF8StringEncoding];
      [sharedInstance.logDB setObject:logData forKey:logId];
      [self uploadData];
      [sharedInstance.logDB removeObjectForKey:logId];
    }
  });
}

+ (void)uploadData {
  NSMutableArray<NSData *> *mainThreadTimeoutLogs = [[NSMutableArray alloc] init];
  ISMonitorCenter *sharedCenter = [self sharedInstance];
  [sharedCenter.logDB enumerateKeysAndObjectsBackward:YES
                                       lazily:NO
                                startingAtKey:nil
                          filteredByPredicate:nil
                                    andPrefix:@"mt_out_"
                                   usingBlock:^(LevelDBKey *key, id value, BOOL *stop) {
                                     if (value) {
                                       [mainThreadTimeoutLogs addObject:value];
                                     }
                                   }];
  
  if (mainThreadTimeoutLogs.count) {
    ISSerialization *serialization = [[ISSerialization alloc] init];
    [mainThreadTimeoutLogs enumerateObjectsUsingBlock:^(NSData * _Nonnull obj,
                                                        NSUInteger idx,
                                                        BOOL * _Nonnull stop) {
      [serialization appendData:obj];
    }];
  
    NSData *logData = [serialization generateDataFromArray];
    [serialization setData:[@"mt_out" dataUsingEncoding:NSUTF8StringEncoding] forKey:@"type"];
    [serialization setData:logData forKey:@"data"];
    [serialization setString:sharedCenter.appVersion forKey:@"app_ver"];
    [serialization setString:sharedCenter.appId forKey:@"app_id"];
    [serialization setString:sharedCenter.deviceUUID forKey:@"dev_uuid"];
    [serialization setString:[self arch] forKey:@"arch"];
    NSData *finalData = [serialization generateDataFromDictionary];
    if (!finalData) {
      return;
    }
    
    NSURLSession *session = sharedCenter.sharedURLSession;
    NSURL *url = [NSURL URLWithString:@"https://192.168.11.13:4000/report"];
    NSMutableURLRequest *request = [NSMutableURLRequest requestWithURL:url];
    request.HTTPMethod = @"POST";
    request.HTTPBody = finalData;
    NSURLSessionDataTask *dataTask =
    [session dataTaskWithRequest:request
               completionHandler:^(NSData * _Nullable data,
                                   NSURLResponse * _Nullable response,
                                   NSError * _Nullable error) {
                 NSLog(@"ret: %@", [[NSString alloc] initWithData:data encoding:NSUTF8StringEncoding]);
    }];
    [dataTask resume];
  }
}

+ (NSString *)arch {
#if defined(__arm64__)
  return @"arm64";
#elif defined(__arm__)
  return @"armv7";
#elif defined(__x86_64__)
  return @"x86_64";
#elif defined(__i386__)
  return @"i386";
#endif
  return @"";
}

#pragma mark - URLSessionDelegate
- (void)URLSession:(NSURLSession *)session didReceiveChallenge:(NSURLAuthenticationChallenge *)challenge completionHandler:(void (^)(NSURLSessionAuthChallengeDisposition, NSURLCredential *))completionHandler
{
  completionHandler(NSURLSessionAuthChallengeUseCredential , [NSURLCredential credentialForTrust:challenge.protectionSpace.serverTrust]);
}
@end
