package crypter

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"
)

type AES struct {
	key1, key2 []byte
	block      cipher.Block
}

func (c *AES) Load(keytext string) error {

	// Decode the keytext from base 64.
	keybytes, err := base64.StdEncoding.DecodeString(keytext)
	if err != nil {
		return err
	}

	// Check that it has the right number of bytes.
	if len(keybytes) != 288 {
		return errors.New("wrong number of bytes in key")
	}

	// Create the cipher.
	block, err := aes.NewCipher(c.key1)
	if err != nil {
		return err
	}

	c.key1 = keybytes[:32]
	c.key2 = keybytes[32:]
	c.block = block

	return nil
}

func (c *AES) hmac(message []byte) []byte {
	mac := hmac.New(sha256.New, c.key2)
	mac.Write(message)
	return mac.Sum(nil)
}

func (c *AES) encode(plainbytes []byte) ([]byte, error) {

	// Initialize size with room for the IV.
	size := aes.BlockSize + len(plainbytes)

	// Add padding if necessary.
	if extra := len(plainbytes) % aes.BlockSize; extra != 0 {
		size += aes.BlockSize - extra
	}

	// Create the cipher text slice and copy in the plainbytes.
	cipherbytes := make([]byte, size)
	copy(cipherbytes[aes.BlockSize:], plainbytes)

	// Use an IV at the front of the cipherbytes, and attempt to read in random bits.
	iv := cipherbytes[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return []byte{}, err
	}

	// Create the encrypter and crypt the plainbytes in place.
	mode := cipher.NewCBCEncrypter(c.block, iv)
	mode.CryptBlocks(cipherbytes[aes.BlockSize:], cipherbytes[aes.BlockSize:])

	return cipherbytes, nil
}

func (c *AES) decode(cipherbytes []byte) ([]byte, error) {

	// We need an IV and at least one block of cipherbytes to proceed.
	if len(cipherbytes) < aes.BlockSize*2 {
		return []byte{}, errors.New("cipherbytes too short")
	}

	// CBC mode always works in whole blocks.
	if len(cipherbytes)%aes.BlockSize != 0 {
		return []byte{}, errors.New("ciphertext is not a multiple of the block size")
	}

	// IV is the first BlockSize bytes of the message.
	iv := cipherbytes[:aes.BlockSize]

	// Allocate a new byte array to hold the plainbytes
	plainbytes := make([]byte, len(cipherbytes)-aes.BlockSize)

	// Decrypt the cipherbytes and trim the result.
	mode := cipher.NewCBCDecrypter(c.block, iv)
	mode.CryptBlocks(plainbytes, cipherbytes[aes.BlockSize:])
	plainbytes = bytes.TrimRight(plainbytes, "\x00")

	return plainbytes, nil
}

func (c *AES) EncodeString(plaintext string) (string, error) {

	// Convert the string to bytes
	plainbytes := []byte(plaintext)

	// Encoded the unencrypted bytes.
	cipherbytes, err := c.encode(plainbytes)
	if err != nil {
		return "", err
	}

	// Sign the cipherbytes.
	hmacbytes := c.hmac(cipherbytes)

	// Copy all the bytes into a single byte string.
	messagebytes := make([]byte, len(hmacbytes)+len(cipherbytes))
	copy(messagebytes[:len(hmacbytes)], hmacbytes)
	copy(messagebytes[len(hmacbytes):], cipherbytes)

	// Convert the result to a base 64 encoded string.
	return base64.StdEncoding.EncodeToString(messagebytes), nil
}

func (c *AES) DecodeString(message string) (string, error) {

	// Decode the base 64 string.
	messagebytes, err := base64.StdEncoding.DecodeString(message)
	if err != nil {
		return "", err
	}

	// Check the signature.
	if hmac.Equal(messagebytes[:32], c.hmac(messagebytes[32:])) != true {
		return "", errors.New("invalid signature")
	}

	// Decode the encrypted bytes.
	plainbytes, err := c.decode(messagebytes[32:])
	if err != nil {
		return "", err
	}

	// Convert the result to a string.
	plaintext := string(plainbytes)
	return plaintext, nil
}
