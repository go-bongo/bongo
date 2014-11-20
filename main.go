package bongo

import (
	// "fmt"
	"github.com/oleiade/reflections"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"reflect"
	// "fmt"
	// "math"
	// "strings"
)

type Config struct {
	ConnectionString           string
	Database                   string
	EncryptionKey              string
	EncryptionKeyPerCollection map[string]string
}

type Connection struct {
	Config  *Config
	Session *mgo.Session
}

// Create a new connection and run Connect()
func Connect(config *Config) *Connection {
	conn := &Connection{
		Config: config,
	}

	conn.Connect()

	return conn
}

// Connect to the database using the provided config
func (m *Connection) Connect() {
	session, err := mgo.Dial(m.Config.ConnectionString)

	if err != nil {
		panic(err)
	}

	m.Session = session
}

func (m *Connection) GetEncryptionKey(collection string) []byte {
	key, has := m.Config.EncryptionKeyPerCollection[collection]

	if has {
		return []byte(key)
	} else {
		return []byte(m.Config.EncryptionKey)
	}

}

func (m *Connection) Collection(name string) *Collection {
	// Just create a new instance - it's cheap and only has name
	return &Collection{
		Connection: m,
		Name:       name,
	}
}

// Get the collection name from an arbitrary interface. Returns type name in snake case
func getCollectionName(mod interface{}) string {
	return ToSnake(reflect.Indirect(reflect.ValueOf(mod)).Type().Name())
}

// Ensure that a particular interface has an "Id" field. Panic if not
func ensureIdField(mod interface{}) {
	has, _ := reflections.HasField(mod, "Id")
	if !has {
		panic("Failed to save - model must have Id field")
	}
}

// Wrappers for the collection-level methods to avoid extra typing if you want the collection name to be interpreted from the struct
func (m *Connection) FindById(id bson.ObjectId, mod interface{}) error {
	col := m.Collection(getCollectionName(mod))

	return col.FindById(id, mod)
}

func (m *Connection) Save(mod interface{}) (error, []string) {
	col := m.Collection(getCollectionName(mod))

	return col.Save(mod)
}

func (m *Connection) Delete(mod interface{}) error {
	col := m.Collection(getCollectionName(mod))

	return col.Delete(mod)
}
