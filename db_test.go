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

func (s *TestSuite) TestSaveAndFind(c *C) {
	config := &MongoConfig{"localhost","gotest"}

	connection := Connect(config)


	defer connection.Session.Close()

	// This needs to always be a pointer, otherwise the encryption component won't like it.
	message := new(FooBar)
	message.Msg = "Foo"
	message.Count = 5


	err := connection.Save(message)

	c.Assert(err, Equals, nil)

	newMessage := new(FooBar)

	connection.FindById(message.Id, newMessage)

	// Make sure the ids are the same
	c.Assert(newMessage.Id.String(), Equals, message.Id.String())
	c.Assert(newMessage.Msg, Equals, message.Msg)
	c.Assert(newMessage.Count, Equals, message.Count)
}

func (s *TestSuite) TestFindNonExistent(c *C) {
	config := &MongoConfig{"localhost","gotest"}

	connection := Connect(config)


	defer connection.Session.Close()

	newMessage := new(FooBar)

	err := connection.FindById(bson.NewObjectId(), newMessage)

	c.Assert(err.Error(), Equals, "not found")
}

func (s *TestSuite) TestDelete(c *C) {
	config := &MongoConfig{"localhost","gotest"}

	connection := Connect(config)


	defer connection.Session.Close()

	// This needs to always be a pointer, otherwise the encryption component won't like it.
	message := new(FooBar)
	message.Msg = "Foo"
	message.Count = 5


	err := connection.Save(message)

	c.Assert(err, Equals, nil)

	connection.Delete(message)

	newMessage := new(FooBar)
	err = connection.FindById(message.Id, newMessage)
	c.Assert(err.Error(), Equals, "not found")
	// Make sure the ids are the same

}
