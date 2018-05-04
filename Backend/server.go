package main

import (
	"net/http"
	"./symbolization"
	"./serialization"
	"encoding/json"
	"io/ioutil"
	"fmt"
	"os"
	"io"
	"archive/zip"
	"path/filepath"
	"path"
	"strings"
	"log"
)

func main() {
	symbolization.OpenDB()
	//initQueue()

	http.HandleFunc("/report", reportHandler)
	http.HandleFunc("/upload_dsym", uploadDsymHandler)
	http.ListenAndServeTLS(":4000",
		 	             "/Users/Sky/Desktop/serverCer/server.cer",
		                 "/Users/Sky/Desktop/serverCer/server.key", nil)
}

type ResultStruct struct {
	Ret int
	ErrMsg string
}

func reportHandler(w http.ResponseWriter, req *http.Request) {
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

	//go cacheReport(&data)
	printReport(&data)
}

func cacheReport(report *[]byte) {
	readChan <- report
}

func persistReport(report *[]byte) {
	<- limitChan
}

func printReport(report *[]byte) {
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

	data0, _ := dataDic.BytesWithKey("data")
	dataArray := serialization.NewAutoSerializedArray()
	dataArray.SetSerializedBytes(data0)
	bsData, _ := dataArray.BytesAtIndex(0)
	bsDic := serialization.NewAutoSerializedDic()
	bsDic.SetSerializedBytes(bsData)
	dur, _ := bsDic.Float64WithKey("dur")
	time, _ := bsDic.Float64WithKey("time")
	fmt.Printf("Runloop Duration: %f\n", dur)
	fmt.Printf("Record Time: %f\n", time)

	bsData0, _ := bsDic.BytesWithKey("bs")
	bsDetailDic := serialization.NewAutoSerializedDic()
	bsDetailDic.SetSerializedBytes(bsData0)
	bsArrayData, _ := bsDetailDic.BytesWithKey("bs")
	uuidMap := make(map[string] string)
	allKeys := bsDetailDic.Allkeys()
	for i := 0; i < len(allKeys); i++ {
		v, _ := bsDetailDic.StringWithKey(allKeys[i])
		uuidMap[allKeys[i]] = *v
	}
	bsArray := serialization.NewAutoSerializedArray()
	bsArray.SetSerializedBytes(bsArrayData)
	fmt.Println("Thread BackTrace: ")
	for i := 0; i < bsArray.Count(); i++ {
		threadDic := serialization.NewAutoSerializedDic()
		threadData, _ := bsArray.BytesAtIndex(i)
		threadDic.SetSerializedBytes(threadData)
		threadName, ok := threadDic.StringWithKey("thread_name")
		if ok {
			fmt.Println("Thread Name: " + *threadName)
		} else {
			fmt.Println("Thread Name: " + "")
		}
		threadBsData, _ := threadDic.BytesWithKey("th_stack")
		threadBsArray := serialization.NewAutoSerializedArray()
		threadBsArray.SetSerializedBytes(threadBsData)
		for i := 0; i < threadBsArray.Count(); i++ {
			threadBsDicData, _ := threadBsArray.BytesAtIndex(i)
			threadBsDic := serialization.NewAutoSerializedDic()
			threadBsDic.SetSerializedBytes(threadBsDicData)
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
				fmt.Printf("%s \t %x \t %s\n", *modeName, retAdr, symbol)
			} else {
				fmt.Printf("%s \t %x \t %x\n", *modeName, retAdr, loadAdr)
			}
		}
		fmt.Printf("\n")
	}
}

func uploadDsymHandler(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case "POST":
		req.ParseMultipartForm(32 << 20)
		file, handler, err := req.FormFile("file")
		if err != nil {
			fmt.Println(err)
			return
		}
		defer file.Close()
		f, err := os.OpenFile("./resource/tmp/" + handler.Filename, os.O_WRONLY|os.O_CREATE, 0666)  // 此处假设当前目录下已存在test目录
		if err != nil {
			fmt.Println(err)
			return
		}
		defer f.Close()
		if _, err = io.Copy(f, file); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			fmt.Println(err)
			return
		}
		fmt.Fprintf(w, "true")
		go handleDSYMFile("./resource/tmp/" + handler.Filename, getFileName(handler.Filename))

	default:
		fmt.Println("get")
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func handleDSYMFile(filepath string, uuid string) {
	destDir := "./resource/tmp/symbols/" + uuid
	if err := Unzip(filepath, destDir); err != nil {
		log.Fatal(err)
		return
	}

	files, err := ioutil.ReadDir(destDir)
	if err != nil {
		log.Fatal(err)
	}

	for _, f := range files {
		symbolization.ImportDSYMTable(destDir + "/" + f.Name(), uuid)
	}
	os.RemoveAll(destDir)
	os.Remove(filepath)
}

func Unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer func() {
		if err := r.Close(); err != nil {
			panic(err)
		}
	}()

	os.MkdirAll(dest, 0755)

	// Closure to address file descriptors issue with all the deferred .Close() methods
	extractAndWriteFile := func(f *zip.File) error {
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer func() {
			if err := rc.Close(); err != nil {
				panic(err)
			}
		}()

		fPath := filepath.Join(dest, f.Name)

		if f.FileInfo().IsDir() {
			os.MkdirAll(fPath, f.Mode())
		} else {
			os.MkdirAll(filepath.Dir(fPath), f.Mode())
			f, err := os.OpenFile(fPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			defer func() {
				if err := f.Close(); err != nil {
					panic(err)
				}
			}()

			_, err = io.Copy(f, rc)
			if err != nil {
				return err
			}
		}
		return nil
	}

	for _, f := range r.File {
		err := extractAndWriteFile(f)
		if err != nil {
			return err
		}
	}

	return nil
}

func getFileName(filePath string) string {
	filenameWithSuffix := path.Base(filePath)
	fileSuffix := path.Ext(filenameWithSuffix)
	return strings.TrimSuffix(filenameWithSuffix, fileSuffix)
}

// Thread Pool
var (
	chanNum = 6
	readChan  = make(chan *[]byte, 100)
	limitChan = make(chan bool, 6)
)

func initQueue() {
	for i := 0; i < chanNum; i++ {
		go Queue(readChan)
	}
}

func Queue(rchan chan *[]byte) {
	for {
		report := <-rchan
		limitChan <- true
		go persistReport(report)
	}
}

