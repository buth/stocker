package crypto

import (
	"encoding/base64"
	"errors"
	"io/ioutil"
	"os"
)

type Crypter interface {
	EncryptString(plaintext string) (message string, err error)
	DecryptString(message string) (plaintext string, err error)
}

type Key []byte

func KeyFromFile(filepath string) (Key, error) {

	// Check the status of the secret file.
	stat, err := os.Stat(filepath)
	if err != nil {
		return nil, err
	}

	// Only proceed if the running user is the only user that can read the
	// secret.
	if stat.Mode() != 0600 && stat.Mode() != 0400 {
		return nil, errors.New("incorrect secret file permissions!")
	}

	// Attempt to read the entire content of the secret file.
	encodedKey, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	// Attempt to decode the encoded content into a new slice of bytes.
	key := make([]byte, len(encodedKey)*4)
	base64.StdEncoding.Decode(key, encodedKey)

	// Return the key
	return key, nil
}
