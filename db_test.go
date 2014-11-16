package frat

import (
	"testing"
	. "gopkg.in/check.v1"
	// "labix.org/v2/mgo/bson"
)



// Hook up gocheck into the "go test" runner.
func DbTest(t *testing.T) { TestingT(t) }

type DbTestSuite struct{}

var _ = Suite(&DbTestSuite{})

func (s *DbTestSuite) TestConnect(c *C) {
	config := &MongoConfig{"localhost","gotest"}

	connection := new(MongoConnection)

	connection.Config = config

	connection.Connect()
	defer connection.Session.Close()

	err := connection.Session.Ping()

	c.Assert(err, Equals, nil)
}


