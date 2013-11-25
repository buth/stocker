package chain

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

// A chain is an encrypter/decrypter set to use a specific encryption key (for
// AES-256 in CBC mode) and signing key (for HMAC SHA-256) combination.
type chain struct {
	signer []byte
	block  cipher.Block
}

// GenerateKey creates and returns a new random key that can be used to create
// a new chain.
func GenerateKey() []byte {
	key := make([]byte, 288)
	io.ReadFull(rand.Reader, key)
	return key
}

// New creates and returns a new chain.chain. The key argument should be the
// combined AES-256 and SHA-256 keys (in that order) for a total length of 288
// bytes.
func New(key []byte) (*chain, error) {

	// If no key is provided, generate one.
	if len(key) == 0 {
		key = GenerateKey()
	}

	// Create the cipher. The cipher itself only stores an expanded version of
	// the key, so there is no need to copy it.
	block, err := aes.NewCipher(key[:32])
	if err != nil {
		return &chain{}, err
	}

	// The new chain object needs its own copy of the signing key.
	signer := make([]byte, 256)
	copy(signer, key[32:])

	return &chain{signer: signer, block: block}, nil
}

// hmac computes and returns SHA-256 HMAC sum using the signing key.
func (c *chain) hmac(message []byte) []byte {
	mac := hmac.New(sha256.New, c.signer)
	mac.Write(message)
	return mac.Sum(nil)
}

// encrypt encrypts a slice of bytes using the AES-256 cipher in CBC mode and
// returns an usigned sice of cipher bytes that begins with the IV.
func (c *chain) encrypt(plainbytes []byte) ([]byte, error) {

	// Initialize size with room for the IV.
	size := aes.BlockSize + len(plainbytes)

	// Add extra padding if the size is not a multiple of the block size.
	if extra := len(plainbytes) % aes.BlockSize; extra != 0 {
		size += aes.BlockSize - extra
	}

	// Create the cipherbytes slice and copy in the plainbytes.
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

// decrypt decrypts a slice of cipherbytes using the AES-256 cipher in CBC
// mode and returns a slice of plain bytes. The first block of the cipherbytes
// argument is expected to be the IV. It does not verify or expect a signature
// to be present in the cipherbytes argument.
func (c *chain) decrypt(cipherbytes []byte) ([]byte, error) {

	// We need an IV and at least one block of cipherbytes to proceed.
	if len(cipherbytes) < aes.BlockSize*2 {
		return []byte{}, errors.New("cipherbytes is too short")
	}

	// CBC mode always works in whole blocks.
	if len(cipherbytes)%aes.BlockSize != 0 {
		return []byte{}, errors.New("cipherbytes is not a multiple of the block size")
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

// EncryptString converts plaintext to signed, base 64 encoded ciphertext by
// encrypting the plaintext using AES-256 and prepending a HMAC SHA-256
// signature.
func (c *chain) EncryptString(plaintext string) (string, error) {

	// Convert the string to a slice of bytes.
	plainbytes := []byte(plaintext)

	// Encrypt the slice of plainbytes, producing cipherbytes.
	cipherbytes, err := c.encrypt(plainbytes)
	if err != nil {
		return "", err
	}

	// Get the signatrue for the cipherbytes.
	hmacbytes := c.hmac(cipherbytes)

	// Copy all the bytes into a single byte string.
	messagebytes := make([]byte, len(hmacbytes)+len(cipherbytes))
	copy(messagebytes[:len(hmacbytes)], hmacbytes)
	copy(messagebytes[len(hmacbytes):], cipherbytes)

	// Convert the result to a base 64 encoded string.
	return base64.StdEncoding.EncodeToString(messagebytes), nil
}

// DecryptString converts signed, base 64 encoded ciphertext to plaintext by
// first validating a prepended HMAC SHA-256 signature and then decrypting the
// remaining message using AES-256.
func (c *chain) DecryptString(message string) (string, error) {

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
	plainbytes, err := c.decrypt(messagebytes[32:])
	if err != nil {
		return "", err
	}

	// Convert the result to a string.
	plaintext := string(plainbytes)
	return plaintext, nil
}
