package routers

import (
	"net/http"
	"strings"
	"strconv"
	"encoding/json"
	"../reportStore"
)

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
				splits := strings.Split(frame.RetSymbol, " ")
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
	issues := reportStore.GetAllIssues()

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

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(retIssues)
}

func HandleQueryIssues(w http.ResponseWriter, req *http.Request) {
	getAllIssues(w, req)
}