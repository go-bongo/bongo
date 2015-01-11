package bongo

import (
	"errors"
	"fmt"
	"github.com/oleiade/reflections"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"time"
	// "log"
	// "reflect"
	// "math"
	// "strings"
)

type Collection struct {
	Name       string
	Connection *Connection
}

func (c *Collection) GetEncryptionKey() []byte {
	return c.Connection.GetEncryptionKey(c.Name)
}

func (c *Collection) Collection() *mgo.Collection {
	return c.Connection.Session.DB(c.Connection.Config.Database).C(c.Name)
}

func (c *Collection) Save(mod interface{}) (result *SaveResult) {
	defer func() {

		if r := recover(); r != nil {
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
	if validator, ok := mod.(interface {
		Validate(*Collection) []string
	}); ok {
		errs := validator.Validate(c)

		if len(errs) > 0 {
			err := NewSaveResult(false, errors.New("Validation failed"))
			err.ValidationErrors = errs
			return err
		}
	}

	if isNew {
		if hook, ok := mod.(interface {
			BeforeCreate(*Collection)
		}); ok {
			hook.BeforeCreate(c)
		}
	} else if hook, ok := mod.(interface {
		BeforeUpdate(*Collection)
	}); ok {
		hook.BeforeUpdate(c)
	}

	if hook, ok := mod.(interface {
		BeforeSave(*Collection)
	}); ok {
		hook.BeforeSave(c)
	}

	// 3) Convert the model into a map using the crypt library
	modelMap := c.PrepDocumentForSave(mod)

	// Add created/modified time
	if isNew {
		modelMap["_created"] = time.Now()
	}
	modelMap["_modified"] = time.Now()

	// 4) Cascade?
	CascadeSave(c, mod, modelMap)

	// 5) Save (upsert)
	_, err = c.Collection().UpsertId(modelMap["_id"], modelMap)

	if err != nil {
		panic(err)
	}

	// 6) Run afterSave hooks
	if isNew {
		if hook, ok := mod.(interface {
			AfterCreate(*Collection)
		}); ok {
			hook.AfterCreate(c)
		}
	} else if hook, ok := mod.(interface {
		AfterUpdate(*Collection)
	}); ok {
		hook.AfterUpdate(c)
	}

	if hook, ok := mod.(interface {
		AfterSave(*Collection)
	}); ok {
		hook.AfterSave(c)
	}

	// Leave this to the user.
	// if trackable, ok := mod.(Trackable); ok {
	// 	tracker := trackable.GetDiffTracker()
	// 	tracker.Reset()
	// }

	return NewSaveResult(true, nil)
}

func (c *Collection) FindById(id bson.ObjectId, mod interface{}) error {
	returnMap := make(map[string]interface{})

	err := c.Collection().FindId(id).One(&returnMap)
	if err != nil {
		return err
	}

	// Decrypt + Marshal into map

	c.InitializeDocumentFromDB(returnMap, mod)

	if hook, ok := mod.(interface {
		AfterFind(*Collection)
	}); ok {
		hook.AfterFind(c)
	}
	return nil
}

// Pass in the sample just so we can get the collection name
func (c *Collection) Find(query interface{}) *ResultSet {

	// Count for testing
	q := c.Collection().Find(query)

	resultset := new(ResultSet)

	resultset.Query = q
	resultset.Collection = c

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
		BeforeDelete(*Collection)
	}); ok {
		hook.BeforeDelete(c)
	}

	err = c.Collection().Remove(bson.M{"_id": id})

	if err != nil {
		return err
	}

	CascadeDelete(c, mod)
	if hook, ok := mod.(interface {
		AfterDelete(*Collection)
	}); ok {
		hook.AfterDelete(c)
	}

	return nil

}
