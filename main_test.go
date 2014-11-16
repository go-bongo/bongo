package frat

import (
	"testing"
	. "gopkg.in/check.v1"
	"labix.org/v2/mgo/bson"
)



// var key = []byte("asdf1234asdf1234")


// Hook up gocheck into the "go test" runner.
// func Test(t *testing.T) { TestingT(t) }

// type TestSuite struct{}

var _ = Suite(&TestSuite{})

type Child struct {
	Foo string
	BazBing string `bson:"bazBing"`
}

type FooBar struct {
	Id    bson.ObjectId   `bson:"_id"`
	Msg   string        `encrypted:"true",bson="msg"`
	Count int           `encrypted:"true",bson="count"`
	Child *Child
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
	ConnectionString:"localhost",
	Database:"gotest",
	EncryptionKey:"asdf1234asdf1234",
}


func (s *TestSuite) TestConnect(c *C) {
	

	connection := new(Connection)

	connection.Config = config

	connection.Connect()
	defer connection.Session.Close()

	err := connection.Session.Ping()

	c.Assert(err, Equals, nil)

	connection.Session.DB(config.Database).DropDatabase()
}

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
		Foo:"foo",
		BazBing:"bar",
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


/////////////////////
/// BENCHMARKS
/////////////////////
func createAndSaveDocument(conn *Connection) {
 	message := &FooBar{
 		Msg:"Foo",
 		Count:5,
 	}

 	err, _ := conn.Save(message)
 	if err != nil {
 		panic(err)
 	}
}


func BenchmarkEncryptedAndSave(b *testing.B) {

	connection := Connect(config)


	defer connection.Session.Close()



	for i := 0; i < b.N; i++ {
	    createAndSaveDocument(connection)
	}
	connection.Session.DB(config.Database).DropDatabase()
}
