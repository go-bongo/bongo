package bongo

import (
	"fmt"
	"github.com/maxwellhealth/mgo/bson"
	. "gopkg.in/check.v1"
	"log"
	"time"
	// "testing"
)

func (s *TestSuite) TestSaveAndFindWithHooks(c *C) {

	// This needs to always be a pointer, otherwise the encryption component won't like it.
	message := new(FooBar)
	message.Msg = "Foo"
	message.Count = 5

	result := connection.Save(message)

	c.Assert(result.Success, Equals, true)

	newMessage := new(FooBar)

	connection.FindById(message.Id, newMessage)

	// Make sure the ids are the same
	c.Assert(newMessage.Id.String(), Equals, message.Id.String())
	c.Assert(newMessage.Msg, Equals, message.Msg)

	// Testing the hook here - it should have run and +1 on BeforeSave and +1 on BeforeCreate and +5 on AfterFind
	c.Assert(newMessage.Count, Equals, 12)

	// Saving it again should run +1 on BeforeSave and +2 on BeforeUpdate
	result = connection.Save(message)

	c.Assert(result.Success, Equals, true)
	c.Assert(message.Count, Equals, 10)

	connection.Session.DB(config.Database).DropDatabase()
}

func (s *TestSuite) TestSaveAndFindWithChild(c *C) {

	// This needs to always be a pointer, otherwise the encryption component won't like it.
	message := new(FooBar)
	message.Msg = "Foo"
	message.Count = 5
	message.Child = &Nested{
		Foo:     "foo",
		BazBing: "bar",
	}
	result := connection.Save(message)

	c.Assert(result.Success, Equals, true)

	newMessage := new(FooBar)

	connection.FindById(message.Id, newMessage)

	c.Assert(newMessage.Child.BazBing, Equals, "bar")
	c.Assert(newMessage.Child.Foo, Equals, "foo")

}

func (s *TestSuite) TestValidationFailure(c *C) {

	message := new(FooBar)
	message.Msg = "Foo"
	message.Count = 3

	result := connection.Save(message)

	c.Assert(result.Err.Error(), Equals, "Validation failed")
	c.Assert(result.ValidationErrors[0], Equals, "count cannot be 3")

}

func (s *TestSuite) TestFindNonExistent(c *C) {

	newMessage := new(FooBar)

	err := connection.FindById(bson.NewObjectId(), newMessage)

	c.Assert(err.Error(), Equals, "Document not found")
}

func (s *TestSuite) TestDelete(c *C) {

	// This needs to always be a pointer, otherwise the encryption component won't like it.
	message := new(FooBar)
	message.Msg = "Foo"
	message.Count = 5

	result := connection.Save(message)

	c.Assert(result.Success, Equals, true)

	connection.Delete(message)

	newMessage := new(FooBar)
	err := connection.FindById(message.Id, newMessage)
	c.Assert(err.Error(), Equals, "Document not found")
	// Make sure the ids are the same
	//

}

func (s *TestSuite) TestFindOne(c *C) {

	// This needs to always be a pointer, otherwise the encryption component won't like it.
	message := new(FooBar)
	message.Msg = "Foo"
	message.Count = 5

	res := connection.Save(message)

	c.Assert(res.Success, Equals, true)

	result := &FooBar{}

	query := bson.M{
		"count": 7,
	}

	err := connection.FindOne(query, result)

	c.Assert(err, Equals, nil)

	c.Assert(string(result.Msg), Equals, "Foo")
	// After find adds 5
	c.Assert(result.Count, Equals, 12)

}

func (s *TestSuite) TestFind(c *C) {

	// This needs to always be a pointer, otherwise the encryption component won't like it.
	message := new(FooBar)
	message.Msg = "Foo"
	message.Count = 5

	result := connection.Save(message)

	c.Assert(result.Success, Equals, true)

	message2 := new(FooBar)
	message2.Msg = "Bar"
	message2.Count = 10

	result = connection.Save(message2)

	c.Assert(result.Success, Equals, true)

	// Now run a find
	results := connection.Find(nil, &FooBar{})

	res := new(FooBar)

	count := 0

	for results.Next(res) {
		count++
		if count == 1 {
			c.Assert(string(res.Msg), Equals, "Foo")
		} else {
			c.Assert(string(res.Msg), Equals, "Bar")
		}
	}

	c.Assert(count, Equals, 2)

}

