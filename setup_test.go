package bongo

import (
	. "gopkg.in/check.v1"
	"labix.org/v2/mgo/bson"
	"testing"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { TestingT(t) }

type TestSuite struct{}

var _ = Suite(&TestSuite{})

var key = []byte("asdf1234asdf1234")

type Nested struct {
	Foo     string
	BazBing string `bson:"bazBing"`
}

type FooBar struct {
	Id    bson.ObjectId `bson:"_id"`
	Msg   string        `encrypted:"true" bson:"msg"`
	Count int           `encrypted:"false" bson:"count" index:"true"`
	Child *Nested
}

func (f *FooBar) Validate() []string {
	errs := []string{}
	if f.Count == 3 {
		errs = append(errs, "count cannot be 3")
	}

	return errs
}

// Add some hooks
func (f *FooBar) BeforeSave() {
	f.Count++
}

func (f *FooBar) BeforeCreate() {
	f.Count++
}

func (f *FooBar) BeforeUpdate() {
	f.Count = f.Count + 2
}

func (f *FooBar) AfterFind() {
	f.Count = f.Count + 5
}

var config = &Config{
	ConnectionString: "localhost",
	Database:         "gotest",
	EncryptionKey:    "asdf1234asdf1234",
}

var connection = Connect(config)

func (s *TestSuite) TearDownTest(c *C) {
	connection.Session.DB(config.Database).DropDatabase()
}

func (s *TestSuite) TearDownSuite(c *C) {
	connection.Session.Close()
}
