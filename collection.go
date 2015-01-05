package bongo

import (
	"errors"
	"fmt"
	"github.com/oleiade/reflections"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"log"
	"reflect"
	// "math"
	// "strings"
)

type Collection struct {
	Name       string
	Connection *Connection
}

func (c *Collection) Collection() *mgo.Collection {
	return c.Connection.Session.DB(c.Connection.Config.Database).C(c.Name)
}

func (c *Collection) Save(mod interface{}) (result *SaveResult) {
	defer func() {

		if r := recover(); r != nil {
			// panic(r)
			// return
			if e, ok := r.(error); ok {
				result = NewSaveResult(false, e)
			} else if e, ok := r.(string); ok {
				result = NewSaveResult(false, errors.New(e))
			} else {
				result = NewSaveResult(false, errors.New(fmt.Sprint(r)))
			}

		}
	}()

	// 1) Make sure mod has an Id field
	ensureIdField(mod)

	// 2) If there's no ID, create a new one
	f, err := reflections.GetField(mod, "Id")
	id := f.(bson.ObjectId)

	if err != nil {
		panic(err)
	}

	isNew := false

	if !id.Valid() {
		id := bson.NewObjectId()
		err := reflections.SetField(mod, "Id", id)

		if err != nil {
			panic(err)
		}

		isNew = true

	}

	// Validate?
	if _, ok := mod.(interface {
		Validate() []string
	}); ok {
		results := reflect.ValueOf(mod).MethodByName("Validate").Call([]reflect.Value{})
		if errs, ok := results[0].Interface().([]string); ok {
			if len(errs) > 0 {
				err := NewSaveResult(false, errors.New("Validation failed"))
				err.ValidationErrors = errs
				return err
			}
		}
	}

	if isNew {
		if hook, ok := mod.(interface {
			BeforeCreate()
		}); ok {
			hook.BeforeCreate()
		}
	} else if hook, ok := mod.(interface {
		BeforeUpdate()
	}); ok {
		hook.BeforeUpdate()
	}

	if hook, ok := mod.(interface {
		BeforeSave()
	}); ok {
		hook.BeforeSave()
	}

	// 3) Convert the model into a map using the crypt library
	modelMap := PrepDocumentForSave(c.Connection.GetEncryptionKey(c.Name), mod)

	// 4) Cascade?
	CascadeSave(mod, modelMap)

	// 5) Save (upsert)
	_, err = c.Collection().UpsertId(modelMap["_id"], modelMap)

	if err != nil {
		panic(err)
	}

	// 6) Run afterSave hooks
	if isNew {
		if hook, ok := mod.(interface {
			AfterCreate()
		}); ok {
			hook.AfterCreate()
		}
	} else if hook, ok := mod.(interface {
		AfterUpdate()
	}); ok {
		hook.AfterUpdate()
	}

	if hook, ok := mod.(interface {
		AfterSave()
	}); ok {
		hook.AfterSave()
	}

	// 7) If the model has a DiffTracker property that is a DiffTracker, instantiate it and save the original model
	t := reflect.ValueOf(mod)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	field := t.FieldByName("DiffTracker")
	if field.IsValid() {
		if field.IsNil() {
			log.Println("DiffTracker is not instantiated, cannot use")
		} else if field.Type().String() != "*bongo.DiffTracker" {
			log.Println("DiffTracker is not pointer to instance of bongo.DiffTracker, cannot use")
		} else {
			if diffed, ok := field.Interface().(*DiffTracker); ok {
				diffed.Reset()
			}
		}
	}

	return NewSaveResult(true, nil)
}

func (c *Collection) FindById(id bson.ObjectId, mod interface{}) error {
	returnMap := make(map[string]interface{})

	err := c.Collection().FindId(id).One(&returnMap)
	if err != nil {
		return err
	}

	// Decrypt + Marshal into map

	InitializeDocumentFromDB(c.Connection.GetEncryptionKey(c.Name), returnMap, mod)

	if hook, ok := mod.(interface {
		AfterFind()
	}); ok {
		hook.AfterFind()
	}
	return nil
}

// Pass in the sample just so we can get the collection name
func (c *Collection) Find(query interface{}) *ResultSet {

	// Count for testing
	q := c.Collection().Find(query)

	resultset := new(ResultSet)

	resultset.Query = q
	resultset.Connection = c.Connection

	return resultset
}

func (c *Collection) FindOne(query interface{}, mod interface{}) error {
	// Now run a find
	results := c.Find(query)

	hasNext := results.Next(mod)

	if !hasNext {
		return errors.New("No results found")
	}

	return nil
}

func (c *Collection) Delete(mod interface{}) error {
	ensureIdField(mod)
	f, err := reflections.GetField(mod, "Id")
	if err != nil {
		return err
	}
	id := f.(bson.ObjectId)
	if hook, ok := mod.(interface {
		BeforeDelete()
	}); ok {
		hook.BeforeDelete()
	}

	err = c.Collection().Remove(bson.M{"_id": id})

	if err != nil {
		return err
	}

	CascadeDelete(mod)
	if hook, ok := mod.(interface {
		AfterDelete()
	}); ok {
		hook.AfterDelete()
	}

	return nil

}
