//
//  ISLevelDB.mm
//
//  Copyright 2011 Pave Labs. All rights reserved. 
//  See LICENCE for details.
//

#import "ISLevelDB.h"

#import <leveldb/db.h>
#import <leveldb/options.h>
#import <leveldb/cache.h>
#import <leveldb/filter_policy.h>
#import <leveldb/write_batch.h>

#include "Common.h"

#define ISMaybeAddSnapshotToOptions(_from_, _to_) \
    leveldb::ReadOptions __to_;\
    leveldb::ReadOptions * _to_ = &__to_;\
    _to_ = &_from_;

#define ISSeekToFirstOrKey(iter, key, _backward_) \
    (key != nil) ? iter->Seek(ISKeyFromStringOrData(key)) : \
    _backward_ ? iter->SeekToLast() : iter->SeekToFirst()

#define ISMoveCursor(_iter_, _backward_) \
    _backward_ ? iter->Prev() : iter->Next()

#define ISEnsureNSData(_obj_) \
    ([_obj_ isKindOfClass:[NSData class]]) ? _obj_ : \
    ([_obj_ isKindOfClass:[NSString class]]) ? [NSData dataWithBytes:[_obj_ cStringUsingEncoding:NSUTF8StringEncoding] \
                                                              length:[_obj_ lengthOfBytesUsingEncoding:NSUTF8StringEncoding]] : nil

#define ISAssertDBExists(_db_) \
    NSAssert(_db_ != NULL, @"Database reference is not existent (it has probably been closed)");

namespace {
    class BatchIterator : public leveldb::WriteBatch::Handler {
    public:
        void (^putCallback)(const leveldb::Slice& key, const leveldb::Slice& value);
        void (^deleteCallback)(const leveldb::Slice& key);
        
        virtual void Put(const leveldb::Slice& key, const leveldb::Slice& value) {
            putCallback(key, value);
        }
        virtual void Delete(const leveldb::Slice& key) {
            deleteCallback(key);
        }
    };
}

NSString * ISNSStringFromLevelDBKey(ISLevelDBKey * key) {
    return [[[NSString alloc] initWithBytes:key->data
                                    length:key->length
                                  encoding:NSUTF8StringEncoding] autorelease];
}
NSData   * ISNSDataFromLevelDBKey(ISLevelDBKey * key) {
    return [NSData dataWithBytes:key->data length:key->length];
}

NSString * ISGetLibraryPath() {
    NSArray *paths = NSSearchPathForDirectoriesInDomains(NSLibraryDirectory, NSUserDomainMask, YES);
    return [paths objectAtIndex:0];
}

NSString * const kISLevelDBChangeType                = @"changeType";
NSString * const kISLevelDBChangeTypePut             = @"put";
NSString * const kISLevelDBChangeTypeDelete          = @"del";
NSString * const kISLevelDBChangeValue               = @"value";
NSString * const kISLevelDBChangeKey                 = @"key";
NSString * const kISLevelDBErrorOccurredNotification = @"kLevelDBErrorOccurredNotification";

ISLevelDBOptions ISMakeLevelDBOptions() {
    return (ISLevelDBOptions) {true, true, false, false, true, 0, 0};
}

@interface ISLevelDB () {
    leveldb::DB * db;
    leveldb::ReadOptions readOptions;
    leveldb::WriteOptions writeOptions;
    const leveldb::Cache * cache;
    const leveldb::FilterPolicy * filterPolicy;
}

@property (nonatomic, readonly) leveldb::DB * db;

@end

@implementation ISLevelDB

@synthesize db   = db;
@synthesize path = _path;

