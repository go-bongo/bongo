package bongo

import (
	"encoding/json"
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

	key := c.GetEncryptionKey()

	var bsonName string
	for fieldName, bsonTag := range fields {
		bsonName = strings.Split(bsonTag, ",")[0]

		if bsonName == "-" {
			continue
		}
		if len(bsonName) == 0 {
			bsonName = strings.ToLower(fieldName)
		}

		tag, _ := reflections.GetFieldTag(doc, fieldName, "bongo")
		bongoTags := getBongoTags(tag)
		val, _ := reflections.GetField(doc, fieldName)

		// Skip if it's populated via cascade
		if len(bongoTags.cascadedFrom) > 0 {
			continue
		}
		// Special types: bson.ObjectId, []bson.ObjectId,
		if bongoTags.encrypted && !c.Connection.Config.DisableEncryption {
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
			// May need to iterate over sub documents with their own bson/encryption settings. It won't be a separate encryption key since it's not cascaded (that will be skipped above if bongoTags.cascaded)
			if shouldRecurse(t) {

				// Recurse only if not nil
				if t.Kind() == reflect.Struct || !rval.IsNil() {
					returnMap[bsonName] = c.PrepDocumentForSave(val)
				}
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

// func stringInSlice(a string, list []string) bool {
// 	for _, b := range list {
// 		if b == a {
// 			return true
// 		}
// 	}
// 	return false
// }

// Decrypt a struct. Use tag `encrypted="true"` to designate fields as needing to be decrypted
func (c *Collection) InitializeDocumentFromDB(encrypted map[string]interface{}, doc interface{}) {

	decoderHook := func(data interface{}, to reflect.Value, decoder *Decoder) (interface{}, error) {
		if c.Connection.Config.DisableEncryption {
			return data, nil
		}
		// If we're inside an encrypted prop, that means it's already been json-decoded and is just being marshaled into its final value. In that case we don't even care, so just let it go through (you can't have nested encryption)
		if decoder.InEncryptedProp {
			return data, nil
		}
		currentField := decoder.CurrentField

		// log.Println("Current field:", currentField)
		if len(currentField.Tag) > 0 {

			colName := c.Name
			if len(decoder.CascadedFrom) > 0 {
				colName = decoder.CascadedFrom
			}

			// Check bongo fields
			bongoConfig := getBongoTags(currentField.Tag.Get("bongo"))
			// log.Println("Decoding", dataVal, to)
			if bongoConfig.encrypted {
				// Decrypt it
				key := c.Connection.GetEncryptionKey(colName)
				// key := c.Connection.GetEncryptionKey("asdf1234asdf1234")

				if str, ok := data.(string); ok {
					decrypted, err := Decrypt(key, str)
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
