// Two-way encryption
// @author Jason Raede <jason@maxwellhealth.com>

// EncryptAndSave - 123850
// EncryptInitializeDocumentFromDB - 102750
package bongo

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	// "encoding/base64"
	"encoding/hex"
	"encoding/json"

	"errors"
	"github.com/maxwellhealth/mgo/bson"
	// "log"
	"strconv"
	"time"
	// "strings"
	// "fmt"

	// "github.com/oleiade/reflections"
	"io"
)

// var encryptionKey = []byte("asdf1234asdf1234")

//** BYTE-LEVEL PRIMITIVE METHODS

// For some reason when we get the raw data back it has a byte value in front
func parseEncryptedString(str string) string {
	// log.Println("Parsing encrypted string")
	return str[1:]
}

// Encrypt an array of bytes for storage in the database as a base64 encoded string
func Encrypt(key, val []byte) (string, error) {

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	ciphertext := make([]byte, aes.BlockSize+len(val))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}
	cfb := cipher.NewCFBEncrypter(block, iv)
	cfb.XORKeyStream(ciphertext[aes.BlockSize:], val)
	result := hex.EncodeToString(ciphertext)
	return result, nil
}

// Decrypt a base64-encoded string retrieved from the database and return an array of bytes
func Decrypt(key []byte, encrypted string) ([]byte, error) {
	val, err := hex.DecodeString(encrypted)

	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	if len(val) < aes.BlockSize {
		return nil, errors.New("ciphertext too short")
	}
	iv := val[:aes.BlockSize]
	val = val[aes.BlockSize:]
	cfb := cipher.NewCFBDecrypter(block, iv)
	cfb.XORKeyStream(val, val)
	// data, err := base64.StdEncoding.DecodeString(string(val))
	if err != nil {
		return nil, err
	}
	return val, nil
}

var encryptionKey = []byte("asdf1234asdf1234")

//* ENCRYPTED TYPES */
type EncryptedString string

func (e EncryptedString) GetBSON() (interface{}, error) {

	return Encrypt(encryptionKey, []byte(string(e)))
}
func (e *EncryptedString) SetBSON(raw bson.Raw) error {

	var str string
	raw.Unmarshal(&str)
	// log.Println("Unmarshaled into", str)
	decrypted, err := Decrypt(encryptionKey, str)
	if err != nil {
		return err
	}
	*e = EncryptedString(string(decrypted))
	return nil
}

type EncryptedInt int

func (e EncryptedInt) GetBSON() (interface{}, error) {
	return Encrypt(encryptionKey, []byte(strconv.Itoa(int(e))))
}
func (e *EncryptedInt) SetBSON(raw bson.Raw) error {
	var str string
	raw.Unmarshal(&str)
	decrypted, err := Decrypt(encryptionKey, str)
	if err != nil {
		return err
	}

	intVal, err := strconv.Atoi(string(decrypted))
	if err != nil {
		return err
	}
	*e = EncryptedInt(intVal)
	return nil
}

type EncryptedFloat float64

func (e EncryptedFloat) GetBSON() (interface{}, error) {
	// return float64(e)
	marshaled, err := json.Marshal(float64(e))
	if err != nil {
		return nil, err
	}
	return Encrypt(encryptionKey, marshaled)
}

func (e *EncryptedFloat) SetBSON(raw bson.Raw) error {
	var str string
	raw.Unmarshal(&str)
	decrypted, err := Decrypt(encryptionKey, str)
	if err != nil {
		return err
	}

	var f float64
	err = json.Unmarshal(decrypted, &f)
	if err != nil {
		return err
	}
	*e = EncryptedFloat(f)
	return nil
}

type EncryptedBool bool

func (e EncryptedBool) GetBSON() (interface{}, error) {
	var toEncrypt []byte
	if e == true {
		toEncrypt = []byte{0x01}
	} else {
		toEncrypt = []byte{0x00}
	}
	return Encrypt(encryptionKey, toEncrypt)
}
func (e *EncryptedBool) SetBSON(raw bson.Raw) error {
	var str string
	raw.Unmarshal(&str)
	decrypted, err := Decrypt(encryptionKey, str)
	if err != nil {
		return err
	}

	if decrypted[0] == 0x01 {
		*e = true
	} else {
		*e = false
	}
	return nil
}

type EncryptedDate time.Time

func (e EncryptedDate) GetBSON() (interface{}, error) {
	d := time.Time(e)
	return Encrypt(encryptionKey, []byte(d.String()))
}
func (e *EncryptedDate) SetBSON(raw bson.Raw) error {

	var str string
	raw.Unmarshal(&str)
	// log.Println("Unmarshaled into", str)
	decrypted, err := Decrypt(encryptionKey, str)

	if err != nil {
		return err
	}

	t, err := time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", string(decrypted))

	if err != nil {
		return err
	}

	*e = EncryptedDate(t)
	return nil
}
