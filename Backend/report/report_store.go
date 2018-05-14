package report

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const URL = "127.0.0.1:27017"

var (
	mgoSession *mgo.Session
	dataBase = "database"
	collection = "report"
)

func getSession() *mgo.Session {
	if mgoSession == nil {
		var err error
		mgoSession, err = mgo.Dial(URL)
		if err != nil {
			panic(err)
		}
	}
	return mgoSession.Clone()
}

type Report struct {
	Time uint64
	Content string
}


func Mongotest() {
	session := getSession()
	defer session.Close()
	c := session.DB(dataBase).C(collection)

	//err := c.Insert(&Report{1552314513, "Test String Test String"})
	//if err != nil {
	//	println(err)
	//}

	result := Report{}

	err := c.Find(bson.M{"time": 1552314513}).One(&result)
	if err != nil {
		println(err)
	}

	println("cotnent", result.Content)
}