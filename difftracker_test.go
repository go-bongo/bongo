package bongo

import (
	. "gopkg.in/check.v1"
	"labix.org/v2/mgo/bson"
	"reflect"
)

type FooChangeTest struct {
	Id          bson.ObjectId `bson:"_id,omitempty"`
	StringVal   string
	IntVal      int
	diffTracker *DiffTracker
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

func (s *TestSuite) TestGetChangedFields(c *C) {
	foo1 := &FooChangeTest{
		StringVal: "foo",
		IntVal:    1,
	}
	foo2 := &FooChangeTest{
		StringVal: "bar",
		IntVal:    2,
	}

	diffs, err := getChangedFields(foo1, foo2)
	c.Assert(err, Equals, nil)
	c.Assert(len(diffs), Equals, 2)
	c.Assert(diffs[0], Equals, "StringVal")
	c.Assert(diffs[1], Equals, "IntVal")

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
		},
		BarVal: "BAR",
	}

	diffs, err = getChangedFields(foobar1, foobar2)
	c.Assert(err, Equals, nil)
	c.Assert(len(diffs), Equals, 2)
	c.Assert(diffs[0], Equals, "FooVal.IntVal")
	c.Assert(diffs[1], Equals, "BarVal")

}

func (s *TestSuite) TestModified(c *C) {
	foo1 := &FooChangeTest{
		StringVal: "foo",
		IntVal:    1,
	}

	foo1.GetDiffTracker().Reset()

	foo1.StringVal = "bar"

	c.Assert(foo1.diffTracker.Modified("StringVal"), Equals, true)
}
