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
)

var dysmDB *gorocksdb.DB
var ro *gorocksdb.ReadOptions
var wo *gorocksdb.WriteOptions

func openDB() {
	ro = gorocksdb.NewDefaultReadOptions()
	wo = gorocksdb.NewDefaultWriteOptions()

	bbto := gorocksdb.NewDefaultBlockBasedTableOptions()
	bbto.SetBlockCache(gorocksdb.NewLRUCache(3 << 30))
	opts := gorocksdb.NewDefaultOptions()
	opts.SetBlockBasedTableFactory(bbto)
	opts.SetCreateIfMissing(true)

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

		endAdr, err1 := strconv.ParseUint(elements[1], 16, 0)
		if err1 != nil {
			return errors.New("parse failed, invalid line :" + line)
		}

		symbols := elements[2]

		if len(symbols) <= 0 {
			return errors.New("parse failed, invalid line :" + line)
		}

		diff := endAdr - startAdr

		for i := 0; i < int(diff); i++ {
			key := uuid + "_" + arch + "_" + strconv.FormatUint(startAdr + uint64(i), 16)
			var value string
			if i == 0 {
				value = symbols
			} else {
				value = strconv.FormatUint(startAdr,16)
			}
			dysmDB.Put(wo, []byte(key), []byte(value))
		}
	}
	fmt.Printf("import %s succeed\n", filePath)
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

	v, err := dysmDB.Get(ro, []byte(uuid + "_" + arch + "_" + offsetStr))
	if err != nil {
		return "", err
	}

	vStr := string(v.Data())
	_, err0 := strconv.ParseUint(vStr, 16, 0)
	if err0 == nil {
		realKey := uuid + "_" + arch + "_" + vStr
		realV, err1 := dysmDB.Get(ro, []byte(realKey))
		if err1 != nil {
			return "", err1
		}
		return string(realV.Data()), nil
	} else {
		return string(vStr), nil
	}
}

func getFileName(filePath string) string {
	filenameWithSuffix := path.Base(filePath)
	fileSuffix := path.Ext(filenameWithSuffix)
	return strings.TrimSuffix(filenameWithSuffix, fileSuffix)
}
