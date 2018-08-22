//
//  ISMonitorCenter.m
//  Sample
//
//  Created by 舒彪 on 2018/4/22.
//  Copyright © 2018年 intsig. All rights reserved.
//

#import "ISMonitorCenter.h"
#import "ISLevelDB.h"
#import "ISSerialization.h"
#include "ISBinaryImageHelper.h"

@interface ISMonitorCenter() <NSURLSessionDelegate>
@property (nonatomic, strong) ISLevelDB *logDB;
@property (nonatomic, strong) dispatch_queue_t logQueue;
@property (nonatomic, strong) NSString *systemVersion;
@property (nonatomic, strong) NSString *appId;
@property (nonatomic, strong) NSString *appVersion;
@property (nonatomic, strong) NSString *binaryImageName;
@property (nonatomic, strong) NSURLSession *sharedURLSession;
@property (nonatomic, strong) NSOperationQueue *sessionQueue;
@property (nonatomic, strong) NSString *serverHost;
@property (nonatomic, assign) BOOL isDebug;
@end

@implementation ISMonitorCenter

NSData *getStackData(ISBSRecorder::Stacks & stacks) {
    ISSerialization *stackInfo = [[ISSerialization alloc] init];
    ISSerialization *imagesInfo = [[ISSerialization alloc] init];
    for (ISBSRecorder::Stacks::iterator it = stacks.begin(); it != stacks.end(); ++it) {
        // get thread stack
        ISSerialization *threadInfo = [[ISSerialization alloc] init];
        auto stack = *it;
        
        // get frames
        for (ISBSRecorder::Frames::iterator it = stack.frames.begin() ;
             it != stack.frames.end();
             ++it) {
            ISSerialization *bufferInfo = [[ISSerialization alloc] init];
            uintptr_t address = *it;
            
            ISBinaryImage::ISBinaryImageInfo *matchedImage = ISBinaryImage::imageContainesAddress(address);
            if (matchedImage) {
                NSString *imageName = [NSString stringWithCString:matchedImage->imageName
                                                         encoding:NSUTF8StringEncoding];
                NSString *uuid = [NSString stringWithCString:matchedImage->uuid
                                                    encoding:NSUTF8StringEncoding];
                if (uuid.length) {
                    uuid = [uuid stringByReplacingOccurrencesOfString:@"-" withString:@""];
                }
                [bufferInfo setString:uuid
                               forKey:@"image_uuid"];
                [bufferInfo setData:[NSData dataWithBytes:&address
                                                   length:sizeof(uintptr_t)]
                             forKey:@"ret_adr"];
                [bufferInfo setData:[NSData dataWithBytes:&matchedImage->baseAddress
                                                   length:sizeof(uintptr_t)]
                             forKey:@"load_adr"];
                [threadInfo appendData:[bufferInfo generateDataFromDictionary]];
                [imagesInfo setString:imageName
                               forKey:uuid];
            } else {
                // no match, just ignore
            }
        }
        
        NSData *framesData = [threadInfo generateDataFromArray];
        if (framesData) {
            const char *threadName = stack.threadName.c_str();
            NSString *threadNameStr = [NSString stringWithCString:threadName
                                                         encoding:NSUTF8StringEncoding];
            threadNameStr = threadNameStr.length > 0 ? threadNameStr : @"Thread";
            [threadInfo setString:threadNameStr
                           forKey:@"thread_name"];
            [threadInfo setData:framesData
                         forKey:@"th_stack"];
            [stackInfo appendData:[threadInfo generateDataFromDictionary]];
        } else {
            NSLog(@"invalid frames, ignore");
        }
    }
    NSData *stacksData = [stackInfo generateDataFromArray];
    if (stacksData) {
        [stackInfo setData:stacksData
                    forKey:@"bs"];
        [stackInfo setData:[imagesInfo generateDataFromDictionary]
                    forKey:@"images"];
        [stackInfo setString:[ISMonitorCenter sharedInstance].binaryImageName forKey:@"appImageName"];
        return [stackInfo generateDataFromDictionary];
    } else {
        return nil;
    }
}

