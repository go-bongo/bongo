package bongo

import (
	"encoding/json"
	"github.com/fatih/structs"
	"github.com/mitchellh/mapstructure"
	"labix.org/v2/mgo/bson"
	"log"
	"reflect"
	"strings"
)

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

		// encrypt := stringInSlice(fieldName, encryptedFields)
		encrypt := typeOfT.Field(i).Tag.Get("encrypted") == "true"
		var bsonName string
		bsonName = typeOfT.Field(i).Tag.Get("bson")
		if len(bsonName) == 0 {
			bsonName = strings.ToLower(fieldName)
		}

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

// Decrypt a struct. Use tag `encrypted="true"` to designate fields as needing to be decrypted
func InitializeDocumentFromDB(key []byte, encrypted map[string]interface{}, doc interface{}) {

	decryptedMap := make(map[string]interface{})

	defer func() {
		if r := recover(); r != nil {
			log.Fatal("Error matching decrypted value to struct: \n", r)
		}
	}()
	s := reflect.ValueOf(doc).Elem()
	typeOfT := s.Type()

	for i := 0; i < s.NumField(); i++ {
		fieldName := string(typeOfT.Field(i).Name)
		f := s.Field(i)

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

					switch f.Kind() {
					case reflect.String:
						var str string
						json.Unmarshal(decrypted, &str)
						decryptedMap[fieldName] = str
					case reflect.Int, reflect.Int64:
						var n int64
						json.Unmarshal(decrypted, &n)
						decryptedMap[fieldName] = n
					case reflect.Float64, reflect.Float32:
						var f float64
						json.Unmarshal(decrypted, &f)
						decryptedMap[fieldName] = f
					case reflect.Bool:
						var b bool
						json.Unmarshal(decrypted, &b)
						decryptedMap[fieldName] = b
					case reflect.Struct:
						// Convert it to a map
						var m map[string]interface{}
						json.Unmarshal(decrypted, &m)
						decryptedMap[fieldName] = m
					case reflect.Slice:
						var a []interface{}
						json.Unmarshal(decrypted, &a)
						decryptedMap[fieldName] = a
					}

				} else {
					panic("not a string")
				}
			} else {
				decryptedMap[fieldName] = encrypted[bsonName]
			}
		}
	}

	err := mapstructure.Decode(decryptedMap, doc)

	if err != nil {
		panic(err)
	}
}
