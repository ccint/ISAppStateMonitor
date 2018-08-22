package routers

import (
	"net/http"
	"strings"
	"strconv"
	"encoding/json"
	"../reportStore"
)

func GetAllApp (w http.ResponseWriter, req *http.Request) {
	results := reportStore.GetAllApps()

	retApps := new([]map[string] string)
	for _, result := range *results {
		retApp := make(map[string] string)
		retApp["appName"] = result.AppName
		retApp["appIdentifier"] = result.AppIdentifier
		*retApps = append(*retApps, retApp)
	}

	ret := map[string] interface{} {"count": len(*results), "data": *retApps}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(ret)
}

func GetAllMissingDYSM (w http.ResponseWriter, req *http.Request) {
	appId := req.URL.Query().Get("appId")
	results := reportStore.GetAllMissingDSYMs(appId)

	ret := map[string] interface{} {"count": len(*results), "data": *results}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(ret)
}

func GetAllReportsOfIssue(w http.ResponseWriter, req *http.Request) {
	issueId := req.URL.Query().Get("id")

	reportIds := reportStore.GetReportsOfIssue(issueId)

	ret := map[string] interface{} {"id": issueId, "sessions": *reportIds}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(ret)
}

func GetReportDetail(w http.ResponseWriter, req *http.Request) {
	reportId := req.URL.Query().Get("id")

	report := reportStore.GetReportOfId(reportId)

	ret := make(map[string] interface{})
	ret["appVersion"] = report.AppVersion
	ret["date"] = report.Timestamp
	ret["duration"] = report.Duration
	ret["appImage"] = report.Backtrace.AppImageName

	var stacks []map[string] interface{}
	for _, stack := range report.Backtrace.Stacks {
		newStack := make(map[string] interface{})
		newStack["threadName"] = stack.ThreadName

		var frames []map[string] string
		for _, frame := range stack.Frames {
			newFrame := make(map[string] string)
			newFrame["imageName"] = frame.ImageName

			source := ""
			if frame.ImageName == report.Backtrace.AppImageName {
				splits := strings.Split(frame.RetSymbol, "\u0009")
				if len(splits) > 1 {
					source = splits[len(splits) - 1]
				}
			}

			newFrame["source"] = source
			newFrame["symbol"] = strings.Replace(frame.RetSymbol, source, "", -1)
			frames = append(frames, newFrame)
		}

		newStack["frames"] = frames
		stacks = append(stacks, newStack)
	}

	ret["stacks"] = stacks
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(ret)
}

func getAllIssues(w http.ResponseWriter, req *http.Request) {
	start, _ := strconv.ParseInt(req.URL.Query().Get("start"), 10, 64)
	pageSize, _ :=  strconv.ParseInt(req.URL.Query().Get("pageSize"), 10, 64)
	appId := req.URL.Query().Get("appId")

	totalCount, issues, unclassfiedCount := reportStore.GetAllIssues(int(start), int(pageSize), appId)

	ret := make(map[string] interface{})

	retIssues := new([]map[string] string)
	for _, issue := range *issues {
		retIssue := make(map[string] string)

		retIssue["title"] = issue.IssueSourceFile
		retIssue["detail"] = strings.Replace(issue.IssueIdentifier, issue.IssueSourceFile, "", -1)
		retIssue["version"] = issue.IssueAffectVersionStart + " - " + issue.IssueAffectVersionEnd
		retIssue["issueCount"] = strconv.FormatUint(issue.IssueCount, 10)
		retIssue["id"] = issue.IssueId.Hex()
		*retIssues = append(*retIssues, retIssue)
	}

	ret["total"] = totalCount
	ret["issues"] = *retIssues
	ret["unclassfiedCount"] = unclassfiedCount

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(ret)
}

func HandleQueryIssues(w http.ResponseWriter, req *http.Request) {
	getAllIssues(w, req)
}