+ (ISMonitorCenter *)sharedInstance {
    static ISMonitorCenter *sharedInstance = nil;
    static dispatch_once_t onceToken;
    dispatch_once(&onceToken, ^{
        sharedInstance = [[ISMonitorCenter alloc] init];
        NSDictionary *info = [[NSBundle mainBundle] infoDictionary];
        NSString *appVersion = [info objectForKey:@"CFBundleShortVersionString"];
        NSString *appid = [[NSBundle mainBundle] bundleIdentifier];
        NSString *exeName = [info objectForKey:@"CFBundleExecutable"];
        NSString *systemVersion = [[UIDevice currentDevice] systemVersion];

        NSString *assertMsg = [NSString stringWithFormat:@"Get Baseinfo failed Av:%@, Ap:%@, exe:%@, sv:%@",
                               appVersion,
                               appid,
                               exeName,
                               systemVersion];
        NSAssert(appVersion.length > 0 &&
               appid.length > 0 &&
               exeName.length > 0 &&
                 systemVersion.length > 0, assertMsg);
        
        sharedInstance.appVersion = appVersion;
        sharedInstance.appId = appid;
        sharedInstance.binaryImageName = exeName;
        sharedInstance.systemVersion = systemVersion;
        
        NSString *libraryPath = NSSearchPathForDirectoriesInDomains(NSLibraryDirectory,
                                                                    NSUserDomainMask,
                                                                    YES).firstObject;
        sharedInstance.logQueue = dispatch_queue_create("ISMonitorLogQueue", NULL);
        sharedInstance.logDB = [[ISLevelDB alloc] initWithPath:[libraryPath
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

+ (void)setUploadHost:(NSString *)host isDebug:(BOOL)isDebug {
    auto sharedCenter = [self sharedInstance];
    sharedCenter.isDebug = isDebug;
    sharedCenter.serverHost = host;
    
    if (isDebug) {
        sharedCenter.appId = [sharedCenter.appId stringByAppendingString:@".DEBUG"];
    }
}

+ (void)logMainTreadTimeoutWithResult:(ISMainThreadChecker::CheckerResultPtr)checkerResultPtr {
    static ISMonitorCenter *sharedInstance = [self sharedInstance];
    dispatch_async(sharedInstance.logQueue, ^{
        NSData *stackData = getStackData(checkerResultPtr->stacks);
        if (!stackData) {
            NSLog(@"invalid stack, just return");
            return ;
        }
        ISSerialization *serialization = [[ISSerialization alloc] init];
        [serialization setData:stackData forKey:@"bs"];
        [serialization setDouble:checkerResultPtr->runloopDuration forKey:@"dur"];
        [serialization setDouble:[[NSDate date] timeIntervalSince1970] * 1000 forKey:@"time"];
        NSData *logData = [serialization generateDataFromDictionary];
        if (logData) {
            [self uploadData:@[logData]];
//            char logIdBuffer[30] = {0};
//            sprintf(logIdBuffer,
//                    "mt_out_%llu",
//                    (unsigned long long)(CFAbsoluteTimeGetCurrent() * 1000));
//            NSString *logId = [[NSString alloc] initWithCString:logIdBuffer
//                                                       encoding:NSUTF8StringEncoding];
//            [sharedInstance.logDB setObject:logData forKey:logId];
        }
    });
}

+ (void)uploadDataFromDatabase {
    NSMutableArray<NSData *> *mainThreadTimeoutLogs = [[NSMutableArray alloc] init];
    ISMonitorCenter *sharedCenter = [self sharedInstance];
    [sharedCenter.logDB enumerateKeysAndObjectsBackward:YES
                                                 lazily:NO
                                          startingAtKey:nil
                                    filteredByPredicate:nil
                                              andPrefix:@"mt_out_"
                                             usingBlock:^(ISLevelDBKey *key, id value, BOOL *stop) {
                                                 if (value) {
                                                     [mainThreadTimeoutLogs addObject:value];
                                                 }
                                             }];
    
    if (mainThreadTimeoutLogs.count) {
        [self uploadData:mainThreadTimeoutLogs];
        [sharedCenter.logDB removeAllObjects];
    }
}

+ (void)uploadData:(NSArray<NSData *> *)logDatas {
    if (logDatas.count <= 0) {
        return;
    }
    
    auto sharedCenter = [self sharedInstance];
    
    ISSerialization *serialization = [[ISSerialization alloc] init];
    [logDatas enumerateObjectsUsingBlock:^(NSData * _Nonnull obj,
                                                        NSUInteger idx,
                                                        BOOL * _Nonnull stop) {
        [serialization appendData:obj];
    }];
    
    NSData *logData = [serialization generateDataFromArray];
    [serialization setData:[@"mt_out" dataUsingEncoding:NSUTF8StringEncoding] forKey:@"type"];
    [serialization setData:logData forKey:@"data"];
    [serialization setString:sharedCenter.appVersion forKey:@"app_ver"];
    [serialization setString:sharedCenter.appId forKey:@"app_id"];
    [serialization setString:sharedCenter.systemVersion forKey:@"sys_ver"];
    [serialization setString:[self arch] forKey:@"arch"];
    NSData *finalData = [serialization generateDataFromDictionary];
    if (!finalData) {
        return;
    }
    
    NSURLSession *session = sharedCenter.sharedURLSession;
    NSURL *url = [NSURL URLWithString:[NSString stringWithFormat:@"%@/report", sharedCenter.serverHost]];
    NSMutableURLRequest *request = [NSMutableURLRequest requestWithURL:url];
    request.HTTPMethod = @"POST";
    request.HTTPBody = finalData;
    NSURLSessionDataTask *dataTask =
    [session dataTaskWithRequest:request
               completionHandler:^(NSData * _Nullable data,
                                   NSURLResponse * _Nullable response,
                                   NSError * _Nullable error) {
                   
               }];
    [dataTask resume];
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
- (void)URLSession:(NSURLSession *)session
    didReceiveChallenge:(NSURLAuthenticationChallenge *)challenge
      completionHandler:(void (^)(NSURLSessionAuthChallengeDisposition, NSURLCredential *))completionHandler
{
    completionHandler(NSURLSessionAuthChallengeUseCredential ,
                      [NSURLCredential credentialForTrust:challenge.protectionSpace.serverTrust]);
}
@end
