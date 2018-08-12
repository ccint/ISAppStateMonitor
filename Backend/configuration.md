CGO_CFLAGS="-I/usr/local/Cellar/rocksdb/5.12.4/include -I./utils/demangle/usr/local/include"
CGO_LDFLAGS="-L/usr/local/Cellar/rocksdb/5.12.4/lib -L./utils/demangle -lrocksdb -ldemangle -lstdc++ -lm -lz -lbz2 -lsnappy -llz4"

  go get github.com/tecbot/gorocksdb
  go get github.com/satori/go.uuid
  go get github.com/globalsign/mgo
  go get github.com/sirupsen/logrus
  go get github.com/lestrrat/go-file-rotatelogs
  go get github.com/rifflock/lfshook
