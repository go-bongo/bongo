package bongo

import (
	"reflect"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestLowerInitial(t *testing.T) {

	Convey("LowerInitial", t, func() {

		Convey("ValidateRequired()", func() {
			So(lowerInitial("foo"), ShouldEqual, "foo")
			So(lowerInitial("Foo"), ShouldEqual, "foo")
			So(lowerInitial("Abba"), ShouldEqual, "abba")
			So(lowerInitial("ABBA"), ShouldEqual, "aBBA")
			So(lowerInitial(""), ShouldEqual, "")
		})

	})
}

func TestBsonName(t *testing.T) {
	Convey("GetBsonName", t, func() {
		type Model struct {
			Property  string `bson:"property" json:"property"`
			Property2 string `json:"property3"`
		}

		Convey("GetBsonName(Model)", func() {
			obj := Model{}

			val := reflect.Indirect(reflect.ValueOf(obj))
			field, _ := val.Type().FieldByName("Property")
			So(GetBsonName(field), ShouldEqual, "property")

			val2 := reflect.Indirect(reflect.ValueOf(obj))
			field2, _ := val2.Type().FieldByName("Property2")
			So(GetBsonName(field2), ShouldEqual, "property2")
		})

	})
}
