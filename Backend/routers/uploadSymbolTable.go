package routers

import (
	"net/http"
	"os"
	"io"
	"log"
	"io/ioutil"
	"archive/zip"
	"path/filepath"
	"github.com/satori/go.uuid"
	"sync"
	"os/exec"
	"strings"
	"../symbolization"
	"../reportStore"
	"encoding/json"
	"time"
	"../logger"
)

func UploadDsymHandler(w http.ResponseWriter, req *http.Request) {
	handleDsymReq(w, req)
}

func handleDsymReq(w http.ResponseWriter, req *http.Request) {
	ret := make(map[string] interface{})
	ignoreRet := req.URL.Query().Get("ignore_ret")
	switch req.Method {
	case "POST":
		ret["ret"] = "-1"
		req.ParseMultipartForm(32 << 20)
		file, _, err := req.FormFile("file")
		if err != nil {
			logger.Log.Error("get upload file failed: ", err)
			ret["msg"] = err.Error()
			break
		}
		defer file.Close()
		fileUUID := uuid.Must(uuid.NewV4()).String()

		tmpPath := "./resource/tmp/"

		if _, err := os.Stat(tmpPath); err != nil && os.IsNotExist(err) {
			os.MkdirAll(tmpPath, 0755)
		}

		tmpFilePath := tmpPath + fileUUID + ".zip"
		f, err := os.OpenFile(tmpFilePath, os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			logger.Log.Error("open copy zip file failed: ", err)
			ret["msg"] = err.Error()
			break
		}
		defer f.Close()
		if _, err = io.Copy(f, file); err != nil {
			logger.Log.Error("copy zip file failed: ", err)
			ret["msg"] = err.Error()
			break
		}

		var result *[]map[string] string

		if ignoreRet == "1" {
			go handleDSYMFiles(tmpFilePath, fileUUID)
		} else {
			result = handleDSYMFiles(tmpFilePath, fileUUID)
			ret["data"] = result
		}
		ret["ret"] = "0"
	default:
		ret["ret"] = "-1"
		ret["msg"] = "get is invalid"
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(ret)
}

func handleDSYMFiles(filepath string, uuid string) *[]map[string] string {

	now := time.Now()
	logger.Log.Info("start handle dysm file")

	destDir := "./resource/tmp/symbols/" + uuid

	if err := Unzip(filepath, destDir); err != nil {
		log.Fatal(err)
		return nil
	}

	{
		files, err := ioutil.ReadDir(destDir)
		if err != nil {
			log.Fatal(err)
			return nil
		}

		var sg sync.WaitGroup
		sg.Add(len(files))

		for _, f := range files {
			var fPath = destDir + "/" + f.Name()
			genST(fPath, f.IsDir(), &sg)
		}

		sg.Wait()
	}

	{
		zips, err := ioutil.ReadDir(destDir)
		if err != nil {
			log.Fatal(err)
			return nil
		}

		for _, zf := range zips {
			var zfPath = destDir + "/" + zf.Name()
			if strings.HasSuffix(zf.Name(),".zip") {
				Unzip(zfPath, destDir)
			}
			os.Remove(zfPath)
		}
	}

	importResult := new([]map[string] string)

	{
		symbols, err := ioutil.ReadDir(destDir)
		if err != nil {
			log.Fatal(err)
			return nil
		}

		var sg sync.WaitGroup

		sg.Add(len(symbols))

		mutex := sync.Mutex{}

		for _, symbol := range symbols {
			var sPath = destDir + "/" + symbol.Name()
			go importDSYMTable(sPath, &sg, importResult, &mutex)
		}

		sg.Wait()
	}

	//if err := os.RemoveAll(destDir); err != nil {
	//	logger.Log.Error("clear symbol dir failed: ", err)
	//}

	if err := os.Remove(filepath); err != nil {
		logger.Log.Error("clear upload file failed: ", err)
	}

	logger.Log.Info("handle dysm finished, total cost time: ", time.Since(now))

	return importResult
}

func genST(fp string, isDir bool, group *sync.WaitGroup) {
	if group != nil {
		defer group.Done()
	}

	fPabsolute, _ := filepath.Abs(fp)
	tPabsolute, _ := filepath.Abs("./libs/symbolicate/buglySymboliOS.jar")

	cmd := exec.Command("java", "-jar", tPabsolute, "-i", fPabsolute)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		logger.Log.Error("generate symbol table failed: ", err)
	}

	if isDir {
		if err := os.RemoveAll(fp); err != nil {
			logger.Log.Error("remove dsym failed: ", err)
		}
	} else {
		if err := os.Remove(fp); err != nil {
			logger.Log.Error("remove dsym failed: ", err)
		}
	}
}

func importDSYMTable(filepath string, group *sync.WaitGroup, results *[]map[string] string, mutex *sync.Mutex) {
	defer group.Done()
	if stUUID, err := symbolization.ImportDSYMTable(filepath); err == nil && len(stUUID) > 0 {
		result := map[string] string{"uuid": stUUID}

		dsym, ok := reportStore.RemoveMissingDSYMS(stUUID)
		if ok {
			result["name"] = dsym.Name
		}

		mutex.Lock()
		*results = append(*results, result)
		mutex.Unlock()
	}
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