+ (ISLevelDBOptions) makeOptions {
    return ISMakeLevelDBOptions();
}
- (id) initWithPath:(NSString *)path andName:(NSString *)name {
    ISLevelDBOptions opts = ISMakeLevelDBOptions();
    return [self initWithPath:path name:name andOptions:opts];
}
- (id) initWithPath:(NSString *)path name:(NSString *)name andOptions:(ISLevelDBOptions)opts {
    self = [super init];
    if (self) {
        _name = [name retain];
        _path = [path retain];
        
        leveldb::Options options;
        
        options.create_if_missing = opts.createIfMissing;
        options.paranoid_checks = opts.paranoidCheck;
        options.error_if_exists = opts.errorIfExists;
        
        if (!opts.compression)
            options.compression = leveldb::kNoCompression;
        
        if (opts.cacheSize > 0) {
            options.block_cache = leveldb::NewLRUCache(opts.cacheSize);
            cache = options.block_cache;
        } else
            readOptions.fill_cache = false;
        
        if (opts.createIntermediateDirectories) {
            NSString *dirpath = [path stringByDeletingLastPathComponent];
            NSFileManager *fm = [NSFileManager defaultManager];
            NSError *crError;
            
            BOOL success = [fm createDirectoryAtPath:dirpath
                         withIntermediateDirectories:true
                                          attributes:nil
                                               error:&crError];
            if (!success) {
                [_name release];
                [_path release];
                NSString *errorInfo = [NSString stringWithFormat:
                                       @"Problem creating parent directory: %@",
                                       crError];
                [[NSNotificationCenter defaultCenter] postNotificationName:kISLevelDBErrorOccurredNotification
                                                                    object:errorInfo];
                return nil;
            }
        }
        
        if (opts.filterPolicy > 0) {
            filterPolicy = leveldb::NewBloomFilterPolicy(opts.filterPolicy);;
            options.filter_policy = filterPolicy;
        }
        leveldb::Status status = leveldb::DB::Open(options, [_path UTF8String], &db);
        
        readOptions.fill_cache = true;
        writeOptions.sync = false;
        
        if(!status.ok()) {
            [_name release];
            [_path release];
            NSString *errorInfo = [NSString stringWithFormat:
                                   @"Problem creating LevelDB database: %s",
                                   status.ToString().c_str()];
            [[NSNotificationCenter defaultCenter] postNotificationName:kISLevelDBErrorOccurredNotification
                                                                object:errorInfo];
            return nil;
        }
        
        self.encoder = ^ NSData *(ISLevelDBKey *key, id object) {
#ifdef DEBUG
            static dispatch_once_t onceToken;
            dispatch_once(&onceToken, ^{
                NSLog(@"No encoder block was set for this database [%@]", name);
                NSLog(@"Using a convenience encoder/decoder pair using NSKeyedArchiver.");
            });
#endif
            return [NSKeyedArchiver archivedDataWithRootObject:object];
        };
        self.decoder = ^ id (ISLevelDBKey *key, NSData *data) {
            return [NSKeyedUnarchiver unarchiveObjectWithData:data];
        };
    }
    
    return self;
}

+ (id) databaseInLibraryWithName:(NSString *)name {
    ISLevelDBOptions opts = ISMakeLevelDBOptions();
    return [self databaseInLibraryWithName:name andOptions:opts];
}
+ (id) databaseInLibraryWithName:(NSString *)name
                      andOptions:(ISLevelDBOptions)opts {
    NSString *path = [ISGetLibraryPath() stringByAppendingPathComponent:name];
    ISLevelDB *ldb = [[[self alloc] initWithPath:path name:name andOptions:opts] autorelease];
    return ldb;
}

- (void) setSafe:(BOOL)safe {
    writeOptions.sync = safe;
}
- (BOOL) safe {
    return writeOptions.sync;
}
- (void) setUseCache:(BOOL)useCache {
    readOptions.fill_cache = useCache;
}
- (BOOL) useCache {
    return readOptions.fill_cache;
}

#pragma mark - Setters

- (void) setObject:(id)value forKey:(id)key {
    ISAssertDBExists(db);
    ISAssertKeyType(key);
    NSParameterAssert(value != nil);
    
    leveldb::Slice k = ISKeyFromStringOrData(key);
    ISLevelDBKey lkey = ISGenericKeyFromSlice(k);

    NSData *data = _encoder(&lkey, value);
    leveldb::Slice v = ISSliceFromData(data);
    
    leveldb::Status status = db->Put(writeOptions, k, v);
    
    if(!status.ok()) {
        NSString *errorInfo = [NSString stringWithFormat:
                               @"Problem storing key/value pair for key '%@' in database: %s",
                               key,
                               status.ToString().c_str()];
        [[NSNotificationCenter defaultCenter] postNotificationName:kISLevelDBErrorOccurredNotification
                                                            object:errorInfo];
    }
}
- (void) setValue:(id)value forKey:(NSString *)key {
    [self setObject:value forKey:key];
}
- (void) setObject:(id)value forKeyedSubscript:(id)key {
    [self setObject:value forKey:key];
}
- (void) addEntriesFromDictionary:(NSDictionary *)dictionary {
    [dictionary enumerateKeysAndObjectsUsingBlock:^(id key, id obj, BOOL *stop) {
        [self setObject:obj forKey:key];
    }];
}

#pragma mark - Getters

