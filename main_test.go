package bongo

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

// For test usage
func getConnection() *Connection {
	conf := &Config{
		ConnectionString: "localhost",
		Database:         "bongotest",
	}

	conn, err := Connect(conf)

	if err != nil {
		panic(err)
	}

	return conn
}

func TestConnect(t *testing.T) {
	Convey("should be able to connect to a database using a config", t, func() {
		conf := &Config{
			ConnectionString: "localhost",
			Database:         "bongotest",
		}

		conn, err := Connect(conf)
		defer conn.Session.Close()
		So(err, ShouldEqual, nil)

		err = conn.Session.Ping()
		So(err, ShouldEqual, nil)
	})
}

func TestRetrieveCollection(t *testing.T) {
	Convey("should be able to retrieve a collection instance from a connection", t, func() {
		conn := getConnection()
		defer conn.Session.Close()
		col := conn.Collection("tests")

		So(col.Name, ShouldEqual, "tests")
		So(col.Connection, ShouldEqual, conn)
	})
}
