package bongo

import (
	"gopkg.in/mgo.v2/bson"
	"reflect"
)

func ValidateRequired(val interface{}) bool {
	valueOf := reflect.ValueOf(val)
	return valueOf.Interface() != reflect.Zero(valueOf.Type()).Interface()
}

func ValidateMongoIdRef(id bson.ObjectId, collection *Collection) bool {
	count, err := collection.Collection().Find(bson.M{"_id": id}).Count()

	if err != nil {
		return false
	}

	if count > 0 {
		return true
	}
	return false
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func ValidateInclusionIn(value string, options []string) bool {
	return stringInSlice(value, options)
}