- (id) objectForKey:(id)key {
    
    ISAssertDBExists(db);
    ISAssertKeyType(key);
    std::string v_string;
    ISMaybeAddSnapshotToOptions(readOptions, readOptionsPtr);
    leveldb::Slice k = ISKeyFromStringOrData(key);
    leveldb::Status status = db->Get(*readOptionsPtr, k, &v_string);
    
    if(!status.ok()) {
        if(!status.IsNotFound()) {
            NSString *errorInfo = [NSString stringWithFormat:
                                   @"Problem retrieving value for key '%@' from database: %s",
                                   key,
                                   status.ToString().c_str()];
            [[NSNotificationCenter defaultCenter] postNotificationName:kISLevelDBErrorOccurredNotification
                                                                object:errorInfo];
        }
        return nil;
    }
    
    ISLevelDBKey lkey = ISGenericKeyFromSlice(k);
    return ISDecodeFromSlice(v_string, &lkey, _decoder);
}
- (id) objectsForKeys:(NSArray *)keys notFoundMarker:(id)marker {
    NSMutableArray *result = [NSMutableArray arrayWithCapacity:keys.count];
    [keys enumerateObjectsUsingBlock:^(id objId, NSUInteger idx, BOOL *stop) {
        id object = [self objectForKey:objId];
        if (object == nil) object = marker;
        result[idx] = object;
    }];
    return [NSArray arrayWithArray:result];
}
- (id) valueForKey:(NSString *)key {
    if ([key characterAtIndex:0] == '@') {
        return [super valueForKey:[key stringByReplacingCharactersInRange:(NSRange){0, 1}
                                                               withString:@""]];
    } else
        return [self objectForKey:key];
}
- (id) objectForKeyedSubscript:(id)key {
    return [self objectForKey:key];
}

- (BOOL) objectExistsForKey:(id)key {
    
    ISAssertDBExists(db);
    ISAssertKeyType(key);
    std::string v_string;
    ISMaybeAddSnapshotToOptions(readOptions, readOptionsPtr);
    leveldb::Slice k = ISKeyFromStringOrData(key);
    leveldb::Status status = db->Get(*readOptionsPtr, k, &v_string);
    
    if (!status.ok()) {
        if (status.IsNotFound())
            return false;
        else {
            NSString *errorInfo = [NSString stringWithFormat:
                                   @"Problem retrieving value for key '%@' from database: %s",
                                   key,
                                   status.ToString().c_str()];
            [[NSNotificationCenter defaultCenter] postNotificationName:kISLevelDBErrorOccurredNotification object:errorInfo];
            return NULL;
        }
    } else
        return true;
}

#pragma mark - Removers

- (void) removeObjectForKey:(id)key {
    ISAssertDBExists(db);
    ISAssertKeyType(key);
    
    leveldb::Slice k = ISKeyFromStringOrData(key);
    leveldb::Status status = db->Delete(writeOptions, k);
    
    if(!status.ok()) {
        NSString *errorInfo = [NSString stringWithFormat:
                               @"Problem deleting key/value pair for key '%@' in database: %s",
                               key,
                               status.ToString().c_str()];
        [[NSNotificationCenter defaultCenter] postNotificationName:kISLevelDBErrorOccurredNotification
                                                            object:errorInfo];
    }
}
- (void) removeObjectsForKeys:(NSArray *)keyArray {
    [keyArray enumerateObjectsUsingBlock:^(id obj, NSUInteger idx, BOOL *stop) {
        [self removeObjectForKey:obj];
    }];
}

- (void) removeAllObjects {
    [self removeAllObjectsWithPrefix:nil];
}
- (void) removeAllObjectsWithPrefix:(id)prefix {
    ISAssertDBExists(db);
    
    leveldb::Iterator * iter = db->NewIterator(readOptions);
    leveldb::Slice lkey;
    
    const void *prefixPtr;
    size_t prefixLen = 0;
    prefix = ISEnsureNSData(prefix);
    if (prefix) {
        prefixPtr = [(NSData *)prefix bytes];
        prefixLen = (size_t)[(NSData *)prefix length];
    }

    for (ISSeekToFirstOrKey(iter, (id)prefix, NO)
         ; iter->Valid()
         ; ISMoveCursor(iter, NO)) {
        
        lkey = iter->key();
        if (prefix && memcmp(lkey.data(), prefixPtr, MIN(prefixLen, lkey.size())) != 0)
            break;
        
        db->Delete(writeOptions, lkey);
    }
    delete iter;
}

#pragma mark - Selection

