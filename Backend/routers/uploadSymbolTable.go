package routers

import (
	"net/http"
	"fmt"
	"os"
	"io"
	"log"
	"io/ioutil"
	"archive/zip"
	"path/filepath"
	"path"
	"strings"
	"../symbolization"
)

func UploadDsymHandler(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case "POST":
		req.ParseMultipartForm(32 << 20)
		file, handler, err := req.FormFile("file")
		if err != nil {
			fmt.Println(err)
			return
		}
		defer file.Close()
		f, err := os.OpenFile("./resource/tmp/" + handler.Filename, os.O_WRONLY|os.O_CREATE, 0666)
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