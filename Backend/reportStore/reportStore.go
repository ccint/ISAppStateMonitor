package reportStore

import (
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"time"
	"strings"
	"errors"
	"sync"
	"../symbolization"
	"../logger"
)

var (
	mgoSession *mgo.Session
	hosts             = 	[]string {"127.0.0.1:27017"}
	dataBase          =  	"database"
	reportCollection  =  	"anrReport"
	issueCollection   =   	"issue"
	dsymCollection    =     "dsym"
	appCollection     =     "app"
	missingDSYMs      =     sync.Map{}
	issues            =     sync.Map{}
	apps 			  =     sync.Map{}
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
		SysVersion  string
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
		AppId			        bson.ObjectId `bson:"_id"`
	}

	MissingDSYM struct {
		Id				bson.ObjectId `bson:"_id"`
		Name 			string
		UUID			string
		Arch			string
		AppName			string
		AppVersion 		string
		SystemVersion	string
		AppId			bson.ObjectId `bson:"_id"`
	}

	App struct {
		Id				bson.ObjectId `bson:"_id"`
		AppName			string
		AppIdentifier   string
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
		s.updateOrCreateNewIssueAndApp(session)
		*reportBuffer = append(*reportBuffer, s)
		return nil
	} else {
		s.updateOrCreateNewIssueAndApp(session)
		c := session.DB(dataBase).C(reportCollection)
		return c.Insert(s)
	}
}

func (s *AnrReport)UpdateToStorage() error {
	session := getSession()
	defer session.Close()
	c := session.DB(dataBase).C(reportCollection)
	return c.UpdateId(s.ReportId, s)
}

