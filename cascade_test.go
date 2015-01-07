package bongo

import (
	. "gopkg.in/check.v1"
	"labix.org/v2/mgo/bson"
	// "log"
)

type Parent struct {
	Id          bson.ObjectId `bson:"_id"`
	Name        string        `bongo:"encrypted"`
	Number      int
	Children    []*ChildRef  `bongo:"cascadedFrom=children"`
	Child       *ChildRef    `bongo:"cascadedFrom=children"`
	DiffTracker *DiffTracker `bson:"-" json:"-"`
}

func (c *Child) GetCascade() []*CascadeConfig {

	cascadeSingle := &CascadeConfig{
		Collection:  connection.Collection("parents").Collection(),
		Properties:  []string{"name"},
		ThroughProp: "child",
		RelType:     REL_ONE,
		Query: bson.M{
			"_id": c.ParentId,
		},
	}

	cascadeMulti := &CascadeConfig{
		Collection:  connection.Collection("parents").Collection(),
		Properties:  []string{"name"},
		ThroughProp: "children",
		RelType:     REL_MANY,
		Query: bson.M{
			"_id": c.ParentId,
		},
	}

	if c.DiffTracker.Modified("ParentId") {

		origId, _ := c.DiffTracker.GetOriginalValue("ParentId")
		if origId != nil {
			oldQuery := bson.M{
				"_id": origId,
			}
			cascadeSingle.OldQuery = oldQuery
			cascadeMulti.OldQuery = oldQuery
		}

	}

	return []*CascadeConfig{cascadeSingle, cascadeMulti}
}

type Child struct {
	Id       bson.ObjectId `bson:"_id"`
	ParentId bson.ObjectId
	Name     string `bongo:"encrypted"`
	// System will automatically instantate the tracker
	DiffTracker *DiffTracker `bson:"-" json:"-"`
}

type ChildRef struct {
	Id   bson.ObjectId `bson:"_id,omitempty"`
	Name string        `bongo:"encrypted"`
}

func (s *TestSuite) TestCascade(c *C) {

	collection := connection.Collection("parents")

	connection.Config.EncryptionKeyPerCollection = map[string][]byte{
		"parents":  []byte("asdf1234asdf1234"),
		"children": []byte("1234asdf1234asdf"),
	}

	childCollection := connection.Collection("children")

	parent := &Parent{
		Name:   "Testy McGee",
		Number: 5,
	}

	parent.DiffTracker = NewDiffTracker(parent)

	parent2 := &Parent{
		Name:   "Other Parent",
		Number: 10,
	}
	parent2.DiffTracker = NewDiffTracker(parent2)

	res := collection.Save(parent)

	c.Assert(res.Success, Equals, true)
	res = collection.Save(parent2)
	c.Assert(res.Success, Equals, true)

	child := &Child{
		ParentId: parent.Id,
		Name:     "Foo McGoo",
	}

	child.DiffTracker = NewDiffTracker(child)

	res = childCollection.Save(child)
	c.Assert(res.Success, Equals, true)

	newParent := &Parent{}
	collection.FindById(parent.Id, newParent)
	c.Assert(newParent.Child.Name, Equals, "Foo McGoo")
	c.Assert(newParent.Child.Id.Hex(), Equals, child.Id.Hex())
	c.Assert(newParent.Children[0].Name, Equals, "Foo McGoo")
	c.Assert(newParent.Children[0].Id.Hex(), Equals, child.Id.Hex())

	// Now change the child parent Id...
	child.ParentId = parent2.Id
	c.Assert(child.DiffTracker.Modified("ParentId"), Equals, true)

	res = childCollection.Save(child)
	c.Assert(res.Success, Equals, true)
	// Now make sure it says the parent id DIDNT change, because we just saved it
	c.Assert(child.DiffTracker.Modified("ParentId"), Equals, false)

	newParent1 := &Parent{}
	collection.FindById(parent.Id, newParent1)
	c.Assert(newParent1.Child, IsNil)
	c.Assert(len(newParent1.Children), Equals, 0)
	newParent2 := &Parent{}
	collection.FindById(parent2.Id, newParent2)
	c.Assert(newParent2.Child.Name, Equals, "Foo McGoo")
	c.Assert(newParent2.Child.Id.Hex(), Equals, child.Id.Hex())
	c.Assert(newParent2.Children[0].Name, Equals, "Foo McGoo")
	c.Assert(newParent2.Children[0].Id.Hex(), Equals, child.Id.Hex())

	newParent3 := &Parent{}
	err := childCollection.Delete(child)
	c.Assert(err, Equals, nil)
	collection.FindById(parent2.Id, newParent3)
	c.Assert(newParent3.Child, IsNil)
	c.Assert(len(newParent3.Children), Equals, 0)

}
