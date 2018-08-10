//
//  demangle.hpp
//  demangle
//
//  Created by 舒彪 on 2018/8/9.
//  Copyright © 2018年 舒彪. All rights reserved.
//

#ifndef demangle_h
#define demangle_h

#include <stdio.h>

#ifdef __cplusplus
extern "C" {
#endif
    const char *cpp_demangle(const char *input);
    
    const char *swift_demangle(const char *input);
#ifdef __cplusplus
}
#endif

#endif /* demangle_h */
