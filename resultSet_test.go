package bongo

import (
	. "github.com/smartystreets/goconvey/convey"
	"github.com/globalsign/mgo/bson"
	"testing"
)

func TestResultSet(t *testing.T) {
	conn := getConnection()
	collection := conn.Collection("tests")
	defer conn.Session.Close()

	Convey("Basic find/pagination", t, func() {
		// Create 10 things
		for i := 0; i < 10; i++ {
			doc := &noHookDocument{}
			collection.Save(doc)
		}

		Convey("should let you iterate through all results without paginating", func() {
			rset := collection.Find(nil)
			defer rset.Free()
			count := 0

			doc := &noHookDocument{}

			for rset.Next(doc) {
				count++
			}

			So(count, ShouldEqual, 10)
		})

		Convey("should let you paginate and get pagination info", func() {
			rset := collection.Find(nil)
			defer rset.Free()
			info, err := rset.Paginate(3, 1)
			So(err, ShouldEqual, nil)
			So(info.TotalRecords, ShouldEqual, 10)
			So(info.TotalPages, ShouldEqual, 4)
			So(info.Current, ShouldEqual, 1)
			So(info.PerPage, ShouldEqual, 3)
			So(info.RecordsOnPage, ShouldEqual, 3)

			rset2 := collection.Find(nil)
			defer rset2.Free()
			info, err = rset2.Paginate(3, 4)
			So(err, ShouldEqual, nil)
			So(info.TotalRecords, ShouldEqual, 10)
			So(info.TotalPages, ShouldEqual, 4)
			So(info.Current, ShouldEqual, 4)
			So(info.PerPage, ShouldEqual, 3)
			So(info.RecordsOnPage, ShouldEqual, 1)
		})

		Reset(func() {
			conn.Session.DB("bongotest").DropDatabase()
		})
	})

	Convey("Find/pagination w/ query", t, func() {
		// Create 10 things
		for i := 0; i < 5; i++ {
			doc := &noHookDocument{}
			doc.Name = "foo"
			collection.Save(doc)
		}
		for i := 0; i < 5; i++ {
			doc := &noHookDocument{}
			doc.Name = "bar"
			collection.Save(doc)
		}

		Convey("should let you iterate through all filtered results without paginating", func() {
			rset := collection.Find(bson.M{
				"name": "foo",
			})
			defer rset.Free()

			count := 0

			doc := &noHookDocument{}

			for rset.Next(doc) {
				count++
			}

			So(count, ShouldEqual, 5)
		})

		Convey("should let you paginate and get pagination info on filtered query", func() {
			rset := collection.Find(bson.M{
				"name": "foo",
			})
			defer rset.Free()
			info, err := rset.Paginate(3, 1)
			So(err, ShouldEqual, nil)
			So(info.TotalRecords, ShouldEqual, 5)
			So(info.TotalPages, ShouldEqual, 2)
			So(info.Current, ShouldEqual, 1)
			So(info.PerPage, ShouldEqual, 3)
			So(info.RecordsOnPage, ShouldEqual, 3)

			rset2 := collection.Find(bson.M{
				"name": "foo",
			})
			defer rset2.Free()
			info, err = rset2.Paginate(3, 2)
			So(err, ShouldEqual, nil)
			So(info.TotalRecords, ShouldEqual, 5)
			So(info.TotalPages, ShouldEqual, 2)
			So(info.Current, ShouldEqual, 2)
			So(info.PerPage, ShouldEqual, 3)
			So(info.RecordsOnPage, ShouldEqual, 2)
		})

		Reset(func() {
			conn.Session.DB("bongotest").DropDatabase()
		})
	})

	Convey("hooks", t, func() {
		// Create 10 things
		for i := 0; i < 10; i++ {
			doc := &hookedDocument{}
			collection.Save(doc)
		}

		Convey("should let you iterate through all results without paginating", func() {
			rset := collection.Find(nil)
			defer rset.Free()
			count := 0

			doc := &hookedDocument{}

			for rset.Next(doc) {
				So(doc.RanAfterFind, ShouldEqual, true)
				count++
			}

			So(count, ShouldEqual, 10)
		})

		Reset(func() {
			conn.Session.DB("bongotest").DropDatabase()
		})
	})
}
