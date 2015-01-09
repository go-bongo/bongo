package bongo

import (
	"errors"
	"fmt"
	"github.com/maxwellhealth/dotaccess"
	// "labix.org/v2/mgo/bson"
	// "log"
	"reflect"
	"strings"
)

type DiffTracker struct {
	original interface{}
	current  interface{}
}

func NewDiffTracker(doc interface{}) *DiffTracker {
	c := &DiffTracker{
		current:  doc,
		original: nil,
	}

	return c
}

func (c *DiffTracker) Reset() {
	// Store a copy of current
	c.original = reflect.Indirect(reflect.ValueOf(c.current)).Interface()
}

func (c *DiffTracker) Modified(field string) bool {
	isNew, diffs, _ := c.Compare()

	if isNew {
		return true
	} else {
		return stringInSlice(field, diffs)
	}
}

func (c *DiffTracker) GetModified() (bool, []string) {
	isNew, diffs, _ := c.Compare()

	return isNew, diffs
}

func (c *DiffTracker) GetOriginalValue(field string) (interface{}, error) {
	if c.original != nil {
		return dotaccess.Get(c.original, field)
	}
	return nil, nil

}

func (c *DiffTracker) Clear() {
	c.original = nil
}

func (c *DiffTracker) Compare() (bool, []string, error) {
	defer func() {

		if r := recover(); r != nil {
			fmt.Println("You probably forgot to initialize the DiffTracker instance on your model")
			panic(r)
		}
	}()
	if c.original != nil {
		diffs, err := getChangedFields(c.original, c.current)
		return false, diffs, err
	} else {
		return true, []string{}, nil
	}
}

func getFields(t reflect.Type) []string {
	fields := []string{}

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fields = append(fields, field.Name)
	}

	return fields

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

		// Skip if not exported
		if len(field.PkgPath) > 0 || field.Name == "DiffTracker" {
			continue
		}

		if childType.Kind() == reflect.Struct {
			var childDiffs []string
			var err error
			// Make sure they aren't zero-value
			if field1.IsNil() && field2.IsNil() {
				return diffs, nil
			} else if field2.IsNil() || field1.IsNil() {
				childDiffs = getFields(childType)

			} else {
				childDiffs, err = getChangedFields(field1.Interface(), field2.Interface())

				if err != nil {
					return diffs, err
				}
			}

			if len(childDiffs) > 0 {
				for _, diff := range childDiffs {
					diffs = append(diffs, strings.Join([]string{fieldName, diff}, "."))
				}
			}
		} else {
			if field1.Interface() != field2.Interface() {
				diffs = append(diffs, fieldName)
			}
		}
	}

	return diffs, nil

}
