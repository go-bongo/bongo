package frat

import (
	"testing"
	. "gopkg.in/check.v1"
	// "labix.org/v2/mgo/bson"
)



// Hook up gocheck into the "go test" runner.
func DocumentTest(t *testing.T) { TestingT(t) }

type DocumentTestSuite struct{}

var _ = Suite(&DocumentTestSuite{})

type msg struct {
	Msg   string        `encrypted:"true",bson="msg"`
	Count int           `encrypted:"true",bson="count"`
}


func (s *DocumentTestSuite) TestDocument(c *C) {
	config := &MongoConfig{"localhost","gotest"}

	connection := new(MongoConnection)

	connection.Config = config

	connection.Connect()

	defer connection.Session.Close()

	// This needs to always be a pointer, otherwise the encryption component won't like it.
	message := new(msg)
	message.Msg = "Foo"
	message.Count = 5

	myDoc := &Document{
		Collection:connection.Collection("message"),
		Connection:connection,
		Model:message,
	}

	err := myDoc.Save()

	c.Assert(err, Equals, nil)

	// Now find it by ID and make sure it has all the right values
	

	// var newMessage msg
	// err = FindById(doc.Id, &newMessage, connection.Collection("message"))
	// if err != nil {
	// 	panic(err)
	// }


	// fmt.Println(newMessage)

	// c.Assert(newMessage.Msg, Equals, "Foo")
	// c.Assert(newMessage.Count, Equals, 5)
	
	
	// err = connection.Session.DB(connection.Config.Database).DropDatabase()

	// if err != nil {
	// 	panic(err)
	// }

}

