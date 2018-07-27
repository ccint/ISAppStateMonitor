package routers

import (
	"net/http"
	"strings"
	"strconv"
	"encoding/json"
	"../reportStore"
)

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
	go getAllIssues(w, req)
}