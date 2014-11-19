package bongo

import (
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
