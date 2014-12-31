package bongo

import (
	"errors"
	"fmt"
	"github.com/maxwellhealth/dotaccess"
	"labix.org/v2/mgo/bson"
	"reflect"
	"strings"
)

type ChangeChecker struct {
	originals map[bson.ObjectId]interface{}
}

func NewChangeChecker() *ChangeChecker {
	c := &ChangeChecker{}

	c.originals = make(map[bson.ObjectId]interface{})

	return c
}

func (c *ChangeChecker) StoreOriginal(id bson.ObjectId, doc interface{}) {
	// Store a copy
	c.originals[id] = reflect.Indirect(reflect.ValueOf(doc)).Interface()
}

func (c *ChangeChecker) Modified(id bson.ObjectId, field string, newDoc interface{}) bool {
	isNew, diffs, _ := c.Compare(id, newDoc)

	if isNew {
		return true
	} else {
		return stringInSlice(field, diffs)
	}
}

func (c *ChangeChecker) GetOriginalValue(id bson.ObjectId, field string) (interface{}, error) {
	if orig, ok := c.originals[id]; ok {
		return dotaccess.Get(orig, field)
	} else {
		return nil, nil
	}
}

func (c *ChangeChecker) Clear() {
	c.originals = make(map[bson.ObjectId]interface{})
}

func (c *ChangeChecker) Compare(id bson.ObjectId, newDoc interface{}) (bool, []string, error) {
	if orig, ok := c.originals[id]; ok {
		diffs, err := getChangedFields(orig, newDoc)
		return false, diffs, err
	} else {
		return true, []string{}, nil
	}
}

func getChangedFields(struct1 interface{}, struct2 interface{}) ([]string, error) {

	diffs := make([]string, 0)
	val1 := reflect.ValueOf(struct1)
	type1 := val1.Type()

	val2 := reflect.ValueOf(struct2)
	type2 := val2.Type()

	if type1.Kind() == reflect.Ptr {
		type1 = type1.Elem()
		val1 = val1.Elem()
	}
	if type2.Kind() == reflect.Ptr {
		type2 = type2.Elem()
		val2 = val2.Elem()
	}

	if type1.String() != type2.String() {
		return diffs, errors.New(fmt.Sprintf("Cannot compare two structs of different types %s and %s", type1.String(), type2.String()))
	}

	if type1.Kind() != reflect.Struct || type2.Kind() != reflect.Struct {
		return diffs, errors.New(fmt.Sprintf("Can only compare two structs or two pointers to structs", type1.Kind(), type2.Kind()))
	}

	for i := 0; i < type1.NumField(); i++ {
		field1 := val1.Field(i)
		field2 := val2.Field(i)

		field := type1.Field(i)
		fieldName := field.Name

		childType := field1.Type()
		// Recurse?
		if childType.Kind() == reflect.Ptr {
			childType = childType.Elem()
		}

		if childType.Kind() == reflect.Struct {
			childDiffs, err := getChangedFields(field1.Interface(), field2.Interface())

			if err != nil {
				return diffs, err
			}

			if len(childDiffs) > 0 {
				for _, diff := range childDiffs {
					diffs = append(diffs, strings.Join([]string{fieldName, diff}, "."))
				}
			}
		} else {
			fmt.Println("Comparing", field1.Interface(), "to", field2.Interface())
			if field1.Interface() != field2.Interface() {
				diffs = append(diffs, fieldName)
			}
		}
	}

	return diffs, nil

}
