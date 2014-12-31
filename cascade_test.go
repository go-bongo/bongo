package bongo

import (
	. "gopkg.in/check.v1"
	"labix.org/v2/mgo/bson"
	"log"
)

type Parent struct {
	Id       bson.ObjectId `bson:"_id"`
	Name     string
	Number   int
	Children []*ChildRef
	Child    *ChildRef
}

var checker = NewChangeChecker()

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

	if checker.Modified(c.Id, "ParentId", c) {
		log.Println("Modified parentid")
		origId, _ := checker.GetOriginalValue(c.Id, "ParentId")
		oldQuery := bson.M{
			"_id": origId,
		}
		cascadeSingle.OldQuery = oldQuery
		cascadeMulti.OldQuery = oldQuery
	}

	return []*CascadeConfig{cascadeSingle, cascadeMulti}
}

type Child struct {
	Id       bson.ObjectId `bson:"_id"`
	ParentId bson.ObjectId
	Name     string
}

type ChildRef struct {
	Id   bson.ObjectId `bson:"_id,omitempty"`
	Name string
}

func (s *TestSuite) TestCascade(c *C) {

	collection := connection.Collection("parents")

	childCollection := connection.Collection("children")

	parent := &Parent{
		Name:   "Testy McGee",
		Number: 5,
	}

	parent2 := &Parent{
		Name:   "Other Parent",
		Number: 10,
	}

	res := collection.Save(parent)

	c.Assert(res.Success, Equals, true)
	checker.StoreOriginal(parent.Id, parent)
	res = collection.Save(parent2)
	c.Assert(res.Success, Equals, true)
	checker.StoreOriginal(parent2.Id, parent2)

	child := &Child{
		ParentId: parent.Id,
		Name:     "Foo McGoo",
	}

	res = childCollection.Save(child)
	c.Assert(res.Success, Equals, true)

	checker.StoreOriginal(child.Id, child)

	// Cascade the parent
	prepped := PrepDocumentForSave(key, child)

	Cascade(child, prepped)

	// Get the child
	// newChild := &Child{
	// 	Parent: &ParentRef{},
	// }
	//
	newParent := &Parent{}
	collection.FindById(parent.Id, newParent)
	c.Assert(newParent.Child.Name, Equals, "Foo McGoo")
	c.Assert(newParent.Child.Id.Hex(), Equals, child.Id.Hex())
	c.Assert(newParent.Children[0].Name, Equals, "Foo McGoo")
	c.Assert(newParent.Children[0].Id.Hex(), Equals, child.Id.Hex())

	// Now change the child parent Id...
	child.ParentId = parent2.Id
	res = childCollection.Save(child)
	c.Assert(res.Success, Equals, true)
	prepped = PrepDocumentForSave(key, child)
	Cascade(child, prepped)

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
	// c.Assert(newParent.Child.Id.Hex(), Equals, child.Id.Hex())
	// c.Assert(newParent.Children[0].Name, Equals, "Foo McGoo")
	// c.Assert(newParent.Children[0].Id.Hex(), Equals, child.Id.Hex())

}
