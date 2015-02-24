package bongo

import (
	"github.com/maxwellhealth/mgo/bson"
	. "gopkg.in/check.v1"
	"log"
	"testing"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { TestingT(t) }

type TestSuite struct{}

var _ = Suite(&TestSuite{})

type NullWriter int

func (NullWriter) Write([]byte) (int, error) { return 0, nil }

var key = []byte("asdf1234asdf1234")

type Nested struct {
	Foo     string
	BazBing string `bson:"bazBing"`
}

type FooBar struct {
	Id    bson.ObjectId   `bson:"_id"`
	Msg   EncryptedString `bongo:"encrypted" bson:"msg"`
	Count int             `bongo:"index"`
	Child *Nested
}

func (f *FooBar) Validate(c *Collection) []string {
	errs := []string{}
	if f.Count == 3 {
		errs = append(errs, "count cannot be 3")
	}

	return errs
}

// Add some hooks
func (f *FooBar) BeforeSave(c *Collection) {
	f.Count++
}

func (f *FooBar) BeforeCreate(c *Collection) {
	f.Count++
}

func (f *FooBar) BeforeUpdate(c *Collection) {
	f.Count = f.Count + 2
}

func (f *FooBar) AfterFind(c *Collection) {
	f.Count = f.Count + 5
}

var config = &Config{
	ConnectionString: "localhost",
	Database:         "gotest",
}

var connection, _ = Connect(config)

func (s *TestSuite) SetUpTest(c *C) {
	EncryptionKey = []byte("asdf1234asdf1234")
	connection.Session.DB(config.Database).DropDatabase()
	if !testing.Verbose() {
		log.SetOutput(new(NullWriter))
	}

}

// func (s *TestSuite) TearDownSuite(c *C) {
// 	connection.Session.Close()
// }
