package bongo

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

// For test usage
func getConnection() *Connection {
	conf := &Config{
		ConnectionString: "localhost",
		Database:         "bongotest",
	}

	conn, err := Connect(conf)
	conn.Context.Set("foo", "bar")

	if err != nil {
		panic(err)
	}

	return conn
}

func TestFailSSLConnec(t *testing.T) {
	Convey("should fail to connect to a database because of unsupported ssl flag", t, func() {
		conf := &Config{
			ConnectionString: "mongodb://localhost?ssl=true",
			Database:         "bongotest",
		}

		_, err := Connect(conf)
		So(err.Error(), ShouldEqual, "cannot parse given URI mongodb://localhost?ssl=true due to error: unsupported connection URL option: ssl=true")
	})
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

		conn.Context.Set("foo", "bar")
		value := conn.Context.Get("foo")
		So(value, ShouldEqual, "bar")

		err = conn.Session.Ping()
		So(err, ShouldEqual, nil)
	})
}

func TestRetrieveCollection(t *testing.T) {
	Convey("should be able to retrieve a collection instance from a connection", t, func() {
		conn := getConnection()
		defer conn.Session.Close()
		col := conn.Collection("tests");
		So(col.Name, ShouldEqual, "tests")
		So(col.Connection, ShouldEqual, conn)

		So(col.Context.Get("foo"), ShouldEqual, "bar")

		So(conn.Config.Database, ShouldEqual, col.Database)
	})
	Convey("should be able to retrieve a collection instance from a connection with different databases", t, func() {
		conn := getConnection()
		defer conn.Session.Close()

		col1 := conn.CollectionFromDatabase("tests", "test1");
		So(col1.Name, ShouldEqual, "tests")
		So(col1.Connection, ShouldEqual, conn)
		So(col1.Database, ShouldEqual, "test1")

		col2 := conn.CollectionFromDatabase("tests", "test2");
		So(col2.Name, ShouldEqual, "tests")
		So(col2.Connection, ShouldEqual, conn)
		So(col2.Database, ShouldEqual, "test2")

		So(col2.Connection, ShouldEqual, col1.Connection)
		So(col1.Database, ShouldNotEqual, col2.Database)
	})
}
