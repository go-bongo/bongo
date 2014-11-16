package frat

import (
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"maxwellhealth/common/crypt"
	"github.com/oleiade/reflections"
	"fmt"
)
type Collection struct {
	Collection *mgo.Collection

}

// var key = []byte("asdf1234asdf1234")


// Save a model struct. Struct must have "Id" property. If that property is not a valid ObjectId then a new one is created.
func (c *Collection) Save(mod interface{}) (error) {

	// 1) Make sure mod has an Id field
	has, _ := reflections.HasField(mod, "Id")
	if !has {
		panic("Failed to save - model must have Id field")
	}

	// 2) If there's no ID, create a new one
	
	f, err := reflections.GetField(mod, "Id")
	id := f.(bson.ObjectId)

	if err != nil {
		panic(err)
	}
	
	if !id.Valid() {
		fmt.Println("Creating new ID")
		id := bson.NewObjectId()
		err := reflections.SetField(mod, "Id", id)  // err != nil

		if err != nil {
			panic(err)
		}
	}
	


	// 2) Convert the model into a map using the crypt library
	modelMap := crypt.EncryptDocument(key, mod)
	err =  c.Collection.Insert(modelMap)

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