package db

import (
	"labix.org/v2/mgo"
	// "labix.org/v2/mgo/bson"
)



type MongoConfig struct {
	ConnectionString string
	Database string
}


type MongoConnection struct {
	Config *MongoConfig
	Session *mgo.Session
}

func (m *MongoConnection) Connect() {
	session, err := mgo.Dial(m.Config.ConnectionString)

	if err != nil {
		panic(err)
	}

	m.Session = session
}

func (m *MongoConnection) Collection(name string) *Collection {
	return &Collection{m.Session.DB(m.Config.Database).C(name)}
}