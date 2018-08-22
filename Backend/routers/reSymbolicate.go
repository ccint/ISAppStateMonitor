package routers

import (
	"net/http"
	"../reportStore"
	"encoding/json"
	"../logger"
)


func HanleReSymbolicate (w http.ResponseWriter, req *http.Request) {
	reportId := req.URL.Query().Get("report_id")

	report := reportStore.GetReportOfId(reportId)
	report.Symbolicate()
	if err := report.UpdateToStorage(); err != nil {
		logger.Log.Error("update report failed: ", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(map[string] string {"ret": "0"})
}

func HandlerReClassfiedReports (w http.ResponseWriter, req *http.Request) {
	appid := req.URL.Query().Get("appid")

	if err := reportStore.SymbolicateUnClassfiedReports(appid); err != nil {
		logger.Log.Error("reclassfied reports failed: ", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(map[string] string {"ret": "0"})
}