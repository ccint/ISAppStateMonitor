//
//  ISSerialization.h
//  LevelDBTest
//
//  Created by 舒彪 on 2018/1/27.
//  Copyright © 2018年 intsig. All rights reserved.
//

#import <Foundation/Foundation.h>

// 高效的序列化工具，暂不支持递归

@interface ISSerialization : NSObject

#pragma mark - ForDictionary
- (void)setData:(NSData *)data forKey:(NSString *)key;

- (void)setDouble:(double)doubleValue forKey:(NSString *)key;

- (void)setInteger:(NSInteger)integer forKey:(NSString *)key;

- (void)setString:(NSString *)string forKey:(NSString *)key;

- (NSData *)generateDataFromDictionary;

- (instancetype)initWithSerializedDictionaryData:(NSData *)data;

- (NSData *)dataForKey:(NSString *)key;

- (double)doubleForKey:(NSString *)key;

- (NSInteger)integerForKey:(NSString *)key;

- (NSString *)stringForKey:(NSString *)key;

#pragma mark - ForArray
- (void)appendData:(NSData *)data;

- (void)appendDouble:(double)doubleValue;

- (void)appendInteger:(NSInteger)integer;

- (void)appendString:(NSString *)string;

- (NSData *)generateDataFromArray;

- (instancetype)initWithSerializedArrayData:(NSData *)data;

- (NSData *)dataAtIndex:(NSInteger)index;

- (double)doubleAtIndex:(NSInteger)index;

- (NSInteger)integerAtIndex:(NSInteger)index;

- (NSString *)stringAtIndex:(NSInteger)index;
@end
