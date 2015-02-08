package bongo

import (
	// "encoding/json"
	"github.com/maxwellhealth/mgo/bson"
	"github.com/oleiade/reflections"
	"reflect"
	"strings"
)

// Encrypts fields on a document
func (c *Collection) PrepDocumentForSave(doc interface{}) map[string]interface{} {
	// defer func() {
	// 	if r := recover(); r != nil {
	// 		// return doc
	// 	}
	// }()

	returnMap := make(map[string]interface{})

	fields, _ := reflections.Tags(doc, "bson")

	var bsonName string
	var bsons []string
	for fieldName, bsonTag := range fields {
		bsons = strings.Split(bsonTag, ",")
		bsonName = bsons[0]

		if bsonName == "-" {
			continue
		}
		if len(bsonName) == 0 {
			left := string(fieldName[0])
			rest := string(fieldName[1:])

			bsonName = strings.Join([]string{strings.ToLower(left), rest}, "")
		}

		// Check if it's bson inline
		inline := false
		for _, t := range bsons {
			if t == "inline" {
				inline = true
				break
			}
		}

		tag, _ := reflections.GetFieldTag(doc, fieldName, "bongo")
		bongoTags := getBongoTags(tag)
		val, _ := reflections.GetField(doc, fieldName)

		// Skip if it's populated via cascade
		if len(bongoTags.cascadedFrom) > 0 {
			continue
		}

		t := reflect.TypeOf(val)
		rval := reflect.ValueOf(val)
		// May need to iterate over sub documents with their own bson/encryption settings. It won't be a separate encryption key since it's not cascaded (that will be skipped above if bongoTags.cascaded)
		if rval.IsValid() {
			if shouldRecurse(t, rval) {

				// Recurse only if not nil
				if t.Kind() == reflect.Struct || !rval.IsNil() {
					rec := c.PrepDocumentForSave(val)
					if inline {
						for k, v := range rec {
							returnMap[k] = v
						}
					} else {
						returnMap[bsonName] = c.PrepDocumentForSave(val)
					}

				}
			} else if t.String() == "bson.ObjectId" {
				// We won't catch "omitempty"
				if idVal, ok := val.(bson.ObjectId); ok {
					if idVal.Valid() {
						returnMap[bsonName] = idVal
					}
				}
			} else {
				returnMap[bsonName] = val
			}
		}

	}

	return returnMap
}

func reflectValue(obj interface{}) reflect.Value {
	var val reflect.Value

	if reflect.TypeOf(obj).Kind() == reflect.Ptr {
		val = reflect.ValueOf(obj).Elem()
	} else {
		val = reflect.ValueOf(obj)
	}

	return val
}

type setValue struct {
	fieldName string
	value     interface{}
}

func shouldRecurse(t reflect.Type, v reflect.Value) bool {
	if (t.Kind() == reflect.Struct || t.Kind() == reflect.Ptr) && !strings.HasSuffix(t.String(), "bson.ObjectId") && !strings.HasSuffix(t.String(), "time.Time") {
		return true
	}
	return false
}

type bongoTags struct {
	encrypted    bool
	index        bool
	unique       bool
	cascadedFrom string
}

func getBongoTags(tag string) *bongoTags {
	ret := &bongoTags{false, false, false, ""}

	tags := strings.Split(tag, ",")

	if stringInSlice("encrypted", tags) {
		ret.encrypted = true
	}

	if stringInSlice("index", tags) {
		ret.index = true
	}

	if stringInSlice("unique", tags) {
		ret.unique = true
	}

	// Check for cascadedFrom so we know how to decrypt
	for _, t := range tags {
		if strings.HasPrefix(t, "cascadedFrom=") {
			ret.cascadedFrom = strings.TrimPrefix(t, "cascadedFrom=")
			break
		}
	}
	return ret

}
