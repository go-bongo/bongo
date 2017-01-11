package bongo

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func Test_Context(t *testing.T) {
	Convey("Context", t, func() {

		Convey("Setting context, checking it and deleting it", func() {
			c := &Context{}
			c.Set("foo", "bar")
			So(c.Get("foo"), ShouldEqual, "bar")
			So(c.Delete("foo"), ShouldBeTrue)
		})

		Convey("Invalid Keys", func() {
			c := &Context{}
			So(c.Get("foo"), ShouldBeNil)
			So(c.Delete("foo"), ShouldBeFalse)
		})

	})

}
