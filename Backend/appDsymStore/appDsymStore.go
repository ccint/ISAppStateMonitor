package appDsymStore

import (
	"../store"
	"fmt"
	"time"
	"github.com/globalsign/mgo/bson"
)

var (
	appDsymCollection =     "appdsym"
	dataBase          =  	"database"
)

type (
	AppDsym struct {
		Id				bson.ObjectId `bson:"_id"`
		Date            int64
		UUID            string
		AppName         string
		IsDebug         bool
	}
)

func (s *AppDsym) Init () {
	s.Id = bson.NewObjectId()
}

func (s *AppDsym) saveStorage () {
	session := store.GetSession()
	defer session.Close()

	c := session.DB(dataBase).C(appDsymCollection)

	if err := c.Find(bson.M{"uuid": s.UUID}).One(nil); err != nil {
		c.Insert(s)
	} else {
		fmt.Println("exist")
	}
}

func AddAppDsymRecord (appName string, uuid string, isDebug bool) {
	newAppDsym := AppDsym{}
	newAppDsym.Init()
	newAppDsym.AppName = appName
	newAppDsym.UUID = uuid
	newAppDsym.IsDebug = isDebug
	newAppDsym.Date = time.Now().Unix()
	newAppDsym.saveStorage()
}

func IsAppDsymRecordExist (uuid string) bool {
	if len(uuid) <= 0 {
		return false
	}

	session := store.GetSession()
	defer session.Close()

	c := session.DB(dataBase).C(appDsymCollection)

	if err := c.Find(bson.M{"uuid": uuid}).One(nil); err != nil {
		return false
	} else {
		return true
	}
}