func (s *TestSuite) TestFindWithPagination(c *C) {

	// This needs to always be a pointer, otherwise the encryption component won't like it.
	message := new(FooBar)
	message.Msg = "Foo"
	message.Count = 5

	result := connection.Save(message)

	c.Assert(result.Success, Equals, true)

	message2 := new(FooBar)
	message2.Msg = "Bar"
	message2.Count = 5

	result = connection.Save(message2)

	c.Assert(result.Success, Equals, true)

	// Now run a find (hooks will add 2)
	results := connection.Find(&bson.M{"count": 7}, &FooBar{})

	results.Paginate(1, 1)
	res := new(FooBar)

	count := 0

	for results.Next(res) {
		count++
		if count == 1 {
			c.Assert(string(res.Msg), Equals, "Foo")
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
			c.Assert(string(res.Msg), Equals, "Bar")
		}
	}

	c.Assert(count2, Equals, 1)

}

type RecursiveChild struct {
	Bar EncryptedString `bson:"bar"`
}
type RecursiveParent struct {
	Id    bson.ObjectId   `bson:"_id"`
	Foo   EncryptedString `bson:"foo"`
	Child *RecursiveChild `bson:"child"`
}

func (s *TestSuite) TestRecursiveSaveWithEncryption(c *C) {
	parent := &RecursiveParent{
		Foo: "foo",
		Child: &RecursiveChild{
			Bar: "bar",
		},
	}

	connection.Save(parent)

	// Fetch natively...

	newParent := &RecursiveParent{}

	// Now fetch using bongo to decrypt...
	connection.Collection("recursive_parent").FindById(parent.Id, newParent)
	c.Assert(string(newParent.Child.Bar), Equals, "bar")

	connection.Collection("recursive_parent").Collection().FindId(parent.Id).One(newParent)

	c.Assert(newParent.Child.Bar, Not(Equals), "bar")
}

// Just to make sure the benchmark will work...
func (s *TestSuite) TestBenchmarkEncryptAndSave(c *C) {
	createAndSaveDocument()
}

func runFindWithBongo(ch chan<- int64) {

	start := time.Now()
	// results := connection.Collection("foobars").Collection().Find(nil)

	// count, _ := results.Count()

	results := connection.Collection("foobars").Find(nil)

	// results.Iter = results.Query.Iter()
	results.Paginate(50, 1)
	// iter := results.Iter()
	for i := 0; i < 50; i++ {
		t := &FooBar{}
		results.Next(t)
	}
	// for i := 0; i < info.RecordsOnPage; i++ {
	// 	t := &FooBar{}
	// 	results.Next(t)
	// }
	elapsed := time.Since(start)

	ch <- elapsed.Nanoseconds()

}

func runFindWithMgo(ch chan<- int64) {

	// sess := connection.Session.Copy()
	// defer sess.Close()
	sess := connection.Session
	start := time.Now()
	results := sess.DB(connection.Config.Database).C("foobars").Find(nil)
	// connection.Collection("foobars").Collection().Find(nil)

	results.Count()

	// results.Paginate(50, 1)

	// arr := make([]*FooBar, 50)
	iter := results.Iter()
	for i := 0; i < 50; i++ {
		t := &FooBar{}
		iter.Next(t)
	}
	// for i := 0; i < info.RecordsOnPage; i++ {
	// 	t := &FooBar{}
	// 	results.Next(t)
	// }
	elapsed := time.Since(start)

	ch <- elapsed.Nanoseconds()

}

func testConcurrentFinds(n int) int {
	ch := make(chan int64)
	for i := 0; i < n; i++ {
		go runFindWithMgo(ch)
	}

	els := 0
	got := 0

loop:
	for {
		select {
		case elapsed := <-ch:
			els += int(elapsed)
			got++

			if got == n {
				break loop
			}
		default:
			time.Sleep(100 * time.Millisecond)
		}
	}

	return els / n
}

func (s *TestSuite) TestConcurrentFinds(c *C) {
	for i := 0; i < 50; i++ {
		// Create 50 tests
		test := &FooBar{
			Msg: "test",
		}
		res := connection.Collection("foobars").Save(test)
		c.Assert(res.Success, Equals, true)
	}

	nums := []int{1, 5, 10, 20, 50, 100, 200}

	for _, n := range nums {
		tot := 0
		for i := 0; i < 10; i++ {
			tot += testConcurrentFinds(n)
		}
		fmt.Println(n, " concurrent finds :: ", ((tot / 10) / 1e6), "ms/req")
	}
}

/////////////////////
/// BENCHMARKS
/////////////////////
func createAndSaveDocument() {
	message := &FooBar{
		Msg:   "Foo",
		Count: 5,
	}

	status := connection.Save(message)
	// log.Println("status:", status.Success)
	if status.Success != true {
		log.Println(status.Error())
		panic(status.Error)
	}
}
func (s *TestSuite) BenchmarkEncryptAndSave(c *C) {

	for i := 0; i < c.N; i++ {
		createAndSaveDocument()
	}
}
