package bongo

import (
	// "fmt"
	"github.com/maxwellhealth/mgo/bson"
	. "gopkg.in/check.v1"
	"log"
	"reflect"
)

type Parent struct {
	Id          bson.ObjectId `bson:"_id"`
	Bar         string        `bongo:"encrypted"`
	Number      int
	FooBar      string
	Children    []ChildRef `bongo:"cascadedFrom=children"`
	Child       ChildRef   `bongo:"cascadedFrom=children"`
	ChildProp   string     `bson:"childProp"`
	diffTracker *DiffTracker
}

func (f *Parent) GetDiffTracker() *DiffTracker {
	log.Println("Getting diff tracker")
	v := reflect.ValueOf(f.diffTracker)
	if !v.IsValid() || v.IsNil() {
		f.diffTracker = NewDiffTracker(f)
	}

	return f.diffTracker
}

func (c *Child) GetCascade(collection *Collection) []*CascadeConfig {

	connection := collection.Connection
	cascadeSingle := &CascadeConfig{
		Collection:  connection.Collection("parents"),
		Properties:  []string{"name", "subChild.foo", "subChild._id"},
		ThroughProp: "child",
		RelType:     REL_ONE,
		Query: bson.M{
			"_id": c.ParentId,
		},
	}

	cascadeCopy := &CascadeConfig{
		Collection:  connection.Collection("parents"),
		Properties:  []string{"childProp"},
		ThroughProp: "",
		RelType:     REL_ONE,
		Query: bson.M{
			"_id": c.ParentId,
		},
	}

	cascadeMulti := &CascadeConfig{
		Collection:  connection.Collection("parents"),
		Properties:  []string{"name", "subChild.foo", "subChild._id"},
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

func (c *SubChild) GetCascade(collection *Collection) []*CascadeConfig {
	connection := collection.Connection
	cascadeSingle := &CascadeConfig{
		Collection:  connection.Collection("children"),
		Properties:  []string{"foo"},
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

type Child struct {
	Id          bson.ObjectId `bson:"_id"`
	ParentId    bson.ObjectId
	Name        string `bongo:"encrypted"`
	SubChild    SubChildRef
	ChildProp   string
	diffTracker *DiffTracker
}

func (f *Child) GetDiffTracker() *DiffTracker {
	v := reflect.ValueOf(f.diffTracker)
	if !v.IsValid() || v.IsNil() {
		f.diffTracker = NewDiffTracker(f)
	}

	return f.diffTracker
}

type SubChild struct {
	Id      bson.ObjectId `bson:"_id,omitempty"`
	Foo     string
	ChildId bson.ObjectId `bson:",omitempty"`
}

type SubChildRef struct {
	Id  bson.ObjectId `bson:"_id,omitempty"`
	Foo string
}

type ChildRef struct {
	Id       bson.ObjectId `bson:"_id,omitempty"`
	Name     string        `bongo:"encrypted"`
	SubChild SubChildRef
}

func (s *TestSuite) TestCascade(c *C) {

	collection := connection.Collection("parents")

	connection.Config.EncryptionKeyPerCollection = map[string][]byte{
		"parents":  []byte("asdf1234asdf1234"),
		"children": []byte("1234asdf1234asdf"),
	}

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

	res := collection.Save(parent)

	c.Assert(res.Success, Equals, true)
	res = collection.Save(parent2)
	c.Assert(res.Success, Equals, true)

	child := &Child{
		ParentId:  parent.Id,
		Name:      "Foo McGoo",
		ChildProp: "Doop McGoop",
	}

	res = childCollection.Save(child)

	if !res.Success {
		log.Println(res.Error())
		return
	}
	c.Assert(res.Success, Equals, true)
	child.GetDiffTracker().Reset()
	newParent := &Parent{}
	collection.FindById(parent.Id, newParent)

	c.Assert(newParent.Child.Name, Equals, "Foo McGoo")

	c.Assert(newParent.Child.Id.Hex(), Equals, child.Id.Hex())
	c.Assert(newParent.Children[0].Name, Equals, "Foo McGoo")
	c.Assert(newParent.Children[0].Id.Hex(), Equals, child.Id.Hex())
	// No through prop should populate directly o the parent
	newMap := make(map[string]interface{})
	collection.Collection().FindId(parent.Id).One(newMap)

	c.Assert(newParent.ChildProp, Equals, "Doop McGoop")

	// Now change the child parent Id...
	child.ParentId = parent2.Id
	c.Assert(child.GetDiffTracker().Modified("ParentId"), Equals, true)

	res = childCollection.Save(child)
	child.diffTracker.Reset()
	c.Assert(res.Success, Equals, true)
	// Now make sure it says the parent id DIDNT change, because we just reset the tracker
	c.Assert(child.GetDiffTracker().Modified("ParentId"), Equals, false)

	newParent1 := &Parent{}
	collection.FindById(parent.Id, newParent1)
	c.Assert(newParent1.Child.Name, Equals, "")
	c.Assert(newParent1.ChildProp, Equals, "")
	c.Assert(len(newParent1.Children), Equals, 0)
	newParent2 := &Parent{}
	collection.FindById(parent2.Id, newParent2)
	c.Assert(newParent2.ChildProp, Equals, "Doop McGoop")
	c.Assert(newParent2.Child.Name, Equals, "Foo McGoo")
	c.Assert(newParent2.Child.Id.Hex(), Equals, child.Id.Hex())
	c.Assert(newParent2.Children[0].Name, Equals, "Foo McGoo")
	c.Assert(newParent2.Children[0].Id.Hex(), Equals, child.Id.Hex())

	// Make a new sub child, save it, and it should cascade to the child AND the parent
	subChild := &SubChild{
		Foo:     "MySubChild",
		ChildId: child.Id,
	}

	res = subchildCollection.Save(subChild)
	c.Assert(res.Success, Equals, true)

	// Fetch the parent
	newParent3 := &Parent{}
	collection.FindById(parent2.Id, newParent3)
	c.Assert(newParent3.Child.SubChild.Foo, Equals, "MySubChild")
	c.Assert(newParent3.Child.SubChild.Id.Hex(), Equals, subChild.Id.Hex())

	newParent4 := &Parent{}
	err := childCollection.Delete(child)
	c.Assert(err, Equals, nil)
	collection.FindById(parent2.Id, newParent4)
	c.Assert(newParent4.Child.Name, Equals, "")
	c.Assert(newParent4.ChildProp, Equals, "")
	c.Assert(len(newParent4.Children), Equals, 0)

}
