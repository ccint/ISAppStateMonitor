package reportStore

import (
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"fmt"
	"time"
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
	identifier := s.getIssueIdentifier()

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

func (s *AnrReport) getIssueIdentifier() *string {
	if len(s.Backtrace.Stacks) == 0 || len(s.Backtrace.Stacks[0].Frames) == 0 {
		return nil
	}
	result := new(string)

	mainStack := s.Backtrace.Stacks[0]
	for _, frame := range  mainStack.Frames {
		if frame.ImageName == s.Backtrace.AppImageName && frame.RetSymbol != "main" {
			*result = frame.RetSymbol
			break
		}
	}
	if len(*result) == 0 {
		*result = mainStack.Frames[0].RetSymbol
	}

	return result
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

	var results *[]AnrReport

	c := session.DB(dataBase).C(reportCollection)
	err := c.Find(nil).All(results)
	if err != nil {
		fmt.Println(err)
	}
	return results
}
