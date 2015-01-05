package bongo

import (
	// "fmt"
	"encoding/json"
	"github.com/oleiade/reflections"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"log"
	"reflect"
	"strings"
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

type SaveResult struct {
	Success          bool
	Err              error
	ValidationErrors []string
}

func (s *SaveResult) Error() string {
	return s.Err.Error()
}

func NewSaveResult(success bool, err error) *SaveResult {
	return &SaveResult{success, err, []string{}}
}

func (s *SaveResult) MarshalJSON() ([]byte, error) {
	// If there are validation errors, just return those as an array. Otherwise marshal the error.Error() string
	if len(s.ValidationErrors) > 0 {
		return json.Marshal(s.ValidationErrors)
	} else {
		return json.Marshal(s.Err.Error())
	}
}

type Connection struct {
	Config  *Config
	Session *mgo.Session
	// collection []Collection
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

// Register a struct for a certain collection, adding any indeces if necessary. For now this just ensures indeces. Later we might store references to collections, etc
func (m *Connection) Register(mod interface{}, colName string) error {

	structName := getCollectionName(mod)
	if len(colName) == 0 {
		colName = structName
	}

	// if val, ok := m.collectionMap[structName]; ok {
	// 	// Already registered. Just return nil
	// 	return nil
	// }

	collection := m.Collection(colName)

	// Look at any indeces. For now we'll only support top level
	v := reflect.ValueOf(mod)

	var s reflect.Value

	if v.Kind() == reflect.Ptr {
		s = v.Elem()
	} else {
		s = v
	}

	// s := reflect.ValueOf(doc).Elem()
	typeOfT := s.Type()
	for i := 0; i < s.NumField(); i++ {
		// f := s.Field(i)
		fieldName := typeOfT.Field(i).Name

		// encrypt := stringInSlice(fieldName, encryptedFields)
		bongo := typeOfT.Field(i).Tag.Get("bongo")
		tags := getBongoTags(bongo)

		var bsonName string
		bsonName = typeOfT.Field(i).Tag.Get("bson")
		if len(bsonName) == 0 {
			bsonName = strings.ToLower(fieldName)
		}

		if tags.index {
			idx := mgo.Index{
				Key:        []string{bsonName},
				Unique:     false,
				DropDups:   true,
				Background: true, // See notes.
				Sparse:     true,
			}

			if tags.unique {
				idx.Unique = true
			}

			log.Printf("Ensuring index on %s.%s\n", colName, bsonName)

			err := collection.Collection().EnsureIndex(idx)
			if err != nil {
				return err
			}
		}
	}

	return nil

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

// Pass in the sample just so we can get the collection name
func (m *Connection) Find(query interface{}, collection interface{}) *ResultSet {

	// If collection is a string, assume that's the collection name
	var colname string
	if str, ok := collection.(string); ok {
		colname = str
	} else {
		colname = getCollectionName(collection)
	}

	col := m.Collection(colname)

	return col.Find(query)
}

// Wrappers for the collection-level methods to avoid extra typing if you want the collection name to be interpreted from the struct
func (m *Connection) FindById(id bson.ObjectId, mod interface{}) error {
	col := m.Collection(getCollectionName(mod))

	return col.FindById(id, mod)
}

func (m *Connection) FindOne(query interface{}, mod interface{}) error {
	col := m.Collection(getCollectionName(mod))

	return col.FindOne(query, mod)
}

func (m *Connection) Save(mod interface{}) *SaveResult {
	col := m.Collection(getCollectionName(mod))

	return col.Save(mod)
}

func (m *Connection) Delete(mod interface{}) error {
	col := m.Collection(getCollectionName(mod))

	return col.Delete(mod)
}
