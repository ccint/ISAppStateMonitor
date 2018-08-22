package store

import (
	"time"
	"github.com/globalsign/mgo"
)

var (
	mgoSession *mgo.Session
	hosts             = 	[]string {"127.0.0.1:27017"}
)

func GetSession() *mgo.Session {
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