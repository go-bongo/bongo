package bongo

import (
	. "gopkg.in/check.v1"
	// "labix.org/v2/mgo/bson"
	"encoding/json"
	// "fmt"
)

func (s *TestSuite) TestSaveProduct(c *C) {
	connection := Connect(config)
	defer connection.Session.Close()

	message := NewProduct()

	_, _ = json.MarshalIndent(message, "", "\t")
	// fmt.Println(string(marshaled))
	// res := connection.Save(message)

	// log.Println(res.Success)

}