- (NSArray *)allKeys {
    NSMutableArray *keys = [[[NSMutableArray alloc] init] autorelease];
    [self enumerateKeysUsingBlock:^(ISLevelDBKey *key, BOOL *stop) {
        [keys addObject:ISNSDataFromLevelDBKey(key)];
    }];
    return [NSArray arrayWithArray:keys];
}
- (NSArray *)keysByFilteringWithPredicate:(NSPredicate *)predicate {
    NSMutableArray *keys = [[[NSMutableArray alloc] init] autorelease];
    [self enumerateKeysAndObjectsBackward:NO lazily:NO
                            startingAtKey:nil
                      filteredByPredicate:predicate
                                andPrefix:nil
                               usingBlock:^(ISLevelDBKey *key, id obj, BOOL *stop) {
                                   [keys addObject:ISNSDataFromLevelDBKey(key)];
                               }];
    return [NSArray arrayWithArray:keys];
}

- (NSDictionary *)dictionaryByFilteringWithPredicate:(NSPredicate *)predicate {
    NSMutableDictionary *results = [NSMutableDictionary dictionary];
    [self enumerateKeysAndObjectsBackward:NO lazily:NO
                            startingAtKey:nil
                      filteredByPredicate:predicate
                                andPrefix:nil
                               usingBlock:^(ISLevelDBKey *key, id obj, BOOL *stop) {
                                   [results setObject:obj forKey:ISNSDataFromLevelDBKey(key)];
                               }];
    
    return [NSDictionary dictionaryWithDictionary:results];
}

#pragma mark - Enumeration

- (void) _startIterator:(leveldb::Iterator*)iter
               backward:(BOOL)backward
                 prefix:(id)prefix
                  start:(id)key {
    
    const void *prefixPtr;
    size_t prefixLen;
    leveldb::Slice lkey, startingKey;
    
    prefix = ISEnsureNSData(prefix);
    if (prefix) {
        prefixPtr = [(NSData *)prefix bytes];
        prefixLen = (size_t)[(NSData *)prefix length];
        startingKey = leveldb::Slice((char *)prefixPtr, prefixLen);
        
        if (key) {
            leveldb::Slice skey = ISKeyFromStringOrData(key);
            if (skey.size() > prefixLen && memcmp(skey.data(), prefixPtr, prefixLen) == 0) {
                startingKey = skey;
            }
        }
        
        /*
         * If a prefix is provided and the iteration is backwards
         * we need to start on the next key (maybe discarding the first iteration)
         */
        if (backward) {
            signed long long i = startingKey.size() - 1;
            void * startingKeyPtr = malloc(startingKey.size());
            unsigned char *keyChar;
            memcpy(startingKeyPtr, startingKey.data(), startingKey.size());
            while (1) {
                if (i < 0) {
                    iter->SeekToLast();
                    break;
                }
                keyChar = (unsigned char *)startingKeyPtr + i;
                if (*keyChar < 255) {
                    *keyChar = *keyChar + 1;
                    iter->Seek(leveldb::Slice((char *)startingKeyPtr, startingKey.size()));
                    if (!iter->Valid()) {
                        iter->SeekToLast();
                    }
                    break;
                }
                i--;
            };
            free(startingKeyPtr);
            if (!iter->Valid())
                return;
            
            lkey = iter->key();
            if (startingKey.size() && prefix) {
                signed int cmp = memcmp(lkey.data(), startingKey.data(), startingKey.size());
                if (cmp > 0) {
                    iter->Prev();
                }
            }
        } else {
            // Otherwise, we start at the provided prefix
            iter->Seek(startingKey);
        }
    } else if (key) {
        iter->Seek(ISKeyFromStringOrData(key));
    } else if (backward) {
        iter->SeekToLast();
    } else {
        iter->SeekToFirst();
    }
}

- (void) enumerateKeysUsingBlock:(ISLevelDBKeyBlock)block {
    
    [self enumerateKeysBackward:FALSE
                  startingAtKey:nil
            filteredByPredicate:nil
                      andPrefix:nil
                     usingBlock:block];
}

