package main

import (
	"net/http"
	"./routers"
	"./symbolization"
)

func main() {
	routers.SetupReportHandler()
	symbolization.InitSymbolization()

	http.HandleFunc("/report", routers.ReportHandler)
	http.HandleFunc("/upload_dsym", routers.UploadDsymHandler)
	http.ListenAndServeTLS(":4000",
		 	             "./certificate/server.cer",
		                 "./certificate/server.key", nil)
}