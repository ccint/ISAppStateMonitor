package symbolization

// #include "demangle.h"
import "C"
import (
	"github.com/tecbot/gorocksdb"
	"log"
	"errors"
	"strings"
	"os"
	"bufio"
	"strconv"
	"bytes"
	"time"
	"../logger"
	"../appDsymStore"
	"fmt"
)

var dysmDB *gorocksdb.DB
var wo *gorocksdb.WriteOptions

type  comparator struct {
}

func (s *comparator) Compare(a, b []byte) int {
	if len(a) == len(b) {
		return bytes.Compare(a, b)
	} else {
		return len(a) - len(b)
	}
}

func (s *comparator) Name() string {
	return "leveldb.BytewiseComparator"
}

func openDB() {
	wo = gorocksdb.NewDefaultWriteOptions()

	bbto := gorocksdb.NewDefaultBlockBasedTableOptions()
	bbto.SetBlockCache(gorocksdb.NewLRUCache(3 << 30))
	filter := gorocksdb.NewBloomFilter(10)
	bbto.SetFilterPolicy(filter)

	opts := gorocksdb.NewDefaultOptions()
	opts.SetBlockBasedTableFactory(bbto)
	opts.SetCreateIfMissing(true)
	opts.SetComparator(new(comparator))
	opts.SetPrefixExtractor(gorocksdb.NewFixedPrefixTransform(44))

	dbPath := "./resource/dsym/"

	if _, err := os.Stat(dbPath); err != nil && os.IsNotExist(err) {
		os.MkdirAll(dbPath, 0755)
	}

	db, err := gorocksdb.OpenDb(opts, dbPath)
	if err != nil {
		log.Fatal(err)
	} else {
		dysmDB = db
	}
}

func closeDB() {
	if dysmDB != nil {
		dysmDB.Close()
	}
}

func InitSymbolization() {
	openDB()
}

func ImportDSYMTable(filePath string) (string, error) {
	timeNow := time.Now()
	logger.Log.Info("start import symbols from: " + filePath)

	file, err := os.Open(filePath)
	if err != nil {
		return "", errors.New("open dysm file failed")
	}

	defer file.Close()

	input := bufio.NewScanner(file)

	writeBatch := gorocksdb.NewWriteBatch()

	defer writeBatch.Destroy()

	var uuid string

	for input.Scan() {
		line := input.Text()
		elements := strings.Split(line, "\u0009")
		if len(elements) == 2 && elements[0] == "UUID:" {
			uuid = strings.ToUpper(elements[1])
			if appDsymStore.IsAppDsymRecordExist(uuid) {
				logger.Log.Info(
					fmt.Sprintf("skip import %s, already exist, uuid %s",
						filePath, uuid))
				return uuid, nil
			}
		} else if len(elements) >= 3 && len(uuid) > 0 {
			startAdr, err0 := strconv.ParseUint(elements[0], 16, 0)
			if err0 != nil {
				return "", errors.New("parse failed, invalid line :" + line)
			}

			symbol := elements[2]

			if strings.HasPrefix(symbol,"_Z") || strings.HasPrefix(symbol,"__Z") {
				symbol = demangleCppSymbol(symbol)
			}

			if strings.HasPrefix(symbol,"_T") || strings.HasPrefix(symbol,"__T") {
				symbol = demangleSwiftSymbol(symbol)
			}

			if len(elements) >= 4 {
				symbol = symbol + "\u0009" + elements[3]
			}

			if len(symbol) <= 0 {
				return "", errors.New("parse failed, invalid line :" + line)
			}

			key := uuid + "_" + strconv.FormatUint(startAdr, 16)
			writeBatch.Put([]byte(key), []byte(symbol))
		}
	}

	writeErr := dysmDB.Write(wo, writeBatch)

	if writeErr != nil {
		logger.Log.Error("writeBath Error: ", writeErr)
		return "", writeErr
	} else {
		logger.Log.Info(
			fmt.Sprintf("import %s succeed, time cost: %s",
				filePath,
				time.Since(timeNow).String()))
	}

	return uuid, nil
}

func Symbol(offset uint64, uuid string) (string, error) {
	offsetStr := strconv.FormatUint(offset, 16)
	if len(offsetStr) <= 0 {
		return "", errors.New("invalid offset")
	}

	if  len(uuid) <= 0 {
		return "", errors.New("invalid uuid")
	}

	iterator := dysmDB.NewIterator(gorocksdb.NewDefaultReadOptions())

	defer iterator.Close()

	prefix := strings.ToUpper(uuid) + "_"
	iterator.SeekForPrev([]byte(prefix + offsetStr))
	if iterator.ValidForPrefix([]byte(prefix)) {
		value := string(iterator.Value().Data())
		if len(value) > 0 {
			return value, nil
		}
	}

	return  "", errors.New("symbol not found")
}

func demangleCppSymbol(symbol string) string {
	if strings.HasPrefix(symbol, "__Z") {
		symbol = strings.Replace(symbol, "__Z","_Z",1)
	}

	result := C.cpp_demangle(C.CString(symbol))
	resultString := C.GoString(result)
	return resultString
}

func demangleSwiftSymbol(symbol string) string {
	if strings.HasPrefix(symbol, "__T") {
		symbol = strings.Replace(symbol, "__T","_T",1)
	}
	result := C.swift_demangle(C.CString(symbol))
	resultString := C.GoString(result)
	return resultString
}