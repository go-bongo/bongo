// Two-way encryption
// @author Jason Raede <jason@maxwellhealth.com>

package bongo

import (
	"crypto/rand"
	"golang.org/x/crypto/nacl/secretbox"
	// "encoding/base64"
	"encoding/hex"
	"encoding/json"

	"errors"
	"github.com/maxwellhealth/mgo/bson"
	"strconv"
	"time"
	// "strings"
	// "fmt"

	// "log"
	// "github.com/oleiade/reflections"
	// "io"
)

// var encryptionKey = []byte("asdf1234asdf1234")

//** BYTE-LEVEL PRIMITIVE METHODS

func generateNonce() ([24]byte, error) {
	var nonce [24]byte

	_, err := rand.Reader.Read(nonce[:])

	return nonce, err
}

// Encrypt an array of bytes for storage in the database as a base64 encoded string
func Encrypt(key [32]byte, val []byte) (string, error) {

	output := []byte{}
	nonce, err := generateNonce()

	if err != nil {
		return "", err
	}

	output = secretbox.Seal(output, val, &nonce, &key)

	res := make([]byte, 24+len(output))

	copy(res[:24], nonce[:])
	copy(res[24:], output)

	result := hex.EncodeToString(res)
	return result, nil
}

// Decrypt a base64-encoded string retrieved from the database and return an array of bytes
func Decrypt(key [32]byte, encrypted string) ([]byte, error) {
	val, err := hex.DecodeString(encrypted)

	if err != nil {
		return nil, err
	}

	var nonce [24]byte
	copy(nonce[:], val[:24])

	toDecrypt := val[24:]

	var decrypted []byte
	var ok bool

	decrypted, ok = secretbox.Open(decrypted, toDecrypt, &nonce, &key)

	if !ok {
		return decrypted, errors.New("Failed to decrypt")
	}

	return decrypted, nil
}

//* ENCRYPTED TYPES */
type EncryptedString string

func (e EncryptedString) GetBSON() (interface{}, error) {
	if EnableEncryption {
		return Encrypt(EncryptionKey, []byte(string(e)))
	} else {
		return string(e), nil
	}

}
func (e *EncryptedString) SetBSON(raw bson.Raw) error {
	if EnableEncryption {
		var str string
		raw.Unmarshal(&str)
		// log.Println("Unmarshaled into", str)
		decrypted, err := Decrypt(EncryptionKey, str)
		if err != nil {
			return err
		}
		*e = EncryptedString(string(decrypted))
	} else {
		var s string
		raw.Unmarshal(&s)
		*e = EncryptedString(s)

	}
	return nil

}

type EncryptedInt int

func (e EncryptedInt) GetBSON() (interface{}, error) {
	if EnableEncryption {
		return Encrypt(EncryptionKey, []byte(strconv.Itoa(int(e))))
	} else {
		return int(e), nil
	}

}
func (e *EncryptedInt) SetBSON(raw bson.Raw) error {
	if EnableEncryption {
		var str string
		raw.Unmarshal(&str)
		decrypted, err := Decrypt(EncryptionKey, str)
		if err != nil {
			return err
		}

		intVal, err := strconv.Atoi(string(decrypted))
		if err != nil {
			return err
		}
		*e = EncryptedInt(intVal)

	} else {
		var i int
		raw.Unmarshal(&i)
		*e = EncryptedInt(i)
	}
	return nil
}

type EncryptedFloat float64

func (e EncryptedFloat) GetBSON() (interface{}, error) {

	if EnableEncryption {
		marshaled, err := json.Marshal(float64(e))
		if err != nil {
			return nil, err
		}
		return Encrypt(EncryptionKey, marshaled)
	} else {
		return float64(e), nil
	}

}

func (e *EncryptedFloat) SetBSON(raw bson.Raw) error {
	if EnableEncryption {
		var str string
		raw.Unmarshal(&str)
		decrypted, err := Decrypt(EncryptionKey, str)
		if err != nil {
			return err
		}

		var f float64
		err = json.Unmarshal(decrypted, &f)
		if err != nil {
			return err
		}
		*e = EncryptedFloat(f)
	} else {
		var f float64
		raw.Unmarshal(&f)
		*e = EncryptedFloat(f)
	}

	return nil
}

type EncryptedBool bool

func (e EncryptedBool) GetBSON() (interface{}, error) {
	if EnableEncryption {
		var toEncrypt []byte
		if e == true {
			toEncrypt = []byte{0x01}
		} else {
			toEncrypt = []byte{0x00}
		}
		return Encrypt(EncryptionKey, toEncrypt)
	} else {
		return bool(e), nil
	}

}
func (e *EncryptedBool) SetBSON(raw bson.Raw) error {
	if EnableEncryption {
		var str string
		raw.Unmarshal(&str)
		decrypted, err := Decrypt(EncryptionKey, str)
		if err != nil {
			return err
		}

		if decrypted[0] == 0x01 {
			*e = true
		} else {
			*e = false
		}
	} else {
		var b bool
		raw.Unmarshal(&b)
		*e = EncryptedBool(b)
	}

	return nil
}

// Making this an extension of time.Time causes errors with marshaling, so we'll make it a string and use time.Time internally
type EncryptedDate string

var iso8601Format = "2006-01-02T15:04:05-0700"

func (e EncryptedDate) GetBSON() (interface{}, error) {
	if len(e) > 0 && e != "null" {
		d, err := e.GetTime()

		if err != nil {
			return nil, err
		}
		if EnableEncryption {

			return Encrypt(EncryptionKey, []byte(d.Format(iso8601Format)))
		} else {
			return d, nil
		}
	} else {
		return Encrypt(EncryptionKey, []byte{})
	}
}
func (e *EncryptedDate) SetBSON(raw bson.Raw) error {
	if EnableEncryption {
		var str string
		raw.Unmarshal(&str)
		// log.Println("Unmarshaled into", str)
		decrypted, err := Decrypt(EncryptionKey, str)
		if err != nil {
			return err
		}

		if len(decrypted) > 0 {
			t, err := time.Parse(iso8601Format, string(decrypted))

			if err != nil {
				return err
			}

			*e = EncryptedDate(t.Format(iso8601Format))
		}
	} else {
		var t time.Time
		raw.Unmarshal(&t)
		*e = EncryptedDate(t.Format(iso8601Format))
	}
	return nil

}

func (e EncryptedDate) GetTime() (time.Time, error) {
	return time.Parse(iso8601Format, string(e))
}

type EncryptedMap map[string]interface{}

func (e EncryptedMap) GetBSON() (interface{}, error) {

	if EnableEncryption {
		marshaled, err := json.Marshal(map[string]interface{}(e))
		if err != nil {
			return nil, err
		}
		return Encrypt(EncryptionKey, marshaled)
	} else {
		return map[string]interface{}(e), nil
	}

}

func (e *EncryptedMap) SetBSON(raw bson.Raw) error {
	if EnableEncryption {
		var str string
		raw.Unmarshal(&str)
		decrypted, err := Decrypt(EncryptionKey, str)
		if err != nil {
			return err
		}

		m := make(map[string]interface{})
		err = json.Unmarshal(decrypted, &m)
		if err != nil {
			return err
		}
		*e = EncryptedMap(m)
	} else {
		m := make(map[string]interface{})
		raw.Unmarshal(m)
		*e = EncryptedMap(m)
	}

	return nil
}
