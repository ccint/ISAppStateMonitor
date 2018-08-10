//
//  demangle.cpp
//  demangle
//
//  Created by 舒彪 on 2018/8/9.
//  Copyright © 2018年 舒彪. All rights reserved.
//

#include "demangle.h"
#include <cxxabi.h>
#include "SwiftDemangle.h"
#include <stdlib.h>
#include <string>

const char *cpp_demangle(const char *input) {
    int status = 0;
    char *result = abi::__cxa_demangle(input, nullptr, nullptr, &status);
    if (status == 0 && result) {
        return result;
    }
    return input;
}

const char *swift_demangle(const char *input) {
    auto buffSize = std::min((int)strlen(input) * 3 + 256, 512);
    auto buffer = (char *)calloc(buffSize, sizeof(char));
    auto result = swift_demangle_getDemangledName(input, buffer, buffSize);
    if (result > 0) {
        return buffer;
    } else {
        free(buffer);
        return input;
    }
}
