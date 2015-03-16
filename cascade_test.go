package bongo

import (
	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/mgo.v2/bson"
	"reflect"
	"testing"
	"time"
)

type Parent struct {
	DocumentBase `bson:",inline"`
	Bar          string
	Number       int
	FooBar       string
	Children     []ChildRef
	Child        ChildRef
	ChildProp    string `bson:"childProp"`
	diffTracker  *DiffTracker
}

func (f *Parent) GetDiffTracker() *DiffTracker {
	v := reflect.ValueOf(f.diffTracker)
	if !v.IsValid() || v.IsNil() {
		f.diffTracker = NewDiffTracker(f)
	}

	return f.diffTracker
}

type Child struct {
	DocumentBase `bson:",inline"`
	ParentId     bson.ObjectId `bson:",omitempty"`
	Name         string
	SubChild     SubChildRef `bson:"subChild"`
	ChildProp    string
	diffTracker  *DiffTracker
}

func (c *Child) GetCascade(collection *Collection) []*CascadeConfig {

	ref := ChildRef{
		Id:       c.Id,
		Name:     c.Name,
		SubChild: c.SubChild,
	}
	connection := collection.Connection
	cascadeSingle := &CascadeConfig{
		Collection:  connection.Collection("parents"),
		Properties:  []string{"_id", "name", "subChild.foo", "subChild._id"},
		Data:        ref,
		ThroughProp: "child",
		RelType:     REL_ONE,
		Query: bson.M{
			"_id": c.ParentId,
		},
	}

	cascadeCopy := &CascadeConfig{
		Collection: connection.Collection("parents"),
		Properties: []string{"childProp"},
		Data: map[string]interface{}{
			"childProp": c.ChildProp,
		},
		RelType: REL_ONE,
		Query: bson.M{
			"_id": c.ParentId,
		},
	}

	cascadeMulti := &CascadeConfig{
		Collection:  connection.Collection("parents"),
		Properties:  []string{"_id", "name", "subChild.foo", "subChild._id"},
		Data:        ref,
		ThroughProp: "children",
		RelType:     REL_MANY,
		Query: bson.M{
			"_id": c.ParentId,
		},
	}

	if c.GetDiffTracker().Modified("ParentId") {

		origId, _ := c.diffTracker.GetOriginalValue("ParentId")
		if origId != nil {
			oldQuery := bson.M{
				"_id": origId,
			}
			cascadeSingle.OldQuery = oldQuery
			cascadeCopy.OldQuery = oldQuery
			cascadeMulti.OldQuery = oldQuery
		}

	}

	return []*CascadeConfig{cascadeSingle, cascadeMulti, cascadeCopy}
}

func (f *Child) GetDiffTracker() *DiffTracker {
	v := reflect.ValueOf(f.diffTracker)
	if !v.IsValid() || v.IsNil() {
		f.diffTracker = NewDiffTracker(f)
	}

	return f.diffTracker
}

type SubChild struct {
	DocumentBase `bson:",inline"`
	Foo          string
	ChildId      bson.ObjectId
}

func (c *SubChild) GetCascade(collection *Collection) []*CascadeConfig {
	ref := SubChildRef{
		Id:  c.Id,
		Foo: c.Foo,
	}
	connection := collection.Connection
	cascadeSingle := &CascadeConfig{
		Collection:  connection.Collection("children"),
		Properties:  []string{"_id", "foo"},
		Data:        ref,
		ThroughProp: "subChild",
		RelType:     REL_ONE,
		Query: bson.M{
			"_id": c.ChildId,
		},
		Nest:     true,
		Instance: &Child{},
	}

	return []*CascadeConfig{cascadeSingle}
}

type SubChildRef struct {
	Id  bson.ObjectId `bson:"_id,omitempty"`
	Foo string
}

type ChildRef struct {
	Id       bson.ObjectId `bson:"_id,omitempty"`
	Name     string
	SubChild SubChildRef
}

