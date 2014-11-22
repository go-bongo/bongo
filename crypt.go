// Two-way encryption
// @author Jason Raede <jason@maxwellhealth.com>

package bongo

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"

	"errors"
	// "fmt"

	// "github.com/oleiade/reflections"
	"io"
)

//** BYTE-LEVEL PRIMITIVE METHODS

// Encrypt an array of bytes for storage in the database as a base64 encoded string
func Encrypt(key, val []byte) (string, error) {

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	b := base64.StdEncoding.EncodeToString(val)
	ciphertext := make([]byte, aes.BlockSize+len(b))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}
	cfb := cipher.NewCFBEncrypter(block, iv)
	cfb.XORKeyStream(ciphertext[aes.BlockSize:], []byte(b))
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt a base64-encoded string retrieved from the database and return an array of bytes
func Decrypt(key []byte, encrypted string) ([]byte, error) {

	val, err := base64.StdEncoding.DecodeString(encrypted)
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
	data, err := base64.StdEncoding.DecodeString(string(val))
	if err != nil {
		return nil, err
	}
	return data, nil
}

//** STRUCT-LEVEL ENCRYPTION/DECRYPTION METHODS

// Encrypt a struct. Use tag `encrypted="true"` to designate fields as needing to be encrypted. Fields are encrypted by converting the properties to lowercase (assuming this is going to go into MongoDB), but you can override that using the traditional MGO tag notation (bson="otherField")