func (s *AnrReport) updateOrCreateNewIssueAndApp(session *mgo.Session) {
	identifier, sourceFile := s.getIssueIdentifierAndSourceFile()

	if identifier == nil {
		logger.Log.Error("identifier is nil")
		return
	}

	var updateIssueErr error
	if issue, ok := getIssue(s.AppId, *identifier); ok == true {
		// update exist issue
		ic := session.DB(dataBase).C(issueCollection)
		change := bson.M{"issuelastupdatetime": float64(time.Now().Unix() * 1000),
			"issuecount": issue.IssueCount + 1}

		if s.AppVersion < issue.IssueAffectVersionStart {
			change["issueaffectversionstart"] = s.AppVersion
		}

		if s.AppVersion > issue.IssueAffectVersionEnd {
			change["issueaffectversionend"] = s.AppVersion
		}
		addIssueToCache(issue)
		updateIssueErr = ic.UpdateId(issue.IssueId, bson.M{"$set": change})
	} else {
		// issue not exist
		issue.Init()
		issue.IssueIdentifier = *identifier
		issue.IssueSourceFile = *sourceFile
		issue.IssueCreateTime = float64(time.Now().Unix() * 1000)
		issue.IssueCount = 1
		issue.IssueAffectVersionStart = s.AppVersion
		issue.IssueAffectVersionEnd = s.AppVersion
		issue.IssueLastUpdateTime = issue.IssueCreateTime
		issue.AppId = bson.ObjectIdHex(s.AppId)
		updateIssueErr = issue.SaveToStorage(session)
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
		if frame.ImageName == s.Backtrace.AppImageName && !strings.Contains(frame.RetSymbol,"main.m") {
			*identifier = frame.RetSymbol
			splits := strings.Split(frame.RetSymbol, "\u0009")
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

func (s *AnrReport)Symbolicate() {
	bs := &(s.Backtrace)
	imagesMap := bs.ImageMaps

	stacks := &(bs.Stacks)

	for i := 0; i < len(*stacks); i++ {
		stack := &((*stacks)[i])

		frams := &((*stack).Frames)

		for i := 0; i < len(*frams); i++ {
			frame := &((*frams)[i])
			if uuid, ok := imagesMap[frame.ImageName]; ok == true {
				offset := frame.RetAddress - frame.LoadAddress
				v, err := symbolization.Symbol(offset, uuid)
				if err != nil {
					// not found, add missing dsym record
					missingDsym := MissingDSYM{}
					missingDsym.Init()
					missingDsym.Name = frame.ImageName
					missingDsym.UUID = uuid
					missingDsym.Arch = s.Arch
					missingDsym.AppVersion = s.AppVersion
					missingDsym.AppName = bs.AppImageName
					missingDsym.SystemVersion = s.SysVersion
					missingDsym.SaveToStorage()
				} else {
					frame.RetSymbol = v
					if frame.ImageName == bs.AppImageName {
						bs.IsSymbolized = true
					}
				}
			}
		}
	}
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

func (s *MissingDSYM) Init() {
	s.Id = bson.NewObjectId()
}

func (s *MissingDSYM) SaveToStorage() error {
	if len(s.UUID) <= 0 {
		return errors.New("UUID is Invalid")
	}

	if _, ok := missingDSYMs.Load(s.UUID); ok == true {
		return nil
	}

	session := getSession()
	defer session.Close()

	c := session.DB(dataBase).C(dsymCollection)

	err := c.Insert(s)

	if err == nil {
		missingDSYMs.Store(s.UUID, *s)
	}

	return err
}

func (s *App) Init() {
	s.Id = bson.NewObjectId()
}

func (s *App) SaveToStorage() error {
	if len(s.AppIdentifier) <= 0 {
		return errors.New("AppIdentifier is Invalid")
	}

	if len(s.AppName) <= 0 {
		return errors.New("AppName is Invalid")
	}

	if _, ok := missingDSYMs.Load(s.AppIdentifier); ok == true {
		return nil
	}

	session := getSession()
	defer session.Close()

	c := session.DB(dataBase).C(appCollection)

	err := c.Insert(s)

	if err == nil {
		apps.Store(s.AppIdentifier, *s)
	}

	return err
}

func GetReportsOfIssue(issueId string) *[]string {
	session := getSession()
	defer session.Close()

	var results []Issue

	c := session.DB(dataBase).C(reportCollection)
	err := c.Find(bson.M{"issue": bson.ObjectIdHex(issueId)}).Sort("-timestamp").All(&results)
	if err != nil {
		logger.Log.Error("find reports of issue failed. ", err)
	}

	var resultIds [] string
	for _, result := range results {
		resultIds = append(resultIds, result.IssueId.Hex())
	}
	return &resultIds
}

func GetReportOfId (reportId string) AnrReport {
	session := getSession()
	defer session.Close()

	var result AnrReport

	c := session.DB(dataBase).C(reportCollection)
	err := c.FindId(bson.ObjectIdHex(reportId)).One(&result)
	if err != nil {
		logger.Log.Error("find report failed. ", err)
	}
	return result
}

func InitCacheData () {
	initApp()
	initIssues()
	initMissingDsym()
}

func initMissingDsym () {
	session := getSession()
	defer session.Close()

	var results []MissingDSYM

	c := session.DB(dataBase).C(dsymCollection)
	err := c.Find(nil).All(&results)
	if err != nil {
		logger.Log.Error("find missing dsyms failed. ", err)
	}

	for _, result := range results {
		missingDSYMs.Store(result.UUID, result)
	}
}

func RemoveMissingDSYMS (uuid string) (*MissingDSYM, bool) {
	if i, ok := missingDSYMs.Load(uuid); ok == true {
		session := getSession()
		defer session.Close()

		c := session.DB(dataBase).C(dsymCollection)
		err := c.Remove(bson.M{"uuid": uuid})
		if err != nil {
			logger.Log.Error("remove missing dsym failed. ", err)
		}
		missingDSYMs.Delete(uuid)
		dsym := i.(MissingDSYM)
		return &dsym, true
	} else {
		return nil, false
	}
}

func GetAllMissingDSYMs () *[]MissingDSYM {
	results := new([]MissingDSYM)

	missingDSYMs.Range(func(k, v interface{}) bool {
		*results = append(*results, v.(MissingDSYM))
		return true
	})

	return results
}

func initApp () {
	session := getSession()
	defer session.Close()

	var results []App

	c := session.DB(dataBase).C(appCollection)
	err := c.Find(nil).All(&results)
	if err != nil {
		logger.Log.Error("find app failed. ", err)
	}

	for _, result := range results {
		apps.Store(result.AppIdentifier, result)
	}
}

func GetAllApps () *[]App {
	results := new([]App)

	apps.Range(func(k, v interface{}) bool {
		*results = append(*results, v.(App))
		return true
	})

	return results
}

func initIssues () {
	session := getSession()
	defer session.Close()

	var results []Issue

	c := session.DB(dataBase).C(issueCollection)
	err := c.Find(nil).All(&results)
	if err != nil {
		logger.Log.Error("find issues failed. ", err)
	}

	for _, result := range results {
		addIssueToCache(&result)
	}
}

func addIssueToCache(issue *Issue) {
	var issueMap sync.Map
	appId := issue.AppId.Hex()
	if v, ok := issues.Load(appId); ok == true {
		issueMap = v.(sync.Map)
	} else {
		issueMap = sync.Map{}
		issues.Store(appId, issueMap)
	}
	issueMap.Store(issue.IssueIdentifier, *issue)
}

func getIssue(appId string, issueIdentifier string) (*Issue, bool) {
	if v, ok := issues.Load(appId); ok == true {
		issueMap := v.(sync.Map)
		if v2, ok = issueMap.Load(issueIdentifier); ok == true {
			issue := v2.(Issue)
			return &issue, ok
		}
	}
	return nil, false
}

func (s *Issue) SaveToStorage(inSession *mgo.Session) error {
	session := inSession
	if session == nil {
		session = getSession()
		defer session.Close()
	}

	ic := session.DB(dataBase).C(issueCollection)

	var err error

	if err = ic.Insert(s); err == nil {
		addIssueToCache(s)
	}

	return err
}

func GetAllIssues(start int, pageSize int, appId string) (int, *[]Issue) {
	session := getSession()
	defer session.Close()

	var results []Issue

	c := session.DB(dataBase).C(issueCollection)
	err := c.Find(bson.M{"appid": bson.ObjectIdHex(appId)}).Sort("-issuecount").Skip(start).Limit(pageSize).All(&results)
	count, err :=  c.Count()
	if err != nil {
		logger.Log.Error("find all issue failed. ", err)
	}

	return count, &results
}