func TestCascade(t *testing.T) {
	connection := getConnection()
	// defer connection.Session.Close()

	Convey("Cascade Save/Delete - full runthrough", t, func() {
		connection.Session.DB("bongotest").DropDatabase()
		collection := connection.Collection("parents")

		childCollection := connection.Collection("children")
		subchildCollection := connection.Collection("subchildren")
		parent := &Parent{
			Bar:    "Testy McGee",
			Number: 5,
		}

		parent2 := &Parent{
			Bar:    "Other Parent",
			Number: 10,
		}

		err := collection.Save(parent)
		So(err, ShouldEqual, nil)
		err = collection.Save(parent2)
		So(err, ShouldEqual, nil)

		child := &Child{
			ParentId:  parent.Id,
			Name:      "Foo McGoo",
			ChildProp: "Doop McGoop",
		}
		err = childCollection.Save(child)

		// Wait a sec for the go routine to finish.
		time.Sleep(100 * time.Millisecond)

		So(err, ShouldEqual, nil)

		child.GetDiffTracker().Reset()
		newParent := &Parent{}
		collection.FindById(parent.Id, newParent)

		So(newParent.Child.Name, ShouldEqual, "Foo McGoo")
		So(newParent.Child.Id.Hex(), ShouldEqual, child.Id.Hex())
		So(newParent.Children[0].Name, ShouldEqual, "Foo McGoo")
		So(newParent.Children[0].Id.Hex(), ShouldEqual, child.Id.Hex())

		// No through prop should populate directly o the parent
		So(newParent.ChildProp, ShouldEqual, "Doop McGoop")

		// Now change the child parent Id...
		child.ParentId = parent2.Id
		So(child.GetDiffTracker().Modified("ParentId"), ShouldEqual, true)

		err = childCollection.Save(child)
		So(err, ShouldEqual, nil)

		// Wait a sec for the go routine to finish.
		time.Sleep(100 * time.Millisecond)

		child.diffTracker.Reset()
		// Now make sure it says the parent id DIDNT change, because we just reset the tracker
		So(child.GetDiffTracker().Modified("ParentId"), ShouldEqual, false)

		newParent1 := &Parent{}
		collection.FindById(parent.Id, newParent1)
		So(newParent1.Child.Name, ShouldEqual, "")
		So(newParent1.ChildProp, ShouldEqual, "")
		So(len(newParent1.Children), ShouldEqual, 0)
		newParent2 := &Parent{}
		collection.FindById(parent2.Id, newParent2)
		So(newParent2.ChildProp, ShouldEqual, "Doop McGoop")
		So(newParent2.Child.Name, ShouldEqual, "Foo McGoo")
		So(newParent2.Child.Id.Hex(), ShouldEqual, child.Id.Hex())
		So(newParent2.Children[0].Name, ShouldEqual, "Foo McGoo")
		So(newParent2.Children[0].Id.Hex(), ShouldEqual, child.Id.Hex())

		// Make a new sub child, save it, and it should cascade to the child AND the parent
		subChild := &SubChild{
			Foo:     "MySubChild",
			ChildId: child.Id,
		}

		err = subchildCollection.Save(subChild)
		So(err, ShouldEqual, nil)

		// Wait a sec for the go routine to finish.
		time.Sleep(100 * time.Millisecond)

		// Fetch the parent
		newParent3 := &Parent{}
		collection.FindById(parent2.Id, newParent3)
		So(newParent3.Child.SubChild.Foo, ShouldEqual, "MySubChild")
		So(newParent3.Child.SubChild.Id.Hex(), ShouldEqual, subChild.Id.Hex())

		newParent4 := &Parent{}
		err = childCollection.DeleteDocument(child)

		// Wait a sec for the go routine to finish.
		time.Sleep(100 * time.Millisecond)

		So(err, ShouldEqual, nil)
		collection.FindById(parent2.Id, newParent4)
		So(newParent4.Child.Name, ShouldEqual, "")
		So(newParent4.ChildProp, ShouldEqual, "")
		So(len(newParent4.Children), ShouldEqual, 0)

	})

	Convey("MapFromCascadeProperties", t, func() {
		parent := &Parent{
			Bar: "bar",
			Child: ChildRef{
				Name: "child",
				SubChild: SubChildRef{
					Foo: "foo",
				},
			},
			Number: 5,
		}

		props := []string{"bar", "child.name"}

		mp := MapFromCascadeProperties(props, parent)

		So(len(mp), ShouldEqual, 2)
		So(mp["bar"], ShouldEqual, "bar")

		submp := mp["child"].(map[string]interface{})
		So(submp["name"], ShouldEqual, "child")

	})

}
