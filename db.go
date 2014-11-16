package frat

import (
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"github.com/oleiade/reflections"
	"reflect"
	// "strings"
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

func (m *MongoConnection) Collection(name string) *mgo.Collection {
	return m.Session.DB(m.Config.Database).C(name)
}

func getCollectionName(mod interface{}) (string) {
	return ToSnake(reflect.Indirect(reflect.ValueOf(mod)).Type().Name())
}

func ensureIdField(mod interface{}) {
	has, _ := reflections.HasField(mod, "Id")
	if !has {
		panic("Failed to save - model must have Id field")
	}
}


func (c *MongoConnection) Save(mod interface{}) (error) {


	// 1) Make sure mod has an Id field
	ensureIdField(mod)

	// 2) If there's no ID, create a new one
	f, err := reflections.GetField(mod, "Id")
	id := f.(bson.ObjectId)

	if err != nil {
		panic(err)
	}
	
	if !id.Valid() {
		id := bson.NewObjectId()
		err := reflections.SetField(mod, "Id", id)  // err != nil

		if err != nil {
			panic(err)
		}
	}
	
	// 2) Convert the model into a map using the crypt library
	modelMap := EncryptDocument(key, mod)
	err =  c.Collection(getCollectionName(mod)).Insert(modelMap)

	return nil
}

func (c *MongoConnection) FindById(id bson.ObjectId, mod interface{}) (error) {
	returnMap := make(map[string]interface{})

	err := c.Collection(getCollectionName(mod)).FindId(id).One(&returnMap)
	if err != nil {
		return err
	}

	// Decrypt
	DecryptDocument(key, returnMap, mod)

	return nil
}



