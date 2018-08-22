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
	"../store"
	"fmt"
)

var (
	dataBase          =  	"database"
	reportCollection  =  	"anrReport"
	issueCollection   =   	"issue"
	dsymCollection    =     "dsym"
	appCollection     =     "app"
	missingDSYMs      =     Cache{make(map[string] interface{}), sync.RWMutex{}}
	issues            =     Cache{make(map[string] interface{}), sync.RWMutex{}}
	apps 			  =     Cache{make(map[string] interface{}), sync.RWMutex{}}
	unClassfiedCount  =     Cache{make(map[string] interface{}), sync.RWMutex{}}
)

type (
	Frame struct {
		ImageName   string
		ImageUUID   string
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
		Issue       string
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
		AppVersion 		string
		SystemVersion	string
		AppIdentifiers  []string
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

func (s *AnrReport) Init() {
	s.ReportId = bson.NewObjectId()
}

func (s *AnrReport) SaveToStorage() error {
	session := store.GetSession()
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
	session := store.GetSession()
	defer session.Close()
	c := session.DB(dataBase).C(reportCollection)
	return c.UpdateId(s.ReportId, s)
}

func (s *AnrReport) updateOrCreateNewIssue (session *mgo.Session) {
	if s.Backtrace.IsSymbolized == false {
		return
	}

	identifier, sourceFile := s.getIssueIdentifierAndSourceFile()

	if identifier == nil || len(*identifier) <= 0 {
		logger.Log.Error("cannot get identifier of report: ", s.ReportId.Hex())
		return
	}

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
		s.Issue = issue.IssueId.Hex()
	}
}

func (s *AnrReport) tryCreateNewApp () {
	app := App{}
	app.Init()
	app.AppName = s.Backtrace.AppImageName
	app.AppIdentifier = s.AppId
	app.SaveToStorage()
}

func (s *AnrReport) recordUnclassfiedCount () {
	if len(s.Issue) <= 0 {
		increaseUnClassfiedCount(s.AppId)
	}
}

func (s *AnrReport) updateOrCreateNewIssueAndApp(session *mgo.Session) {

	s.tryCreateNewApp()

	s.updateOrCreateNewIssue(session)

	s.recordUnclassfiedCount()
}

func (s *AnrReport) getIssueIdentifierAndSourceFile() (*string, *string) {
	if len(s.Backtrace.Stacks) == 0 || len(s.Backtrace.Stacks[0].Frames) == 0 {
		return nil, nil
	}
	var identifier string
	var sourceFile string

	mainStack := s.Backtrace.Stacks[0]
	for _, frame := range  mainStack.Frames {
		if frame.ImageName == s.Backtrace.AppImageName && !strings.Contains(frame.RetSymbol,"main.m") {
			identifier = frame.RetSymbol
			splits := strings.Split(frame.RetSymbol, "\u0009")
			if len(splits) > 1 {
				sourceFile = splits[len(splits) - 1]
			} else {
				sourceFile = frame.ImageName
			}
			return &identifier, &sourceFile
		}
	}

	identifier = mainStack.Frames[0].RetSymbol
	sourceFile = mainStack.Frames[0].ImageName

	return &identifier, &sourceFile
}

func (s *AnrReport)Symbolicate() {
	bs := &(s.Backtrace)

	stacks := &(bs.Stacks)

	for i := 0; i < len(*stacks); i++ {
		stack := &((*stacks)[i])

		frams := &((*stack).Frames)

		for i := 0; i < len(*frams); i++ {
			frame := &((*frams)[i])
			if len(frame.RetSymbol) > 0 {
				continue
			}

			if len(frame.ImageUUID) > 0 {
				offset := frame.RetAddress - frame.LoadAddress
				v, err := symbolization.Symbol(offset, frame.ImageUUID)
				if err != nil {
					// not found, add missing dsym record
					missingDsym := MissingDSYM{}
					missingDsym.Init()
					missingDsym.Name = frame.ImageName
					missingDsym.UUID = frame.ImageUUID
					missingDsym.Arch = s.Arch
					missingDsym.AppVersion = s.AppVersion
					missingDsym.AppName = bs.AppImageName
					missingDsym.AppIdentifiers = []string{s.AppId}
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
	session := store.GetSession()
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
		session = store.GetSession()
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

	if len(s.AppIdentifiers) <= 0 {
		return errors.New("appidentifier not exist when save missingdsym, just ignore")
	}

	missingDSYMs.mutex.Lock()
	defer missingDSYMs.mutex.Unlock()

	var oldDsym *MissingDSYM

	if v, ok := missingDSYMs.data[s.UUID]; ok == true {
		oldDsym = v.(*MissingDSYM)
		if oldDsym.AppIsIncluded(s.AppIdentifiers[0]) {
			return nil
		} else {
			oldDsym.AppIdentifiers = append(oldDsym.AppIdentifiers, s.AppIdentifiers[0])
		}
	}

	session := store.GetSession()
	defer session.Close()

	c := session.DB(dataBase).C(dsymCollection)

	var err error

	if oldDsym != nil {
		err = c.UpdateId(oldDsym.Id, oldDsym)
	} else {
		if err = c.Insert(s); err == nil {
			missingDSYMs.data[s.UUID] = s
		}
	}

	return err
}

func (s *MissingDSYM) AppIsIncluded (appidentifier string) bool {
	for _, identifier := range s.AppIdentifiers {
		if identifier == appidentifier {
			return true
		}
	}
	return false
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

	session := store.GetSession()
	defer session.Close()

	c := session.DB(dataBase).C(appCollection)

	err := c.Insert(s)

	if err == nil {
		apps.data[s.AppIdentifier] = *s
	}

	return err
}

func GetReportsOfIssue(issueId string) *[]string {
	session := store.GetSession()
	defer session.Close()

	var results []Issue

	c := session.DB(dataBase).C(reportCollection)
	err := c.Find(bson.M{"issue": issueId}).Sort("-timestamp").All(&results)
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
	session := store.GetSession()
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
	initUnClassfiedCount()
}

func initUnClassfiedCount () {
	session := store.GetSession()
	defer session.Close()

	var reports []AnrReport

	c := session.DB(dataBase).C(reportCollection)
	if err := c.Find(bson.M{"issue": ""}).All(&reports); err != nil {
		logger.Log.Error("find all reports failed. ", err)
	}

	for _, report := range reports {
		increaseUnClassfiedCount(report.AppId)
	}
}

func increaseUnClassfiedCount (appid string) error {
	if len(appid) <= 0 {
		return errors.New("appid is nil, when increase unclassfiedCounts")
	}

	unClassfiedCount.mutex.Lock()
	defer unClassfiedCount.mutex.Unlock()

	var count int
	if v, ok := unClassfiedCount.data[appid]; ok == true {
		oldCount := v.(int)
		count = oldCount + 1
	} else {
		count = 1
	}
	unClassfiedCount.data[appid] = count
	return nil
}


func removeAllUnClassfiedCount (appid string) error {
	if len(appid) <= 0 {
		return errors.New("appid is nil, when reduce unclassfiedCounts")
	}

	unClassfiedCount.mutex.Lock()
	defer unClassfiedCount.mutex.Unlock()

	unClassfiedCount.data[appid] = 0
	return nil
}

func reduceUnClassfiedCount (appid string, lockFree bool) error {
	if len(appid) <= 0 {
		return errors.New("appid is nil, when reduce unclassfiedCounts")
	}

	if lockFree == false {
		unClassfiedCount.mutex.Lock()
		defer unClassfiedCount.mutex.Unlock()
	}

	if v, ok := unClassfiedCount.data[appid]; ok == true {
		oldCount := v.(int)
		if oldCount >= 1 {
			unClassfiedCount.data[appid] = oldCount - 1
		}
	}
	return nil
}

func initMissingDsym () {
	session := store.GetSession()
	defer session.Close()

	var results []MissingDSYM

	c := session.DB(dataBase).C(dsymCollection)
	err := c.Find(nil).All(&results)
	if err != nil {
		logger.Log.Error("find missing dsyms failed. ", err)
	}

	for _, result := range results {
		missingDSYMs.data[result.UUID] = &result
	}
}

func RemoveMissingDSYMS (uuid string) (*MissingDSYM, bool) {
	missingDSYMs.mutex.Lock()
	defer missingDSYMs.mutex.Unlock()

	if i, ok := missingDSYMs.data[uuid]; ok == true {
		session := store.GetSession()
		defer session.Close()

		c := session.DB(dataBase).C(dsymCollection)
		err := c.Remove(bson.M{"uuid": uuid})
		if err != nil {
			logger.Log.Error("remove missing dsym failed. ", err)
			return nil, false
		}
		delete(missingDSYMs.data, uuid)
		dsym := i.(*MissingDSYM)
		return dsym, true
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
		dsym := v.(*MissingDSYM)
		if dsym.AppIsIncluded(appId) == true {
			*results = append(*results, *dsym)
		}
	}

	return results
}

func initApp () {
	session := store.GetSession()
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
	session := store.GetSession()
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

func GetAllIssues(start int, pageSize int, appId string) (int, *[]Issue, int) {
	results := new([]Issue)

	session := store.GetSession()
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
	}

	unClassfiedCount.mutex.RLock()
	defer unClassfiedCount.mutex.RUnlock()
	unClassfiedReportsCount := 0
	if v, ok := unClassfiedCount.data[appId]; ok == true {
		unClassfiedReportsCount = v.(int)
	}

	return count, results, unClassfiedReportsCount
}

func SymbolicateUnClassfiedReports (appid string) error {
	if len(appid) <= 0 {
		return errors.New("appid is nil when symbolicate unclassfied reports")
	}

	unClassfiedCount.mutex.Lock()
	defer unClassfiedCount.mutex.Unlock()

	session := store.GetSession()
	defer session.Close()

	var reports []AnrReport

	c := session.DB(dataBase).C(reportCollection)
	if err := c.Find(bson.M{"issue": "", "appid": appid}).All(&reports); err != nil {
		return err
	}

	for _, report := range reports {
		report.Symbolicate()
		report.updateOrCreateNewIssue(session)
		if err := report.UpdateToStorage(); err == nil && len(report.Issue) > 0 {
			reduceUnClassfiedCount(appid, true)
		} else if err != nil {
			logger.Log.Error("update report failed when resymbolicate unclassfied report: ", report.ReportId.Hex())
		}
	}
	return nil
}