- (void) enumerateKeysBackward:(BOOL)backward
                 startingAtKey:(id)key
           filteredByPredicate:(NSPredicate *)predicate
                     andPrefix:(id)prefix
                    usingBlock:(ISLevelDBKeyBlock)block {
    
    ISAssertDBExists(db);
    ISMaybeAddSnapshotToOptions(readOptions, readOptionsPtr);
    leveldb::Iterator* iter = db->NewIterator(*readOptionsPtr);
    leveldb::Slice lkey;
    BOOL stop = false;
    
    NSData *prefixData = ISEnsureNSData(prefix);
    
    ISLevelDBKeyValueBlock iterate = (predicate != nil)
        ? ^(ISLevelDBKey *lk, id value, BOOL *stop) {
            if ([predicate evaluateWithObject:value])
                block(lk, stop);
          }
        : ^(ISLevelDBKey *lk, id value, BOOL *stop) {
            block(lk, stop);
          };
    
    for ([self _startIterator:iter backward:backward prefix:prefix start:key]
         ; iter->Valid()
         ; ISMoveCursor(iter, backward)) {
        
        lkey = iter->key();
        if (prefix && memcmp(lkey.data(), [prefixData bytes], MIN((size_t)[prefixData length], lkey.size())) != 0)
            break;
        
        ISLevelDBKey lk = ISGenericKeyFromSlice(lkey);
        id v = (predicate == nil) ? nil : ISDecodeFromSlice(iter->value(), &lk, _decoder);
        iterate(&lk, v, &stop);
        if (stop) break;
    }
    
    delete iter;
}

- (void) enumerateKeysAndObjectsUsingBlock:(ISLevelDBKeyValueBlock)block {
    [self enumerateKeysAndObjectsBackward:FALSE
                                   lazily:FALSE
                            startingAtKey:nil
                      filteredByPredicate:nil
                                andPrefix:nil
                               usingBlock:block];
}

- (void) enumerateKeysAndObjectsBackward:(BOOL)backward
                                  lazily:(BOOL)lazily
                           startingAtKey:(id)key
                     filteredByPredicate:(NSPredicate *)predicate
                               andPrefix:(id)prefix
                              usingBlock:(id)block{
    
    ISAssertDBExists(db);
    ISMaybeAddSnapshotToOptions(readOptions, readOptionsPtr);
    leveldb::Iterator* iter = db->NewIterator(*readOptionsPtr);
    leveldb::Slice lkey;
    BOOL stop = false;
    
    ISLevelDBLazyKeyValueBlock iterate = (predicate != nil)
    
        // If there is a predicate:
        ? ^ (ISLevelDBKey *lk, ISLevelDBValueGetterBlock valueGetter, BOOL *stop) {
            // We need to get the value, whether the `lazily` flag was set or not
            id value = valueGetter();
            
            // If the predicate yields positive, we call the block
            if ([predicate evaluateWithObject:value]) {
                if (lazily)
                    ((ISLevelDBLazyKeyValueBlock)block)(lk, valueGetter, stop);
                else
                    ((ISLevelDBKeyValueBlock)block)(lk, value, stop);
            }
        }
    
        // Otherwise, we call the block
        : ^ (ISLevelDBKey *lk, ISLevelDBValueGetterBlock valueGetter, BOOL *stop) {
            if (lazily)
                ((ISLevelDBLazyKeyValueBlock)block)(lk, valueGetter, stop);
            else
                ((ISLevelDBKeyValueBlock)block)(lk, valueGetter(), stop);
        };
    
    NSData *prefixData = ISEnsureNSData(prefix);
    
    ISLevelDBValueGetterBlock getter;
    for ([self _startIterator:iter backward:backward prefix:prefix start:key]
         ; iter->Valid()
         ; ISMoveCursor(iter, backward)) {
        
        lkey = iter->key();
        // If there is prefix provided, and the prefix and key don't match, we break out of iteration
        if (prefix && memcmp(lkey.data(), [prefixData bytes], MIN((size_t)[prefixData length], lkey.size())) != 0)
            break;
        
        __block ISLevelDBKey lk = ISGenericKeyFromSlice(lkey);
        __block id v = nil;
        
        getter = ^ id {
            if (v) return v;
            v = ISDecodeFromSlice(iter->value(), &lk, _decoder);
            return v;
        };
        
        iterate(&lk, getter, &stop);
        if (stop) break;
    }
    
    delete iter;
}

#pragma mark - Bookkeeping

- (void) deleteDatabaseFromDisk {
    [self close];
    NSFileManager *fileManager = [NSFileManager defaultManager];
    NSError *error;
    [fileManager removeItemAtPath:_path error:&error];
}

- (void) close {
    @synchronized(self) {
        if (db) {
            delete db;
            
            if (cache)
                delete cache;
            
            if (filterPolicy)
                delete filterPolicy;
            
            db = NULL;
        }
    }
}
- (BOOL) closed {
    return db == NULL;
}
- (void) dealloc {
    [self close];
    if (_path) [_path release];
    if (_name) [_name release];
    if (_encoder) [_encoder release];
    if (_decoder) [_decoder release];
    [super dealloc];
}

@end
