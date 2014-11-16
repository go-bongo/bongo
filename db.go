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
	EncryptionKey string
	EncryptionKeyPerCollection map[string]string
}

type MongoConnection struct {
	Config *MongoConfig
	Session *mgo.Session
}

// Create a new connection and run Connect()
func Connect(config *MongoConfig) *MongoConnection {
	conn := &MongoConnection{
		Config:config,
	}

	conn.Connect()

	return conn
}

// Connect to the database using the provided config
func (m *MongoConnection) Connect() {
	session, err := mgo.Dial(m.Config.ConnectionString)

	if err != nil {
		panic(err)
	}

	m.Session = session
}

// Convenience for retrieving a collection by name based on the config passed to the MongoConnection
func (m *MongoConnection) Collection(name string) *mgo.Collection {
	return m.Session.DB(m.Config.Database).C(name)
}

func (m *MongoConnection) GetEncryptionKey(collection string) []byte {
	key, has := m.Config.EncryptionKeyPerCollection[collection]

	if has {
		return []byte(key)
	} else {
		return []byte(m.Config.EncryptionKey)
	}

}

// Get the collection name from an arbitrary interface. Returns type name in snake case
func getCollectionName(mod interface{}) (string) {
	return ToSnake(reflect.Indirect(reflect.ValueOf(mod)).Type().Name())
}

// Ensure that a particular interface has an "Id" field. Panic if not
func ensureIdField(mod interface{}) {
	has, _ := reflections.HasField(mod, "Id")
	if !has {
		panic("Failed to save - model must have Id field")
	}
}


// Save a document. Collection name is interpreted from name of struct
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
	
	colname := getCollectionName(mod)
	// 3) Convert the model into a map using the crypt library
	modelMap := EncryptDocument(c.GetEncryptionKey(colname), mod)
	err =  c.Collection(colname).Insert(modelMap)

	return nil
}

// Find a document by ID. Collection name is interpreted from name of struct
func (c *MongoConnection) FindById(id bson.ObjectId, mod interface{}) (error) {
	returnMap := make(map[string]interface{})

	colname := getCollectionName(mod)
	err := c.Collection(colname).FindId(id).One(&returnMap)
	if err != nil {
		return err
	}

	// Decrypt
	
	DecryptDocument(c.GetEncryptionKey(colname), returnMap, mod)

	return nil
}

// Delete a document. Collection name is interpreted from name of struct
func (c *MongoConnection) Delete(mod interface{}) (error) {
	ensureIdField(mod)
	f, err := reflections.GetField(mod, "Id")
	if err != nil {
		return err
	}
	id := f.(bson.ObjectId)
	colname := getCollectionName(mod)

	return c.Collection(colname).Remove(bson.M{"_id": id})
}



