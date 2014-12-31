package bongo

import (
	// "encoding/json"
	"fmt"
	. "gopkg.in/check.v1"
	"labix.org/v2/mgo/bson"
	// "testing"
	"time"
)

/**
 * Full document encryption
 */
type Name struct {
	First string
	Last  string
}

type Person struct {
	Name           `bongo:"encrypted"`
	Phone          string
	Number         int             `bongo:"encrypted" bson:"Foo"`
	Other          bool            `bongo:"encrypted"`
	Arr            []string        `bongo:"encrypted"`
	IdVal          bson.ObjectId   `bongo:"encrypted" bson:",omitempty"`
	IdArr          []bson.ObjectId `bongo:"encrypted"`
	DateVal        time.Time       `bongo:"encrypted"`
	DateArr        []time.Time     `bongo:"encrypted"`
	UnencryptedArr []string
}

func (s *TestSuite) TestEncryptInitializeDocumentFromDB(c *C) {
	id := bson.NewObjectId()

	currentTime := time.Now()

	p := &Person{
		Name: Name{
			First: "Jason",
			Last:  "Raede",
		},
		Phone:   "555-555-5555",
		Number:  5,
		Arr:     []string{"foo", "bar", "baz", "bing"},
		IdVal:   id,
		IdArr:   []bson.ObjectId{id},
		DateVal: currentTime,
		DateArr: []time.Time{currentTime},
		// UnencryptedArr: []string{"foo", "bar"},
	}

	fmt.Println(p)
	/**
	 * @type map[string]interface{}
	 */
	encrypted := PrepDocumentForSave(key, p)

	// Name should be a string, encrypted from the json encoding of the Name struct
	c.Assert(encrypted["name"], Not(Equals), "Jason")
	// Phone is not encrypted
	c.Assert(encrypted["phone"], Equals, "555-555-5555")
	// Number is encrypted as "Foo"
	c.Assert(encrypted["Foo"], Not(Equals), 5)
	c.Assert(encrypted["arr"], Not(Equals), p.Arr)

	newP := new(Person)

	fmt.Println(encrypted)
	InitializeDocumentFromDB(key, encrypted, newP)

	// Encrypted structs should be converted from JSON string to the actual struct
	c.Assert(newP.Name.First, Equals, "Jason")
	c.Assert(newP.Name.Last, Equals, "Raede")

	// Unencrypted fields should remain that way
	c.Assert(newP.Phone, Equals, "555-555-5555")

	c.Assert(newP.Number, Equals, 5)

	c.Assert(len(newP.Arr), Equals, 4)

	c.Assert(newP.Arr[0], Equals, "foo")
	c.Assert(newP.Arr[1], Equals, "bar")
	c.Assert(newP.Arr[2], Equals, "baz")
	c.Assert(newP.Arr[3], Equals, "bing")

	c.Assert(newP.IdVal.Hex(), Equals, id.Hex())
	c.Assert(newP.IdArr[0].Hex(), Equals, id.Hex())

	c.Assert(newP.DateVal.Unix(), Equals, currentTime.Unix())
	c.Assert(newP.DateArr[0].Unix(), Equals, currentTime.Unix())
}

/////////////////////
/// BENCHMARKS
/////////////////////
func encryptInitializeDocumentFromDB() {
	p := &Person{
		Name: Name{
			First: "Jason",
			Last:  "Raede",
		},
		Phone:  "555-555-5555",
		Number: 5,
	}

	encrypted := PrepDocumentForSave(key, p)
	newP := new(Person)

	InitializeDocumentFromDB(key, encrypted, newP)
}
func (s *TestSuite) TestEncryptInitializeWithMissingValues(c *C) {
	encryptInitializeDocumentFromDB()
}

// Note - potential for this to be ~20% faster if on the first pass we make an array of all the encrypted strings and bson values so we don't have to introspect the tags every time for the same Type. OK for now.
func (s *TestSuite) BenchmarkEncryptInitializeDocumentFromDB(c *C) {
	for i := 0; i < c.N; i++ {
		encryptInitializeDocumentFromDB()
	}
}
