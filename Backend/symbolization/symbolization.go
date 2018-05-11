package symbolization

import (
	"github.com/tecbot/gorocksdb"
	"log"
	"errors"
	"strings"
	"path"
	"os"
	"bufio"
	"strconv"
	"fmt"
	"bytes"
	"time"
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

	db, err := gorocksdb.OpenDb(opts, "./resource/dsym/")
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

func ImportDSYMTable(filePath string, uuid string) error {
	fileName := getFileName(filePath)
	fileInfos := strings.Split(fileName, "&")
	arch := fileInfos[len(fileInfos) - 2]
	timeNow := time.Now()
	fmt.Println("start import symbols from: " + filePath)

	if len(uuid) <= 0 {
		return errors.New("failed to get uuid")
	}

	file, err := os.Open(filePath)
	if err != nil {
		return errors.New("open dysm file failed")
	}

	defer file.Close()

	input := bufio.NewScanner(file)

	writeBatch := gorocksdb.NewWriteBatch()

	defer writeBatch.Destroy()

	for input.Scan() {
		line := input.Text()
		elements := strings.Split(line, "\u0009")
		if len(elements) != 3 {
			return errors.New("parse failed, invalid line :" + line)
		}
		startAdr, err0 := strconv.ParseUint(elements[0], 16, 0)
		if err0 != nil {
			return errors.New("parse failed, invalid line :" + line)
		}

		symbols := elements[2]

		if len(symbols) <= 0 {
			return errors.New("parse failed, invalid line :" + line)
		}

		key := uuid + "_" + arch + "_" + strconv.FormatUint(startAdr, 16)
		writeBatch.Put([]byte(key), []byte(symbols))
	}

	writeErr := dysmDB.Write(wo, writeBatch)

	if writeErr != nil {
		fmt.Println(writeErr)
		return writeErr
	} else {
		fmt.Printf("import %s succeed\n", filePath)
		fmt.Printf("time cost: ")
		fmt.Println(time.Since(timeNow))
	}

	return nil
}

func Symbol(offset uint64, uuid string, arch string) (string, error) {
	offsetStr := strconv.FormatUint(offset, 16)
	if len(offsetStr) <= 0 {
		return "", errors.New("invalid offset")
	}

	if  len(uuid) <= 0 {
		return "", errors.New("invalid uuid")
	}

	iterator := dysmDB.NewIterator(gorocksdb.NewDefaultReadOptions())

	defer iterator.Close()

	prefix := uuid + "_" + arch + "_"
	iterator.SeekForPrev([]byte(prefix + offsetStr))
	if iterator.ValidForPrefix([]byte(prefix)) {
		value := string(iterator.Value().Data())
		if len(value) > 0 {
			return value, nil
		}
	}

	return  "", nil
}

func getFileName(filePath string) string {
	filenameWithSuffix := path.Base(filePath)
	fileSuffix := path.Ext(filenameWithSuffix)
	return strings.TrimSuffix(filenameWithSuffix, fileSuffix)
}
