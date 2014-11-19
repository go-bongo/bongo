package bongo

import (
	. "gopkg.in/check.v1"
	"labix.org/v2/mgo/bson"
)

func (s *TestSuite) TestFind(c *C) {

	connection := Connect(config)

	defer connection.Session.Close()

	// This needs to always be a pointer, otherwise the encryption component won't like it.
	message := new(FooBar)
	message.Msg = "Foo"
	message.Count = 5

	err, _ := connection.Save(message)

	c.Assert(err, Equals, nil)

	message2 := new(FooBar)
	message2.Msg = "Bar"
	message2.Count = 10

	err, _ = connection.Save(message2)

	c.Assert(err, Equals, nil)

	// Now run a find
	results := connection.Find(nil, &FooBar{})

	res := new(FooBar)

	count := 0

	for results.Next(res) {
		count++
		if count == 1 {
			c.Assert(res.Msg, Equals, "Foo")
		} else {
			c.Assert(res.Msg, Equals, "Bar")
		}
	}

	c.Assert(count, Equals, 2)

	connection.Session.DB(config.Database).DropDatabase()
}

func (s *TestSuite) TestFindWithPagination(c *C) {

	connection := Connect(config)

	defer connection.Session.Close()

	// This needs to always be a pointer, otherwise the encryption component won't like it.
	message := new(FooBar)
	message.Msg = "Foo"
	message.Count = 5

	err, _ := connection.Save(message)

	c.Assert(err, Equals, nil)

	message2 := new(FooBar)
	message2.Msg = "Bar"
	message2.Count = 5

	err, _ = connection.Save(message2)

	c.Assert(err, Equals, nil)

	// Now run a find (hooks will add 2)
	results := connection.Find(&bson.M{"count": 7}, &FooBar{})

	results.Paginate(1, 1)
	res := new(FooBar)

	count := 0

	for results.Next(res) {
		count++
		if count == 1 {
			c.Assert(res.Msg, Equals, "Foo")
		}
	}

	c.Assert(count, Equals, 1)
	// hooks will add 2
	resultsPage2 := connection.Find(&bson.M{"count": 7}, &FooBar{})

	resultsPage2.Paginate(1, 2)

	count2 := 0
	for resultsPage2.Next(res) {
		count2++
		if count2 == 1 {
			c.Assert(res.Msg, Equals, "Bar")
		}
	}

	c.Assert(count2, Equals, 1)

	// connection.Session.DB(config.Database).DropDatabase()
}
