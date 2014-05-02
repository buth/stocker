package crypto

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

const (
	KeyLen = 288
)

// Key is a slice of bytes representing a signing key and the symetric
// encryption key.
type Key []byte

// Valid checks that the correct ammount of bytes is present in the key.
func (k Key) Valid() bool {
	return len(k) == KeyLen
}

// String returns the value of the key encoded as a base 64 string.
func (k Key) String() string {
	return fmt.Sprintf("%s", base64.StdEncoding.EncodeToString(k))
}

// String returns the value of the key encoded as a base 64 string.
func (k Key) ToFile(filename string) error {

	// Create a new file. This will wipe out any existing file (if we can
	// write to it) and set permissions to 666.
	out, err := os.Create(filename)
	if err != nil {
		return err
	}

	if err := out.Chmod(0600); err != nil {
		return err
	}

	if _, err := out.WriteString(k.String()); err != nil {
		return err
	}

	if err := out.Close(); err != nil {
		return err
	}

	return nil
}

// NewKey creates and returns a new random key that can be used to create a
// new crypter.
func NewKey() Key {
	key := make([]byte, 288)
	io.ReadFull(rand.Reader, key)
	return key
}

// NewKeyFromFile creates and returns a new key described by a given filepath.
func NewKeyFromFile(filepath string) (Key, error) {

	// Check the status of the secret file.
	stat, err := os.Stat(filepath)
	if err != nil {
		return Key{}, err
	}

	// Only proceed if the running user is the only user that can read the
	// secret.
	if mode := stat.Mode(); mode != 0600 && mode != 0400 {
		return Key{}, KeyPermissionsError{mode}
	}

	// Attempt to read the entire content of the secret file.
	encodedKey, err := ioutil.ReadFile(filepath)
	if err != nil {
		return Key{}, err
	}

	// Attempt to decode the encoded content into a new slice of bytes.
	key := make([]byte, KeyLen)
	base64.StdEncoding.Decode(key, encodedKey)

	// Return the key
	return key, nil
}

type KeyPermissionsError struct {
	Mode os.FileMode
}

func (e KeyPermissionsError) Error() string {
	return fmt.Sprintf("crypto: incorrect key permissions \"%s\"", e.Mode)
}
