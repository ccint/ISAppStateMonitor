package main

import (
	"net/http"
	"./routers"
	"./symbolization"
	"./reportStore"
	"./logger"
)

func main() {
	logger.Init()
	routers.SetupReportHandler()
	symbolization.InitSymbolization()
	reportStore.InitMissingDsym()

	http.HandleFunc("/report", routers.ReportHandler)
	http.HandleFunc("/upload_dsym", routers.UploadDsymHandler)
	http.HandleFunc("/query_issues", routers.HandleQueryIssues)
	http.HandleFunc("/issue_detail", routers.GetAllReportsOfIssue)
	http.HandleFunc("/issue_session", routers.GetReportDetail)
	http.HandleFunc("/missing_dsym", routers.GetAllMissingDYSM)
	http.HandleFunc("/resymbolicate", routers.HanleReSymbolicate)

	go http.ListenAndServe(":4000", nil)
	http.ListenAndServeTLS(":4001", "./certificate/server.cer", "./certificate/server.key",nil)
}