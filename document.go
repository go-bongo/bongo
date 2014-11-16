package frat

import (
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"maxwellhealth/common/crypt"
	// "github.com/oleiade/reflections"
	"fmt"
)
type Document struct {
	Id bson.ObjectId
	Model interface{}
	Connection *MongoConnection
	Collection *mgo.Collection
}

// var key = []byte("asdf1234asdf1234")


// Save a document
func (d *Document) Save() (error) {

	// 1) If there's no ID, create a new one
	if !d.Id.Valid() {
		fmt.Println("Creating new ID")
		d.Id = bson.NewObjectId()
	}
	
	// 2) Convert the model into a map using the crypt library
	modelMap := crypt.EncryptDocument(key, d.Model)
	modelMap["_id"] = d.Id
	err :=  d.Collection.Insert(modelMap)

	return err
}

// type msg struct {
// 	Msg   string        `encrypted:"true",bson="msg"`
// 	Count int           `encrypted:"true",bson="count"`
// }

// func FindById(id bson.ObjectId, obj interface{}, c *mgo.Collection) (error) {

// 	fmt.Println(obj)
// 	s := reflect.ValueOf(obj)
// 	fmt.Println(s.Kind())
// 	fmt.Println(s.Type().Name())
// 	// doc := new(Document)

// 	resultMap := make(map[string]interface{})
// 	err := c.Find(bson.M{"_id": id}).One(&resultMap)
// 	// err := doc.Collection.FindId(doc.Id).One(&resultMap)


	
// 	if err != nil {
// 		fmt.Println("HAS ERROR")
// 		return err
// 	}

// 	fmt.Println(resultMap)

// 	// returnObj := new(msg)
// 	// Decrypt
// 	crypt.DecryptDocument(key, resultMap, obj)


// 	// doc.Model = obj


// 	return nil
// }