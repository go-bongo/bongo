package bongo

import (
	// "encoding/json"
	// "fmt"
	. "gopkg.in/check.v1"
	// "github.com/maxwellhealth/mgo/bsoncrypt"
	// "testing"
	// "log"
	"time"
)

type EncryptedStruct struct {
	Bool   EncryptedBool
	String EncryptedString
	Float  EncryptedFloat
	Int    EncryptedInt
	Date   EncryptedDate
	Map    EncryptedMap
}

func (s *TestSuite) TestRawCryptDecrypt(c *C) {
	val := "my string"

	encrypted, err := Encrypt(key, []byte(val))

	c.Assert(err, Equals, nil)

	decrypted, err := Decrypt(key, encrypted)

	c.Assert(err, Equals, nil)
	c.Assert(string(decrypted), Equals, "my string")
}

func (s *TestSuite) TestEncryptedTypes(c *C) {
	date := time.Now()

	mp := map[string]interface{}{
		"foo":  "bar",
		"baz":  5,
		"boop": 10.5,
		"childMap": map[string]interface{}{
			"bing": "boop",
			"boof": "foop",
		},
	}
	myStruct := &EncryptedStruct{true, "foo", 5.555, 6, EncryptedDate(date), mp}

	// connection.Session.EncryptionKey = key
	connection.Collection("tests").Collection().Insert(myStruct)

	newStruct := &EncryptedStruct{}

	err := connection.Collection("tests").Collection().Find(nil).One(newStruct)

	c.Assert(err, Equals, nil)
	c.Assert(bool(newStruct.Bool), Equals, true)
	c.Assert(string(newStruct.String), Equals, "foo")
	c.Assert(float64(newStruct.Float), Equals, 5.555)
	c.Assert(int(newStruct.Int), Equals, 6)
	c.Assert(time.Time(newStruct.Date), Equals, date)

	newMp := map[string]interface{}(newStruct.Map)
	c.Assert(newMp["foo"], Equals, "bar")

}

// func (s *TestSuite) TestEncryptedTypesWithNoEncryption(c *C) {
// 	EncryptionKey = []byte{}
// 	date := time.Now()
// 	myStruct := &EncryptedStruct{true, "foo", 5.555, 6, EncryptedDate(date)}

// 	// connection.Session.EncryptionKey = key
// 	connection.Collection("tests").Collection().Insert(myStruct)

// 	newStruct := &EncryptedStruct{}

// 	err := connection.Collection("tests").Collection().Find(nil).One(newStruct)

// 	c.Assert(err, Equals, nil)
// 	c.Assert(bool(newStruct.Bool), Equals, true)
// 	c.Assert(string(newStruct.String), Equals, "foo")
// 	c.Assert(float64(newStruct.Float), Equals, 5.555)
// 	c.Assert(int(newStruct.Int), Equals, 6)

// 	// BSON loses some nanosecond precision on raw marshal/unmarshal
// 	c.Assert(time.Time(newStruct.Date).Format(time.RFC1123Z), Equals, date.Format(time.RFC1123Z))
// }
