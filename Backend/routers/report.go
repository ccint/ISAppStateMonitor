package routers

import (
	"net/http"
	"io/ioutil"
	"encoding/json"
	"fmt"
	"time"
	"../symbolization"
	"../serialization"
	"github.com/tecbot/gorocksdb"
	"log"
)

// Thread Pool
var (
	chanNum = 8
	readChan  = make(chan *[]byte, 100)
	limitChan = make(chan bool, 8)
)


func SetupReportHandler() {
	openDB()

	for i := 0; i < chanNum; i++ {
		go queue(readChan)
	}
}

func queue(rchan chan *[]byte) {
	for {
		report := <-rchan
		limitChan <- true
		go persistReport(report)
	}
}

var cacheDB *gorocksdb.DB
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

	db, err := gorocksdb.OpenDb(opts, "./reportCache/dbCaches/")
	if err != nil {
		log.Fatal(err)
	} else {
		cacheDB = db
	}
}

func closeDB() {
	if cacheDB != nil {
		cacheDB.Close()
	}
}

type ResultStruct struct {
	Ret int
	ErrMsg string
}

func ReportHandler(w http.ResponseWriter, req *http.Request) {
	data, err := ioutil.ReadAll(req.Body)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		payload := ResultStruct{-1, err.Error()}
		json.NewEncoder(w).Encode(payload)
		return
	}


	w.Header().Set("Content-Type", "application/json")
	payload := ResultStruct{0, ""}
	json.NewEncoder(w).Encode(payload)

	go cacheReport(&data)
}

func cacheReport(report *[]byte) {
	readChan <- report
}

func persistReport(report *[]byte) {
	printReport(report)
	<- limitChan
}

func printReport(report *[]byte) {
	timeNow := time.Now()
	dataDic := serialization.NewAutoSerializedDic()
	dataDic.SetSerializedBytes(report)

	appVersion, _ := dataDic.StringWithKey("app_ver")
	appId, _ := dataDic.StringWithKey("app_id")
	devUUID, _ := dataDic.StringWithKey("dev_uuid")
	arch, _ := dataDic.StringWithKey("arch")
	reportType, _ := dataDic.StringWithKey("type")
	fmt.Println("Report Start")
	fmt.Println("Type: " + *reportType)
	fmt.Println("AppVersion: " + *appVersion)
	fmt.Println("AppID: " + *appId)
	fmt.Println("Device: " + *devUUID)
	fmt.Println("Architecture: " + *arch + "\n")
	fmt.Println("Report Content: ")

	dataArray, _ := dataDic.ArrayWithKey("data")
	bsDic, _ := dataArray.DicAtIndex(0)
	dur, _ := bsDic.Float64WithKey("dur")
	date, _ := bsDic.Float64WithKey("time")
	fmt.Printf("Runloop Duration: %f\n", dur)
	fmt.Printf("Record Time: %f\n", date)

	bsDetailDic, _ := bsDic.DicWithKey("bs")
	uuidMap := make(map[string] string)
	allKeys := bsDetailDic.Allkeys()
	for i := 0; i < len(allKeys); i++ {
		v, _ := bsDetailDic.StringWithKey(allKeys[i])
		uuidMap[allKeys[i]] = *v
	}
	bsArray, _ := bsDetailDic.ArrayWithKey("bs")
	fmt.Println("Thread BackTrace: ")
	for i := 0; i < bsArray.Count(); i++ {
		threadDic, _ := bsArray.DicAtIndex(i)
		threadName, ok := threadDic.StringWithKey("thread_name")
		if ok {
			fmt.Println("Thread Name: " + *threadName)
		} else {
			fmt.Println("Thread Name: " + "")
		}
		threadBsArray, _ := threadDic.ArrayWithKey("th_stack")
		if threadBsArray != nil {
			for i := 0; i < threadBsArray.Count(); i++ {
				threadBsDic, _ := threadBsArray.DicAtIndex(i)
				modeName, _ := threadBsDic.StringWithKey("mod_name")
				retAdr, _ := threadBsDic.Uint64WithKey("ret_adr")
				loadAdr, _ := threadBsDic.Uint64WithKey("load_adr")
				var symbol string
				if v, ok := uuidMap[*modeName]; ok == true {
					offset := retAdr - loadAdr
					v, err := symbolization.Symbol(offset, v, *arch)
					if err != nil {
						fmt.Println("get symbol err: " + err.Error())
					} else {
						symbol = v
					}
				}
				if len(symbol) > 0 {
					fmt.Printf("%-2d %-30s  0x%x  %s\n", i, *modeName, retAdr, symbol)
				} else {
					fmt.Printf("%-2d %-30s  0x%x  0x%x\n", i, *modeName, retAdr, loadAdr)
				}
			}
			fmt.Printf("\n")
		}
	}

	fmt.Printf("PrintTime: ")
	fmt.Println(time.Since(timeNow))
}