//
//  Header.h
//  Pods
//
//  Created by Mathieu D'Amours on 5/8/13.
//
//

#pragma once


#define ISAssertKeyType(_key_)\
    NSParameterAssert([_key_ isKindOfClass:[NSString class]] || [_key_ isKindOfClass:[NSData class]])

#define ISSliceFromString(_string_)           leveldb::Slice((char *)[_string_ UTF8String], [_string_ lengthOfBytesUsingEncoding:NSUTF8StringEncoding])
#define ISStringFromSlice(_slice_)            [[[NSString alloc] initWithBytes:_slice_.data() length:_slice_.size() encoding:NSUTF8StringEncoding] autorelease]

#define ISSliceFromData(_data_)               leveldb::Slice((char *)[_data_ bytes], [_data_ length])
#define ISDataFromSlice(_slice_)              [NSData dataWithBytes:_slice_.data() length:_slice_.size()]

#define ISDecodeFromSlice(_slice_, _key_, _d) _d(_key_, ISDataFromSlice(_slice_))
#define ISEncodeToSlice(_object_, _key_, _e)  ISSliceFromData(_e(_key_, _object_))

#define ISKeyFromStringOrData(_key_)          ([_key_ isKindOfClass:[NSString class]]) ? ISSliceFromString(_key_) \
                                            : ISSliceFromData(_key_)

#define ISGenericKeyFromSlice(_slice_)        (ISLevelDBKey) { .data = _slice_.data(), .length = _slice_.size() }
#define ISGenericKeyFromNSDataOrString(_obj_) ([_obj_ isKindOfClass:[NSString class]]) ? \
                                                (ISLevelDBKey) { \
                                                    .data   = [_obj_ cStringUsingEncoding:NSUTF8StringEncoding], \
                                                    .length = [_obj_ lengthOfBytesUsingEncoding:NSUTF8StringEncoding] \
                                                } \
                                            :   (ISLevelDBKey) { \
                                                    .data = [_obj_ bytes], \
                                                    .length = [_obj_ length] \
                                                }
