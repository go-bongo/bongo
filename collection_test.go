package bongo

import (
	. "gopkg.in/check.v1"
	"labix.org/v2/mgo/bson"
	// "testing"
)

func (s *TestSuite) TestSaveAndFindWithHooks(c *C) {

	connection := Connect(config)

	defer connection.Session.Close()

	// This needs to always be a pointer, otherwise the encryption component won't like it.
	message := new(FooBar)
	message.Msg = "Foo"
	message.Count = 5

	err, _ := connection.Save(message)

	c.Assert(err, Equals, nil)

	newMessage := new(FooBar)

	connection.FindById(message.Id, newMessage)

	// Make sure the ids are the same
	c.Assert(newMessage.Id.String(), Equals, message.Id.String())
	c.Assert(newMessage.Msg, Equals, message.Msg)

	// Testing the hook here - it should have run and +1 on BeforeSave and +1 on BeforeCreate and +5 on AfterFind
	c.Assert(newMessage.Count, Equals, 12)

	// Saving it again should run +1 on BeforeSave and +2 on BeforeUpdate
	err, _ = connection.Save(message)

	c.Assert(err, Equals, nil)
	c.Assert(message.Count, Equals, 10)

	connection.Session.DB(config.Database).DropDatabase()
}

func (s *TestSuite) TestSaveAndFindWithChild(c *C) {

	connection := Connect(config)

	defer connection.Session.Close()

	// This needs to always be a pointer, otherwise the encryption component won't like it.
	message := new(FooBar)
	message.Msg = "Foo"
	message.Count = 5
	message.Child = &Child{
		Foo:     "foo",
		BazBing: "bar",
	}
	err, _ := connection.Save(message)

	c.Assert(err, Equals, nil)

	newMessage := new(FooBar)

	connection.FindById(message.Id, newMessage)

	c.Assert(newMessage.Child.BazBing, Equals, "bar")
	c.Assert(newMessage.Child.Foo, Equals, "foo")

	connection.Session.DB(config.Database).DropDatabase()
}

func (s *TestSuite) TestValidationFailure(c *C) {

	connection := Connect(config)

	defer connection.Session.Close()

	message := new(FooBar)
	message.Msg = "Foo"
	message.Count = 3

	err, errs := connection.Save(message)

	c.Assert(err.Error(), Equals, "Validation failed")
	c.Assert(errs[0], Equals, "count cannot be 3")

	connection.Session.DB(config.Database).DropDatabase()
}

func (s *TestSuite) TestFindNonExistent(c *C) {

	connection := Connect(config)

	defer connection.Session.Close()

	newMessage := new(FooBar)

	err := connection.FindById(bson.NewObjectId(), newMessage)

	c.Assert(err.Error(), Equals, "not found")
	connection.Session.DB(config.Database).DropDatabase()
}

func (s *TestSuite) TestDelete(c *C) {

	connection := Connect(config)

	defer connection.Session.Close()

	// This needs to always be a pointer, otherwise the encryption component won't like it.
	message := new(FooBar)
	message.Msg = "Foo"
	message.Count = 5

	err, _ := connection.Save(message)

	c.Assert(err, Equals, nil)

	connection.Delete(message)

	newMessage := new(FooBar)
	err = connection.FindById(message.Id, newMessage)
	c.Assert(err.Error(), Equals, "not found")
	// Make sure the ids are the same
	//
	connection.Session.DB(config.Database).DropDatabase()

}

func (s *TestSuite) TestFindOne(c *C) {

	connection := Connect(config)

	defer connection.Session.Close()

	// This needs to always be a pointer, otherwise the encryption component won't like it.
	message := new(FooBar)
	message.Msg = "Foo"
	message.Count = 5

	err, _ := connection.Save(message)

	c.Assert(err, Equals, nil)

	result := &FooBar{}

	query := bson.M{
		"count": 7,
	}

	err = connection.FindOne(query, result)

	c.Assert(err, Equals, nil)

	c.Assert(result.Msg, Equals, "Foo")
	c.Assert(result.Count, Equals, 7)

	connection.Session.DB(config.Database).DropDatabase()

}

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

	connection.Session.DB(config.Database).DropDatabase()
}

/////////////////////
/// BENCHMARKS
/////////////////////
func createAndSaveDocument(conn *Connection) {
	message := &FooBar{
		Msg:   "Foo",
		Count: 5,
	}

	err, _ := conn.Save(message)
	if err != nil {
		panic(err)
	}
}
func (s *TestSuite) BenchmarkEncryptAndSave(c *C) {

	connection := Connect(config)

	defer connection.Session.Close()

	c.ResetTimer()
	for i := 0; i < c.N; i++ {
		createAndSaveDocument(connection)
	}
	connection.Session.DB(config.Database).DropDatabase()
}
