package bongo

import (
	"reflect"
	"strings"
	"unicode"
)

//Lower cases first char of string
func lowerInitial(str string) string {
	for i, v := range str {
		return string(unicode.ToLower(v)) + str[i+1:]
	}
	return ""
}

func GetBsonName(field reflect.StructField) string {
	tag := field.Tag.Get("bson")
	tags := strings.Split(tag, ",")

	if len(tags[0]) > 0 {
		return tags[0]
	} else {
		return lowerInitial(field.Name)
	}

}
