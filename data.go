package bongo

import (
	"encoding/json"
	// "errors"
	// "fmt"
	// "github.com/fatih/structs"
	// "github.com/mitchellh/mapstructure"
	"github.com/oleiade/reflections"
	// "labix.org/v2/mgo/bson"
	"log"
	"reflect"
	"strings"
)

// Encrypts fields on a document
func PrepDocumentForSave(key []byte, doc interface{}) map[string]interface{} {
	// defer func() {
	// 	if r := recover(); r != nil {
	// 		// return doc
	// 	}
	// }()

	returnMap := make(map[string]interface{})

	fields, _ := reflections.Tags(doc, "bson")

	var bsonName string
	for fieldName, bsonTag := range fields {
		bsonName = strings.Split(bsonTag, ",")[0]
		if len(bsonName) == 0 {
			bsonName = strings.ToLower(fieldName)
		}

		tag, _ := reflections.GetFieldTag(doc, fieldName, "bongo")
		bongoTags := getBongoTags(tag)
		val, _ := reflections.GetField(doc, fieldName)

		// Special types: bson.ObjectId, []bson.ObjectId,
		if bongoTags.encrypted {
			bytes, err := json.Marshal(val)
			if err != nil {
				panic(err)
			}
			encrypted, err := Encrypt(key, bytes)

			if err != nil {
				panic(err)
			}

			returnMap[bsonName] = encrypted
		} else {
			t := reflect.TypeOf(val)
			rval := reflect.ValueOf(val)
			// May need to iterate over sub documents with their own bson/encryption settings
			if shouldRecurse(t) {
				// Is it a time? Allow it through if so.
				// if string(f.Type().Name()) == "Time" {
				// 	returnMap[bsonName] = structs.Map(f.Interface())
				// } else {
				// 	// iterate

				// Recurse only if not nil
				if t.Kind() == reflect.Struct || !rval.IsNil() {
					log.Println("Recursing", bsonName, val)
					returnMap[bsonName] = PrepDocumentForSave(key, val)
				}
				// }

				// } else if id, ok := f.Interface().(bson.ObjectId); ok {

				// 	// Skip invalid objectIds - these should be validated if they're needed, but otherwise they should just be nil
				// 	if id.Valid() {
				// 		returnMap[bsonName] = id
				// 	} else {
				// 		returnMap[bsonName] = nil
				// 	}
			} else {
				returnMap[bsonName] = val
			}
		}
	}

	// log.Println(returnMap)

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

func shouldRecurse(t reflect.Type) bool {
	if (t.Kind() == reflect.Struct || t.Kind() == reflect.Ptr) && !strings.HasSuffix(t.String(), "bson.ObjectId") && !strings.HasSuffix(t.String(), "time.Time") {
		return true
	}
	return false
}

type bongoTags struct {
	encrypted bool
	index     bool
	unique    bool
	cascaded  bool
}

func getBongoTags(tag string) *bongoTags {
	ret := &bongoTags{false, false, false, false}

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

	if stringInSlice("cascaded", tags) {
		ret.cascaded = true
	}

	return ret

}

// func stringInSlice(a string, list []string) bool {
// 	for _, b := range list {
// 		if b == a {
// 			return true
// 		}
// 	}
// 	return false
// }

// Decrypt a struct. Use tag `encrypted="true"` to designate fields as needing to be decrypted
func InitializeDocumentFromDB(key []byte, encrypted map[string]interface{}, doc interface{}) {
	decoderHook := func(data interface{}, to reflect.Value, currentField *reflect.StructField) (interface{}, error) {
		// dataVal := reflect.ValueOf(data)

		if len(currentField.Tag) > 0 {
			// Check bongo fields
			bongoConfig := getBongoTags(currentField.Tag.Get("bongo"))
			// log.Println("Decoding", dataVal, to)
			if bongoConfig.encrypted {
				// Decrypt it
				if str, ok := data.(string); ok {
					decrypted, err := Decrypt(key, str)
					// log.Println("Decrypting", str)
					if err != nil {
						panic(err)
					}

					newVal := reflect.New(to.Type()).Interface()

					// Special case for object ID since it'll whine if it's not set
					if strings.HasSuffix(to.Type().String(), "ObjectId") && string(decrypted) == "\"\"" {
						return "", nil
					}
					// log.Println(reflect.TypeOf(newVal))
					err = json.Unmarshal(decrypted, newVal)
					if err != nil {
						panic(err)
					}

					if !reflect.ValueOf(newVal).IsValid() {
						// log.Println(newVal, "isn't valid")
						return data, nil
					}

					value := reflectValue(newVal).Interface()
					// log.Println("Decrypted into", value)
					return value, nil

				}

			}
		}

		return data, nil
	}

	// New decoder using the bson mapping
	decoderConfig := &DecoderConfig{
		TagName:    "bson",
		Result:     doc,
		DecodeHook: decoderHook,
	}

	decoder, err := NewDecoder(decoderConfig)

	// Decode the decrypted map into the doc, then set the other fields on the doc
	err = decoder.Decode(encrypted)

	if err != nil {
		panic(err)
	}

}
