package frat

import (
	. "gopkg.in/check.v1"
	"labix.org/v2/mgo/bson"
)






func (s *TestSuite) TestValidateRequired(c *C) {
	c.Assert(ValidateRequired("foo"), Equals, true)
	c.Assert(ValidateRequired(""), Equals, false)
	c.Assert(ValidateRequired(0), Equals, false)
	c.Assert(ValidateRequired(1), Equals, true)
}

func (s *TestSuite) TestValidateInclusionIn(c *C) {
	c.Assert(ValidateInclusionIn("foo", []string{"foo","bar","baz"}), Equals, true)
	c.Assert(ValidateInclusionIn("bing", []string{"foo","bar","baz"}), Equals, false)
}

func (s *TestSuite) TestValidateMongoIdRef(c *C) {
	// Make the doc
	connection := Connect(config)
	defer connection.Session.Close()

	message := new(FooBar)
	message.Msg = "Foo"
	message.Count = 5


	err, _ := connection.Save(message)

	c.Assert(err, Equals, nil)
	c.Assert(ValidateMongoIdRef(message.Id, connection.Collection("foo_bar")), Equals, true)
	c.Assert(ValidateMongoIdRef(bson.NewObjectId(), connection.Collection("foo_bar")), Equals, false)
	c.Assert(ValidateMongoIdRef(bson.NewObjectId(), connection.Collection("other_collection")), Equals, false)
}