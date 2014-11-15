package frat

import (
	"testing"
	. "gopkg.in/check.v1"
)



// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { TestingT(t) }

type TestSuite struct{}

var _ = Suite(&TestSuite{})

var key = []byte("asdf1234asdf1234")


/**
 * Full document encryption
 */
type Document struct {

}

type Name struct {
	First string
	Last string
}

type Person struct {
	Document
	Name `encrypted:"true"`
	Phone string
	Number int `encrypted:"true" bson:"Foo"`
	Other bool `encrypted:"true""`
	Arr []string `encrypted:"true"`
}


func (s *TestSuite) TestEncryptDecryptDocument(c *C) {
 	p := &Person{
 		Name:Name{
	 		First:"Jason",
	 		Last:"Raede",
	 	},
 		Phone:"555-555-5555",
 		Number:5,
 		Arr:[]string{"foo","bar","baz","bing"},
 	}


 	/**
 	 * @type map[string]interface{}
 	 */
 	encrypted := EncryptDocument(key, p)

 	// Name should be a string, encrypted from the json encoding of the Name struct
 	c.Assert(encrypted["name"], Not(Equals), "Jason")
 	// Phone is not encrypted
 	c.Assert(encrypted["phone"], Equals, "555-555-5555")
 	// Number is encrypted as "Foo"
 	c.Assert(encrypted["Foo"], Not(Equals), 5)

 	c.Assert(encrypted["arr"], Not(Equals), p.Arr)


 	newP := new(Person)

 	DecryptDocument(key, encrypted, newP)

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


}


/////////////////////
/// BENCHMARKS
/////////////////////
func encryptDecryptDocument() {
 	p := &Person{
 		Name:Name{
	 		First:"Jason",
	 		Last:"Raede",
	 	},
 		Phone:"555-555-5555",
 		Number:5,
 	}

 	encrypted := EncryptDocument(key, p)
 	newP := new(Person)

 	DecryptDocument(key, encrypted, newP)
}

// Note - potential for this to be ~20% faster if on the first pass we make an array of all the encrypted strings and bson values so we don't have to introspect the tags every time for the same Type. OK for now.
func BenchmarkEncryptDecryptDocument(b *testing.B) {
    for i := 0; i < b.N; i++ {
        encryptDecryptDocument()
    }
}
