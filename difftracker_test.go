package bongo

import (
	. "github.com/smartystreets/goconvey/convey"
	"reflect"
	"testing"
	"time"
)

type FooChangeTest struct {
	DocumentBase `bson:",inline"`
	StringVal    string
	IntVal       int
	Timestamp    time.Time
	diffTracker  *DiffTracker
	Arr          []string
}

func (f *FooChangeTest) GetDiffTracker() *DiffTracker {
	v := reflect.ValueOf(f.diffTracker)
	if !v.IsValid() || v.IsNil() {
		f.diffTracker = NewDiffTracker(f)
	}

	return f.diffTracker
}

type FooBarChangeTest struct {
	FooVal *FooChangeTest
	BarVal string
}

func TestDiffTracker(t *testing.T) {
	Convey("DiffTracker", t, func() {
		Convey("should get changed fields when comparing two structs", func() {
			foo1 := &FooChangeTest{
				StringVal: "foo",
				IntVal:    1,
				Arr:       []string{},
			}
			foo2 := &FooChangeTest{
				StringVal: "bar",
				IntVal:    2,
				Arr:       []string{},
			}

			diffs, err := GetChangedFields(foo1, foo2, false)
			So(err, ShouldEqual, nil)
			So(len(diffs), ShouldEqual, 2)
			So(diffs[0], ShouldEqual, "StringVal")
			So(diffs[1], ShouldEqual, "IntVal")
		})

		Convey("should get changed fields when comparing two structs with pointers", func() {
			foobar1 := &FooBarChangeTest{
				FooVal: &FooChangeTest{
					StringVal: "foo",
					IntVal:    5,
				},
				BarVal: "bar",
			}

			foobar2 := &FooBarChangeTest{
				FooVal: &FooChangeTest{
					StringVal: "foo",
					IntVal:    10,
					Timestamp: time.Now(),
				},
				BarVal: "BAR",
			}

			diffs, err := GetChangedFields(foobar1, foobar2, false)
			So(err, ShouldEqual, nil)
			So(len(diffs), ShouldEqual, 3)
			So(diffs[0], ShouldEqual, "FooVal.IntVal")
			So(diffs[1], ShouldEqual, "FooVal.Timestamp")
			So(diffs[2], ShouldEqual, "BarVal")
		})

		Convey("should get changed fields when comparing two structs with nil pointers", func() {
			foobar1 := &FooBarChangeTest{
				BarVal: "bar",
			}

			foobar2 := &FooBarChangeTest{
				BarVal: "BAR",
			}

			diffs, err := GetChangedFields(foobar1, foobar2, false)
			So(err, ShouldEqual, nil)
			So(len(diffs), ShouldEqual, 1)
			So(diffs[0], ShouldEqual, "BarVal")
		})

		Convey("should get fields modified since difftracker reset", func() {
			foo1 := &FooChangeTest{
				StringVal: "foo",
				IntVal:    1,
			}

			foo1.GetDiffTracker().Reset()

			sess, _ := foo1.GetDiffTracker().NewSession(false)

			So(sess.Modified("StringVal"), ShouldEqual, false)
			foo1.StringVal = "bar"

			sess, _ = foo1.GetDiffTracker().NewSession(false)
			So(sess.Modified("StringVal"), ShouldEqual, true)
		})
	})

}
