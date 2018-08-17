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
	"fmt"
)

var (
	mgoSession *mgo.Session
	hosts             = 	[]string {"127.0.0.1:27017"}
	dataBase          =  	"database"
	reportCollection  =  	"anrReport"
	issueCollection   =   	"issue"
	dsymCollection    =     "dsym"
	appCollection     =     "app"
	missingDSYMs      =     Cache{make(map[string] interface{}), sync.RWMutex{}}
	issues            =     Cache{make(map[string] interface{}), sync.RWMutex{}}
	apps 			  =     Cache{make(map[string] interface{}), sync.RWMutex{}}
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
		AppIdentifier			string
	}

	MissingDSYM struct {
		Id				bson.ObjectId `bson:"_id"`
		Name 			string
		UUID			string
		Arch			string
		AppName			string
		AppIdentifier   string
		AppVersion 		string
		SystemVersion	string
	}

	App struct {
		Id				bson.ObjectId `bson:"_id"`
		AppName			string
		AppIdentifier   string
	}

	Cache struct {
		data   map[string] interface{}
		mutex  sync.RWMutex
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

	app := App{}
	app.Init()
	app.AppName = s.Backtrace.AppImageName
	app.AppIdentifier = s.AppId
	app.SaveToStorage()

	issue := Issue{}
	issue.IssueIdentifier = *identifier
	issue.IssueSourceFile = *sourceFile
	issue.IssueCreateTime = float64(time.Now().Unix() * 1000)
	issue.IssueCount = 1
	issue.IssueAffectVersionStart = s.AppVersion
	issue.IssueAffectVersionEnd = s.AppVersion
	issue.IssueLastUpdateTime = issue.IssueCreateTime
	issue.AppIdentifier = s.AppId
	updateIssueErr := issue.SaveToStorage(session)

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
					missingDsym.AppIdentifier = s.AppId
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

func (s *Issue) SaveToStorage(inSession *mgo.Session) error {
	session := inSession
	if session == nil {
		session = getSession()
		defer session.Close()
	}

	issues.mutex.Lock()
	defer issues.mutex.Unlock()

	ic := session.DB(dataBase).C(issueCollection)

	var updateErr error

	if oldIssue, ok := getIssue(s.AppIdentifier, s.IssueIdentifier, true); ok == true {
		change := bson.M{"issuelastupdatetime": float64(time.Now().Unix() * 1000),
			"issuecount": oldIssue.IssueCount + 1}
		s.IssueCount = oldIssue.IssueCount + 1
		s.IssueId = oldIssue.IssueId

		if s.IssueAffectVersionStart < oldIssue.IssueAffectVersionStart {
			change["issueaffectversionstart"] = s.IssueAffectVersionStart
		} else {
			s.IssueAffectVersionStart = oldIssue.IssueAffectVersionStart
		}

		if s.IssueAffectVersionEnd > oldIssue.IssueAffectVersionEnd {
			change["issueaffectversionend"] = s.IssueAffectVersionEnd
		} else {
			s.IssueAffectVersionEnd = oldIssue.IssueAffectVersionEnd
		}
		updateErr = ic.UpdateId(oldIssue.IssueId, bson.M{"$set": change})
	} else {
		s.Init()
		updateErr = ic.Insert(s)
	}

	if updateErr == nil {
		addIssueToCache(s, true)
	}

	return updateErr
}

func (s *MissingDSYM) Init() {
	s.Id = bson.NewObjectId()
}

func (s *MissingDSYM) SaveToStorage() error {
	if len(s.UUID) <= 0 {
		return errors.New("UUID is Invalid")
	}

	missingDSYMs.mutex.Lock()
	defer missingDSYMs.mutex.Unlock()

	if _, ok := missingDSYMs.data[s.UUID]; ok == true {
		return nil
	}

	session := getSession()
	defer session.Close()

	c := session.DB(dataBase).C(dsymCollection)

	err := c.Insert(s)

	if err == nil {
		missingDSYMs.data[s.UUID] = *s
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

	apps.mutex.Lock()
	defer apps.mutex.Unlock()

	if _, ok := apps.data[s.AppIdentifier]; ok == true {
		return nil
	}

	session := getSession()
	defer session.Close()

	c := session.DB(dataBase).C(appCollection)

	err := c.Insert(s)

	if err == nil {
		apps.data[s.AppIdentifier] = *s
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
		missingDSYMs.data[result.UUID] = result
	}
}

func RemoveMissingDSYMS (uuid string) (*MissingDSYM, bool) {
	missingDSYMs.mutex.Lock()
	defer missingDSYMs.mutex.Unlock()

	if i, ok := missingDSYMs.data[uuid]; ok == true {
		session := getSession()
		defer session.Close()

		c := session.DB(dataBase).C(dsymCollection)
		err := c.Remove(bson.M{"uuid": uuid})
		if err != nil {
			logger.Log.Error("remove missing dsym failed. ", err)
			return nil, false
		}
		delete(missingDSYMs.data, uuid)
		dsym := i.(MissingDSYM)
		return &dsym, true
	} else {
		return nil, false
	}
}

func GetAllMissingDSYMs (appId string) *[]MissingDSYM {
	if len(appId) <= 0 {
		return nil
	}

	results := new([]MissingDSYM)

	missingDSYMs.mutex.RLock()
	defer missingDSYMs.mutex.RUnlock()

	for _, v := range missingDSYMs.data {
		dsym := v.(MissingDSYM)
		if dsym.AppIdentifier == appId {
			*results = append(*results, dsym)
		}
	}

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
		apps.data[result.AppIdentifier] = result
	}
}

func GetAllApps () *[]App {
	results := new([]App)

	apps.mutex.RLock()
	defer apps.mutex.RUnlock()

	for _, app := range apps.data {
		*results = append(*results, app.(App))
	}

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
		addIssueToCache(&result, true)
	}
}

func addIssueToCache(issue *Issue, lockFree bool) {
	var issueMap map[string] Issue
	if lockFree == false {
		issues.mutex.Lock()
		defer issues.mutex.Unlock()
	}

	if v, ok := issues.data[issue.AppIdentifier]; ok == true {
		issueMap = v.(map[string] Issue)
	} else {
		issueMap = make(map[string] Issue)
		issues.data[issue.AppIdentifier] = issueMap
	}
	issueMap[issue.IssueIdentifier] = *issue
}

func getIssue(appIdentifier string, issueIdentifier string, lockFree bool) (*Issue, bool) {
	if lockFree == false {
		issues.mutex.RLock()
		defer issues.mutex.RUnlock()
	}

	if v, ok := issues.data[appIdentifier]; ok == true {
		issueMap := v.(map[string] Issue)
		i, ok := issueMap[issueIdentifier]
		if ok == true {
			return &i, ok
		}
	}
	return nil, false
}

func GetAllIssues(start int, pageSize int, appId string) (int, *[]Issue) {
	results := new([]Issue)

	session := getSession()
	defer session.Close()

	c := session.DB(dataBase).C(issueCollection)
	if err := c.Find(bson.M{"appidentifier": appId}).Sort("-issuecount").Skip(start).Limit(pageSize).All(results); err != nil {
		logger.Log.Error(fmt.Sprintf("find all issue for app %s failed. ", appId), err)
	}

	issues.mutex.RLock()
	defer issues.mutex.RUnlock()

	count := 0

	if v, ok := issues.data[appId]; ok == true {
		issueMap := v.(map[string] Issue)
		count = len(issueMap)
	} else {
		logger.Log.Error(fmt.Sprintf("issues cache for app %s cannot find.", appId))
	}

	return count, results
}