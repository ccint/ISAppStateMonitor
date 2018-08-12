### environment
```
CGO_CFLAGS="-Ipath/to/rocksdb/include -Ipath/to/project_dir/libs/demangle/usr/local/include"
```
```
CGO_LDFLAGS="-Lpath/to/rocksdb/lib -Lpath/to/project_dir/libs/demangle -lrocksdb -ldemangle -lstdc++ -lm -lz -lbz2 -lsnappy -llz4"
```
### dependency
```
  go get github.com/tecbot/gorocksdb
  go get github.com/satori/go.uuid
  go get github.com/globalsign/mgo
  go get github.com/sirupsen/logrus
  go get github.com/lestrrat/go-file-rotatelogs
  go get github.com/rifflock/lfshook
```
