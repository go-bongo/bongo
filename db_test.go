package frat

import (
	// "testing"
	. "gopkg.in/check.v1"
	"labix.org/v2/mgo/bson"
)


// var key = []byte("asdf1234asdf1234")


// Hook up gocheck into the "go test" runner.
// func Test(t *testing.T) { TestingT(t) }

// type TestSuite struct{}

var _ = Suite(&TestSuite{})


type FooBar struct {
	Id    bson.ObjectId   `bson:"_id"`
	Msg   string        `encrypted:"true",bson="msg"`
	Count int           `encrypted:"true",bson="count"`
}


func (s *TestSuite) TestConnect(c *C) {
	config := &MongoConfig{"localhost","gotest"}

	connection := new(MongoConnection)

	connection.Config = config

	connection.Connect()
	defer connection.Session.Close()

	err := connection.Session.Ping()

	c.Assert(err, Equals, nil)
}

func (s *TestSuite) TestSave(c *C) {
	config := &MongoConfig{"localhost","gotest"}

	connection := new(MongoConnection)

	connection.Config = config

	connection.Connect()

	defer connection.Session.Close()

	// This needs to always be a pointer, otherwise the encryption component won't like it.
	message := new(FooBar)
	message.Msg = "Foo"
	message.Count = 5


	err := connection.Save(message)

	c.Assert(err, Equals, nil)
}



