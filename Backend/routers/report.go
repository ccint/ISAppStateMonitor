package routers

import (
	"net/http"
	"io/ioutil"
	"encoding/json"
	"fmt"
	"time"
	"../serialization"
	"../reportStore"
	"github.com/tecbot/gorocksdb"
	"log"
	"strings"
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
	handleReport(w, req)
}

func handleReport(w http.ResponseWriter, req *http.Request) {
	data, err := ioutil.ReadAll(req.Body)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		payload := ResultStruct{-1, err.Error()}
		json.NewEncoder(w).Encode(payload)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	payload := ResultStruct{0, "Hello Girl"}
	json.NewEncoder(w).Encode(payload)
	go cacheReport(&data)
}

func cacheReport(report *[]byte) {
	readChan <- report
}

func persistReport(report *[]byte) {
	archiveReport(report)
	<- limitChan
}

func archiveReport(report *[]byte) {
	timeNow := time.Now()
	dataDic := serialization.NewAutoSerializedDic()
	dataDic.SetSerializedBytes(report)

	appVersion, _ := dataDic.StringWithKey("app_ver")
	appId, _ := dataDic.StringWithKey("app_id")
	sysVersion, _ := dataDic.StringWithKey("sys_ver")
	arch, _ := dataDic.StringWithKey("arch")
	reportType, _ := dataDic.StringWithKey("type")

	if *reportType == "mt_out" {
		dataArray, _ := dataDic.ArrayWithKey("data")

		for i := 0; i < dataArray.Count(); i++ {
			dataDic, _ := dataArray.DicAtIndex(0)
			dur, _ := dataDic.Float64WithKey("dur")
			date, _ := dataDic.Float64WithKey("time")

			anrReport := reportStore.AnrReport{}
			anrReport.Init()
			anrReport.AppVersion = *appVersion
			anrReport.AppId = *appId
			anrReport.SysVersion = *sysVersion
			anrReport.Arch = *arch

			anrReport.Duration = dur
			anrReport.Timestamp = date

			backtrace := reportStore.Backtrace{}

			bsDetailDic, _ := dataDic.DicWithKey("bs")
			uuidMap, _ := bsDetailDic.DicWithKey("images")
			appImageName, _ := bsDetailDic.StringWithKey("appImageName")
			imageMaps := make(map[string] string)

			allKeys := uuidMap.Allkeys()
			for i := 0; i < len(allKeys); i++ {
				v, _ := uuidMap.StringWithKey(allKeys[i])
				imageMaps[allKeys[i]] = strings.Replace(*v, "-", "", -1)
			}

			backtrace.ImageMaps = imageMaps
			backtrace.AppImageName = *appImageName

			bsArray, _ := bsDetailDic.ArrayWithKey("bs")
			stacks := new([]reportStore.Stack)
			for i := 0; i < bsArray.Count(); i++ {
				stack := reportStore.Stack{}
				threadDic, _ := bsArray.DicAtIndex(i)
				threadName, ok := threadDic.StringWithKey("thread_name")
				if ok {
					stack.ThreadName = *threadName
				}

				threadBsArray, _ := threadDic.ArrayWithKey("th_stack")
				if threadBsArray != nil {
					frames := new([]reportStore.Frame)
					for i := 0; i < threadBsArray.Count(); i++ {
						frame := reportStore.Frame{}
						threadBsDic, _ := threadBsArray.DicAtIndex(i)
						modeName, _ := threadBsDic.StringWithKey("mod_name")
						retAdr, _ := threadBsDic.Uint64WithKey("ret_adr")
						loadAdr, _ := threadBsDic.Uint64WithKey("load_adr")

						frame.ImageName = *modeName
						frame.RetAddress = retAdr
						frame.LoadAddress = loadAdr

						*frames = append(*frames, frame)
					}
					stack.Frames = *frames
				}
				*stacks = append(*stacks, stack)
			}
			backtrace.Stacks = *stacks
			anrReport.Backtrace = backtrace
			anrReport.Symbolicate()
			anrReport.SaveToStorage()
		}
	}
	fmt.Printf("saveTime: ")
	fmt.Println(time.Since(timeNow))
}