package reportStore

import (
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"fmt"
	"time"
	"strings"
)

var (
	mgoSession *mgo.Session
	hosts = 				[]string {"127.0.0.1:27017"}
	dataBase = 				"database"
	reportCollection =  	"anrReport"
	issueCollection =   	"issue"
)

type (
	Frame struct {
		ImageName   string
		RetAddress  uint64
		LoadAddress uint64
		RetSymbol   string
	}

	Stack struct {
		ThreadName string
		Frames     []Frame
	}

	Backtrace struct {
		IsSymbolized  bool
		AppImageName  string
		ImageMaps     map[string] string
		Stacks        []Stack
	}

	AnrReport struct {
		// extend data
		ReportId 	bson.ObjectId `bson:"_id"`
		AppVersion  string
		AppId 	    string
		DeviveUUID  string
		Arch 		string

		// anr data
		Duration    float64
		Timestamp   float64
		Backtrace   Backtrace
		Issue       bson.ObjectId
	}

	Issue struct {
		IssueId                 bson.ObjectId `bson:"_id"`
		IssueSourceFile 		string
		IssueIdentifier         string
		IssueCount              uint64
		IssueAffectVersionStart string
		IssueAffectVersionEnd   string
		IssueCreateTime         float64
		IssueLastUpdateTime     float64
	}
)

func getSession() *mgo.Session {
	if mgoSession == nil {
		var err error

		mongoDBDialInfo := &mgo.DialInfo{
			Addrs:     hosts,
			Direct:    false,
			Timeout:   time.Second * 1,
			PoolLimit: 4096,
		}

		mgoSession, err = mgo.DialWithInfo(mongoDBDialInfo)

		if err != nil {
			panic(err)
		}
	}
	return mgoSession.Copy()
}

func (s *AnrReport) Init() {
	s.ReportId = bson.NewObjectId()
}

func (s *AnrReport) SaveToStorage() error {
	session := getSession()
	defer session.Close()

	if isInBulk {
		s.updateOrCreateNewIssue(session)
		*reportBuffer = append(*reportBuffer, s)
		return nil
	} else {
		s.updateOrCreateNewIssue(session)
		c := session.DB(dataBase).C(reportCollection)
		return c.Insert(s)
	}
}

func (s *AnrReport) updateOrCreateNewIssue(session *mgo.Session) {
	identifier, sourceFile := s.getIssueIdentifierAndSourceFile()

	if identifier == nil {
		fmt.Println("error: identifier is nil")
		return
	}

	ic := session.DB(dataBase).C(issueCollection)

	issue := Issue{}

	err := ic.Find(bson.M{"issueidentifier": identifier}).One(&issue)
	var updateIssueErr error
	if err != nil {
		// issue not exist
		issue.Init()
		issue.IssueIdentifier = *identifier
		issue.IssueSourceFile = *sourceFile
		issue.IssueCreateTime = float64(time.Now().Unix() * 1000)
		issue.IssueCount = 1
		issue.IssueAffectVersionStart = s.AppVersion
		issue.IssueAffectVersionEnd = s.AppVersion
		issue.IssueLastUpdateTime = issue.IssueCreateTime
		updateIssueErr = ic.Insert(&issue)
	} else {
		// update exist issue
		change := bson.M{"issuelastupdatetime": float64(time.Now().Unix() * 1000),
			             "issuecount": issue.IssueCount + 1}

		if s.AppVersion < issue.IssueAffectVersionStart {
			change["issueaffectversionstart"] = s.AppVersion
		}

		if s.AppVersion > issue.IssueAffectVersionEnd {
			change["issueaffectversionend"] = s.AppVersion
		}
		updateIssueErr = ic.UpdateId(issue.IssueId, bson.M{"$set": change})
	}

	if updateIssueErr == nil {
		s.Issue = issue.IssueId
	}
}

func (s *AnrReport) getIssueIdentifierAndSourceFile() (*string, *string) {
	if len(s.Backtrace.Stacks) == 0 || len(s.Backtrace.Stacks[0].Frames) == 0 {
		return nil, nil
	}
	identifier := new(string)
	sourceFile := new(string)

	mainStack := s.Backtrace.Stacks[0]
	for _, frame := range  mainStack.Frames {
		if frame.ImageName == s.Backtrace.AppImageName && !strings.HasPrefix(frame.RetSymbol,"main main.") {
			*identifier = frame.RetSymbol
			splits := strings.Split(frame.RetSymbol, " ")
			if len(splits) > 1 {
				*sourceFile = splits[len(splits) - 1]
			} else {
				*sourceFile = frame.ImageName
			}
			return identifier, sourceFile
		}
	}

	*identifier = mainStack.Frames[0].RetSymbol
	*sourceFile = mainStack.Frames[0].ImageName
	return identifier, sourceFile
}

var reportBuffer *[]interface{}
var isInBulk bool

func BeginReportBulk() {
	if isInBulk == true {
		return
	}
	reportBuffer = new([]interface{})
	isInBulk = true
}

func FinishReportBulk() (error) {
	if len(*reportBuffer) == 0 {
		isInBulk = false
		return nil
	}
	session := getSession()
	defer session.Close()
	bulk := session.DB(dataBase).C(reportCollection).Bulk()
	bulk.Insert(*reportBuffer...)
	_, err := bulk.Run()
	isInBulk = false
	return err
}

func (s *Issue) Init() {
	s.IssueId = bson.NewObjectId()
}


func GetAllReports() *[]AnrReport {
	session := getSession()
	defer session.Close()

	var results []AnrReport

	c := session.DB(dataBase).C(reportCollection)
	err := c.Find(nil).All(&results)
	if err != nil {
		fmt.Println(err)
	}
	return &results
}

func GetAllIssues(start int, pageSize int) (int, *[]Issue) {
	session := getSession()
	defer session.Close()

	var results []Issue

	c := session.DB(dataBase).C(issueCollection)
	err := c.Find(nil).Sort("-issuecount").Skip(start).Limit(pageSize).All(&results)
	count, err :=  c.Find(nil).Count()
	if err != nil {
		fmt.Println(err)
	}

	return count, &results
}

func GetReportsOfIssue(issueId string) *[]string {
	session := getSession()
	defer session.Close()

	var results []Issue

	c := session.DB(dataBase).C(reportCollection)
	err := c.Find(bson.M{"issue": bson.ObjectIdHex(issueId)}).Sort("-timestamp").All(&results)
	if err != nil {
		fmt.Println(err)
	}

	var resultIds [] string
	for _, result := range results {
		resultIds = append(resultIds, result.IssueId.Hex())
	}
	return &resultIds
}

func GetReportOfId(reportId string) AnrReport {
	session := getSession()
	defer session.Close()

	var result AnrReport

	c := session.DB(dataBase).C(reportCollection)
	err := c.FindId(bson.ObjectIdHex(reportId)).One(&result)
	if err != nil {
		fmt.Println(err)
	}
	return result
}

