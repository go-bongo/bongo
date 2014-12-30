package bongo

import (
	"encoding/json"
	"errors"
	// "fmt"
	. "gopkg.in/check.v1"
)

func (s *TestSuite) TestConnect(c *C) {

	connection := new(Connection)

	connection.Config = config

	connection.Connect()
	defer connection.Session.Close()

	err := connection.Session.Ping()

	c.Assert(err, Equals, nil)

	connection.Session.DB(config.Database).DropDatabase()
}

func (s *TestSuite) TestRegister(c *C) {
	connection := new(Connection)

	connection.Config = config

	connection.Connect()

	connection.Register(&FooBar{}, "foo_bar")

	indexes, err := connection.Collection("foo_bar").Collection().Indexes()

	c.Assert(err, Equals, nil)

	c.Assert(len(indexes), Equals, 2)
	c.Assert(indexes[0].Key[0], Equals, "_id")
	c.Assert(indexes[1].Key[0], Equals, "count")

}

type errorHolder struct {
	Err *SaveResult
}

func (s *TestSuite) TestMarshalResult(c *C) {

	err := NewSaveResult(false, errors.New("Failed to save"))

	holder := &errorHolder{
		Err: err,
	}

	marshaled, e := json.Marshal(holder)
	c.Assert(e, IsNil)

	c.Assert(string(marshaled), Equals, `{"Err":"Failed to save"}`)

	holder.Err.ValidationErrors = []string{"foo", "bar"}

	marshaled, e = json.Marshal(holder)
	c.Assert(e, IsNil)

	c.Assert(string(marshaled), Equals, `{"Err":["foo","bar"]}`)
	// fmt.Println(string(marshaled))

}
