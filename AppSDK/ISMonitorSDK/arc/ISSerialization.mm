//
//  ISSerialization.mm
//  LevelDBTest
//
//  Created by 舒彪 on 2018/1/27.
//  Copyright © 2018年 intsig. All rights reserved.
//

#import "ISSerialization.h"
#import <CoreGraphics/CoreGraphics.h>
#include <vector>

typedef struct ISSerializationPair {
  NSString *key;
  uint8_t keylen;
  NSData *data;
  uint32_t dataLen;
} ISSerializationPair;

typedef struct ISSerializationElement {
  NSData *data;
  uint32_t dataLen;
} ISSerializationElement;

@interface ISSerialization() {
  std::vector<ISSerializationPair> _pairs;
  std::vector<ISSerializationElement> _elements;
}
@property (nonatomic, strong) NSDictionary *unSerializedDic;
@property (nonatomic, strong) NSArray<NSData *> *unSerializedArray;
@end

@implementation ISSerialization

#pragma mark - ForDictionay
- (void)setData:(NSData *)data forKey:(NSString *)key {
  if (!key || !data) {
    return;
  }
  ISSerializationPair pair = {
    key,
    (uint8_t)key.length,
    data,
    (uint32_t)data.length
  };
  _pairs.push_back(pair);
}

- (void)setDouble:(double)doubleValue forKey:(NSString *)key {
  if (!key) {
    return;
  }
  
  NSData *newData = [[NSData alloc] initWithBytes:&doubleValue
                                           length:sizeof(doubleValue)];
  if (newData) {
    [self setData:newData forKey:key];
  }
}

- (void)setInteger:(NSInteger)integer forKey:(NSString *)key {
  if (!key) {
    return;
  }
  
  NSData *newData = [[NSData alloc] initWithBytes:&integer
                                           length:sizeof(NSInteger)];
  if (newData) {
    [self setData:newData forKey:key];
  }
}

- (void)setString:(NSString *)string forKey:(NSString *)key {
  if (!key || !string) {
    return;
  }
  
  NSData *newData = [string dataUsingEncoding:NSUTF8StringEncoding];
  if (newData) {
    [self setData:newData forKey:key];
  }
}

- (NSData *)generateDataFromDictionary {
  // |----------------------------|
  // | 1byte | 4byte | key | data |
  // |----------------------------|
  // 第1个字节保存key的长度 第2-5个字节保存data的长度
  // key存储key的数据，data存储data的数据
  // 通过以上数据可以算出key、data以及下一个pair的偏移
  NSUInteger bytesLen = 0;
  
  for (std::vector<ISSerializationPair>::iterator it = _pairs.begin() ;
       it != _pairs.end();
       ++it) {
    bytesLen += it->keylen + it->dataLen + 5;
  }
  
  if (bytesLen) {
    char *bytes = (char *)malloc(sizeof(char) * bytesLen);
    char *cursor = bytes;
    for (std::vector<ISSerializationPair>::iterator it = _pairs.begin() ;
         it != _pairs.end();
         ++it) {
      uint8_t keyLen = it->keylen;
      uint32_t dataLen = it->dataLen;
      memcpy(cursor, &keyLen, 1);
      cursor += 1;
      memcpy(cursor, &dataLen, 4);
      cursor += 4;
      memcpy(cursor, [it->key cStringUsingEncoding:NSUTF8StringEncoding], it->keylen);
      cursor += it->keylen;
      memcpy(cursor, it->data.bytes, it->dataLen);
      cursor += it->dataLen;
    }
    NSData *combinedData = [NSData dataWithBytesNoCopy:(void *)bytes
                                                length:bytesLen];
    return combinedData;
  }
  return nil;
}

- (instancetype)initWithSerializedDictionaryData:(NSData *)data {
  self = [super init];
  if (!self) {
    return nil;
  }
  
  if (data.length == 0) {
    return nil;
  }
  
  NSMutableDictionary *dic = [[NSMutableDictionary alloc] init];
  
  char *bytes = (char *)data.bytes;
  char *cursor = bytes;
  while (cursor - bytes < data.length) {
    uint8_t keylen = 0;
    memcpy(&keylen, cursor, 1);
    cursor += 1;
    uint32_t datalen = 0;
    memcpy(&datalen, cursor, 4);
    cursor += 4;
    char *key = (char *)malloc(sizeof(char) * keylen);
    memcpy(key, cursor, keylen);
    cursor += keylen;
    char *data = (char *)malloc(sizeof(char) * datalen);
    memcpy(data, cursor, datalen);
    cursor += datalen;
    NSString *ocKey = [[NSString alloc] initWithBytes:key
                                               length:keylen
                                             encoding:NSUTF8StringEncoding];
    free(key);
    NSData *ocData = [NSData dataWithBytes:data length:datalen];
    free(data);
    [dic setObject:ocData forKey:ocKey];
  }
  _unSerializedDic = dic;
  return self;
}

