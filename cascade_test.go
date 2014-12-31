package bongo

import (
	. "gopkg.in/check.v1"
	"labix.org/v2/mgo/bson"
	"log"
)

type Parent struct {
	Id     bson.ObjectId `bson:"_id"`
	Name   string
	Number int
}

func (p *Parent) GetCascade(encryptedData map[string]interface{}) []*CascadeConfig {
	return []*CascadeConfig{
		&CascadeConfig{
			Collection:  connection.Collection("children").Collection(),
			Properties:  []string{"name"},
			ThroughProp: "parent",
			RelType:     REL_ONE,
			Query: bson.M{
				"parentid": p.Id,
			},
		},
	}
}

type Child struct {
	Id       bson.ObjectId `bson:"_id"`
	ParentId bson.ObjectId
	Parent   *ParentRef
}

type ParentRef struct {
	Id   bson.ObjectId `bson:"_id,omitempty"`
	Name string
}

type MormonChild struct {
	Id        bson.ObjectId
	ParentIds []bson.ObjectId
	Parents   []*ParentRef
}

func (s *TestSuite) TestCascadeSingle(c *C) {
	collection := connection.Collection("parents")

	childCollection := connection.Collection("children")

	parent := &Parent{
		Name:   "Testy McGee",
		Number: 5,
	}

	res := collection.Save(parent)

	c.Assert(res.Success, Equals, true)

	child := &Child{
		ParentId: parent.Id,
	}

	res = childCollection.Save(child)
	c.Assert(res.Success, Equals, true)

	// Cascade the parent
	prepped := PrepDocumentForSave(key, parent)

	Cascade(parent, prepped)

	// Get the child
	// newChild := &Child{
	// 	Parent: &ParentRef{},
	// }
	//
	newChild := &Child{
		Parent: &ParentRef{},
	}
	childCollection.Find(nil).Next(newChild)
	// err := childCollection.FindById(child.Id, newChild)

	log.Println("Got parent:", newChild.Parent)
	c.Assert(newChild.Parent.Name, Equals, "Testy McGee")
	c.Assert(newChild.Parent.Id.Hex(), Equals, newChild.ParentId.Hex())
	// log.Println(newChild)

}
