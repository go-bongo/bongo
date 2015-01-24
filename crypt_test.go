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
	myStruct := &EncryptedStruct{true, "foo", 5.555, 6, EncryptedDate(date)}

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

}
