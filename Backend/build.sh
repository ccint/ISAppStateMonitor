#!/bin/bash

export CGO_CFLAGS="-Ipath/to/rocksdb/include -Ipath/to/project_dir/libs/demangle/usr/local/include"
export CGO_LDFLAGS="-Lpath/to/rocksdb/lib -Lpath/to/project_dir/libs/demangle -lrocksdb -ldemangle -lstdc++ -lm -lz -lbz2 -lsnappy -llz4"
go build server.go