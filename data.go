package bongo

import (
	"encoding/json"
	// "errors"
	// "fmt"
	"github.com/fatih/structs"
	"github.com/mitchellh/mapstructure"
	"github.com/oleiade/reflections"
	"labix.org/v2/mgo/bson"
	// "log"
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

	v := reflect.ValueOf(doc)

	var s reflect.Value

	if v.Kind() == reflect.Ptr {
		s = v.Elem()
	} else {
		s = v
	}

	// s := reflect.ValueOf(doc).Elem()
	typeOfT := s.Type()

	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)
		fieldName := typeOfT.Field(i).Name
		// fieldType := f.Type().String()

		// var iface interface{}
		// encrypt := stringInSlice(fieldName, encryptedFields)
		encrypt := typeOfT.Field(i).Tag.Get("encrypted") == "true"
		var bsonName string
		bsonName = typeOfT.Field(i).Tag.Get("bson")
		if len(bsonName) == 0 {
			bsonName = strings.ToLower(fieldName)
		}

		// Special types: bson.ObjectId, []bson.ObjectId,
		if encrypt {
			bytes, err := json.Marshal(f.Interface())
			if err != nil {
				panic(err)
			}
			encrypted, err := Encrypt(key, bytes)

			if err != nil {
				panic(err)
			}

			returnMap[bsonName] = encrypted
		} else {
			// May need to iterate over sub documents with their own bson/encryption settings
			if f.Kind() == reflect.Struct {
				// Is it a time? Allow it through if so.
				if string(f.Type().Name()) == "Time" {
					returnMap[bsonName] = structs.Map(f.Interface())
				} else {
					// iterate
					returnMap[bsonName] = PrepDocumentForSave(key, f.Interface())
				}

			} else if id, ok := f.Interface().(bson.ObjectId); ok {

				// Skip invalid objectIds - these should be validated if they're needed, but otherwise they should just be nil
				if id.Valid() {
					returnMap[bsonName] = id
				} else {
					returnMap[bsonName] = nil
				}
			} else {
				returnMap[bsonName] = f.Interface()
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

// Decrypt a struct. Use tag `encrypted="true"` to designate fields as needing to be decrypted
func InitializeDocumentFromDB(key []byte, encrypted map[string]interface{}, doc interface{}) {

	decryptedMap := make(map[string]interface{})

	setValues := []setValue{}

	// defer func() {
	// 	if r := recover(); r != nil {
	// 		log.Fatal("Error matching decrypted value to struct: \n", r)
	// 	}
	// }()
	s := reflect.ValueOf(doc).Elem()
	typeOfT := s.Type()

	for i := 0; i < s.NumField(); i++ {

		fieldName := string(typeOfT.Field(i).Name)
		// f := s.Field(i)

		var bsonName string
		bsonName = typeOfT.Field(i).Tag.Get("bson")
		if len(bsonName) == 0 {
			bsonName = strings.ToLower(fieldName)
		}
		_, hasField := encrypted[bsonName]

		if hasField {
			decrypt := typeOfT.Field(i).Tag.Get("encrypted") == "true"
			var decrypted []byte
			var err error
			if decrypt {
				if str, ok := encrypted[bsonName].(string); ok {
					decrypted, err = Decrypt(key, str)
					if err != nil {
						panic(err)
					}

					// If decrypted is null, leave it at zero value
					if string(decrypted) == "null" {
						continue
					}

					iface, _ := reflections.GetField(doc, fieldName)
					origType := reflect.TypeOf(iface)

					// json.Unmarshal uses map[string]interface{} unless we create a new instance of this type
					newType := reflect.New(origType).Interface()
					// newType := iface
					//
					// bson.ObjectId whines when you're trying to marshal an empty string, so we'll skip those
					if origType.String() == "bson.ObjectId" && string(decrypted) == "\"\"" {
						continue
					}

					err = json.Unmarshal(decrypted, &newType)
					if err != nil {
						panic(err)
					}

					// Need to get the underlying value since reflect.New always gives a pointer

					value := reflectValue(newType).Interface()

					setValues = append(setValues, setValue{fieldName, value})
					// err = reflections.SetField(doc, typeOfT.Field(i).Name, value)

					if err != nil {
						panic(err)
					}

				} else {
					panic("not a string")
				}
			} else {

				// Only set if it's not the zero value and not nil

				decryptedMap[fieldName] = encrypted[bsonName]

			}
		}
	}

	// Decode the decrypted map into the doc, then set the other fields on teh doc
	err := mapstructure.Decode(decryptedMap, doc)

	if err != nil {
		panic(err)
	}

	for _, val := range setValues {
		err := reflections.SetField(doc, val.fieldName, val.value)

		if err != nil {
			panic(err)
		}
	}

}

// func decodeSlice(name string, data interface{}, val reflect.Value) error {
// 	dataVal := reflect.Indirect(reflect.ValueOf(data))
// 	dataValKind := dataVal.Kind()
// 	valType := val.Type()
// 	valElemType := valType.Elem()
// 	sliceType := reflect.SliceOf(valElemType)

// 	// Check input type
// 	if dataValKind != reflect.Array && dataValKind != reflect.Slice {
// 		// Accept empty map instead of array/slice in weakly typed mode
// 		if d.config.WeaklyTypedInput && dataVal.Kind() == reflect.Map && dataVal.Len() == 0 {
// 			val.Set(reflect.MakeSlice(sliceType, 0, 0))
// 			return nil
// 		} else {
// 			return fmt.Errorf(
// 				"'%s': source data must be an array or slice, got %s", name, dataValKind)
// 		}
// 	}

// 	// Make a new slice to hold our result, same size as the original data.
// 	valSlice := reflect.MakeSlice(sliceType, dataVal.Len(), dataVal.Len())

// 	// Accumulate any errors
// 	errors := make([]string, 0)

// 	for i := 0; i < dataVal.Len(); i++ {
// 		currentData := dataVal.Index(i).Interface()
// 		currentField := valSlice.Index(i)

// 		fieldName := fmt.Sprintf("%s[%d]", name, i)
// 		if err := d.decode(fieldName, currentData, currentField); err != nil {
// 			errors = appendErrors(errors, err)
// 		}
// 	}

// 	// Finally, set the value to the slice we built up
// 	val.Set(valSlice)

// 	// If there were errors, we return those
// 	if len(errors) > 0 {
// 		return &Error{errors}
// 	}

// 	return nil
// }
