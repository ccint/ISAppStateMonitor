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
	http.HandleFunc("/query_issues", routers.HandleQueryIssues)
	http.HandleFunc("/issue_detail", routers.GetAllReportsOfIssue)
	http.HandleFunc("/issue_session", routers.GetReportDetail)
	http.ListenAndServe(":4000", nil)
}