- (NSData *)dataForKey:(NSString *)key {
  return [self.unSerializedDic objectForKey:key];
}

- (double)doubleForKey:(NSString *)key {
  NSData *data = [_unSerializedDic objectForKey:key];
  double result = CGFLOAT_MAX;
  if (data) {
    [data getBytes:&result length:sizeof(double)];
  }
  return result;
}

- (NSInteger)integerForKey:(NSString *)key {
  NSData *data = [_unSerializedDic objectForKey:key];
  NSInteger result = NSNotFound;
  if (data) {
    [data getBytes:&result length:sizeof(NSInteger)];
  }
  return result;
}

- (NSString *)stringForKey:(NSString *)key {
  NSData *data = [_unSerializedDic objectForKey:key];
  if (data) {
    return [[NSString alloc] initWithData:data
                                 encoding:NSUTF8StringEncoding];
  }
  return nil;
}

#pragma mark - ForArray

- (void)appendData:(NSData *)data {
  if (!data) {
    return;
  }
  ISSerializationElement element = {
    data,
    (uint32_t)data.length
  };
  _elements.push_back(element);
}

- (void)appendDouble:(double)doubleValue {
  NSData *newData = [[NSData alloc] initWithBytes:&doubleValue
                                           length:sizeof(doubleValue)];
  if (newData) {
    [self appendData:newData];
  }
}

- (void)appendInteger:(NSInteger)integer {
  NSData *newData = [[NSData alloc] initWithBytes:&integer
                                           length:sizeof(NSInteger)];
  if (newData) {
    [self appendData:newData];
  }
}

- (void)appendString:(NSString *)string {
  if (!string) {
    return;
  }
  NSData *newData = [string dataUsingEncoding:NSUTF8StringEncoding];
  if (newData) {
    [self appendData:newData];
  }
}

- (NSData *)generateDataFromArray {
  // |---------------|
  // | 4byte | data |
  // |---------------|
  // 前4个字节保存data的长度
  // data存储data的数据
  NSUInteger bytesLen = 0;
  
  for (std::vector<ISSerializationElement>::iterator it = _elements.begin() ;
       it != _elements.end();
       ++it) {
    bytesLen += it->dataLen + 4;
  }
  
  if (bytesLen) {
    char *bytes = (char *)malloc(sizeof(char) * bytesLen);
    char *cursor = bytes;
    for (std::vector<ISSerializationElement>::iterator it = _elements.begin() ;
         it != _elements.end();
         ++it) {
      uint32_t dataLen = it->dataLen;
      memcpy(cursor, &dataLen, 4);
      cursor += 4;
      memcpy(cursor, it->data.bytes, it->dataLen);
      cursor += it->dataLen;
    }
    NSData *combinedData = [NSData dataWithBytesNoCopy:(void *)bytes
                                                length:bytesLen];
    return combinedData;
  }
  return nil;
}

- (instancetype)initWithSerializedArrayData:(NSData *)data {
  self = [super init];
  if (!self) {
    return nil;
  }
  
  if (data.length == 0) {
    return nil;
  }
  
  NSMutableArray *array = [[NSMutableArray alloc] init];
  
  char *bytes = (char *)data.bytes;
  char *cursor = bytes;
  while (cursor - bytes < data.length) {
    uint32_t datalen = 0;
    memcpy(&datalen, cursor, 4);
    cursor += 4;
    char *data = (char *)malloc(sizeof(char) * datalen);
    memcpy(data, cursor, datalen);
    cursor += datalen;
    NSData *ocData = [NSData dataWithBytesNoCopy:data length:datalen];
    if (ocData) {
      [array addObject:ocData];
    }
  }
  _unSerializedArray = array;
  return self;
}

- (NSData *)dataAtIndex:(NSInteger)index {
  if (index < _unSerializedArray.count && index >= 0) {
     return [_unSerializedArray objectAtIndex:index];
  }
  return nil;
}

- (double)doubleAtIndex:(NSInteger)index {
  NSData *data = [self dataAtIndex:index];
  double result = CGFLOAT_MAX;
  if (data) {
    [data getBytes:&result length:sizeof(double)];
  }
  return result;
}

- (NSInteger)integerAtIndex:(NSInteger)index {
  NSData *data = [self dataAtIndex:index];
  NSInteger result = NSNotFound;
  if (data) {
    [data getBytes:&result length:sizeof(NSInteger)];
  }
  return result;
}

- (NSString *)stringAtIndex:(NSInteger)index {
  NSData *data = [self dataAtIndex:index];
  if (data) {
    return [[NSString alloc] initWithData:data encoding:NSUTF8StringEncoding];
  }
  return nil;
}
@end
