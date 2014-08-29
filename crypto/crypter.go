package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"io"
	"os"
)

const (

	// SymetricKeyLength is the length in bytes of the key used with the
	// AES-256 algorithm
	SymetricKeyLength = 32

	// HmacKeyLength is the length in bytes of the key used in the HMAC
	// SHA-512 algorithm
	HmacKeyLength = 128

	// HmacOutputLength is the length in bytes of the sum produced by the HMAC
	// SHA-512 algorithm.
	HmacOutputLength = 64
)

// A Crypter is an enCrypter/deCrypter set to use a specific encryption key (for
// AES-256 in CBC mode) and signing key (for HMAC SHA-512) combination.
type Crypter struct {
	hmacKey, symetricKey []byte
	block                cipher.Block
}

// New creates and returns a new Crypter. Keys are obtained by reading from
// the provided reader.
func NewCrypter(key io.Reader) (*Crypter, error) {

	// Create a new Crypter object.
	Crypter := &Crypter{}

	// Set the hmac key.
	Crypter.hmacKey = make([]byte, HmacKeyLength)
	if _, err := io.ReadFull(key, Crypter.hmacKey); err != nil {
		return Crypter, err
	}

	// Set the symetric key.
	Crypter.symetricKey = make([]byte, SymetricKeyLength)
	if _, err := io.ReadFull(key, Crypter.symetricKey); err != nil {
		return Crypter, err
	}

	// Try to create a cipher from the symetric key.
	block, err := aes.NewCipher(Crypter.symetricKey)
	if err != nil {
		return Crypter, err
	}

	// Set the block.
	Crypter.block = block

	return Crypter, nil
}

func NewRandomCrypter() (*Crypter, error) {
	return NewCrypter(rand.Reader)
}

func NewCrypterFromFile(filepath string) (*Crypter, error) {

	// Check the status of the secret file.
	stat, err := os.Stat(filepath)
	if err != nil {
		return nil, err
	}

	// Only proceed if the running user is the only user that can read the
	// secret.
	if mode := stat.Mode(); mode != 0600 && mode != 0400 {
		return nil, CrypterError{"incorrect file mode for key"}
	}

	// Attempt to read the entire content of the secret file.
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}

	// Defer the closing of the file, ignoring any error.
	defer file.Close()

	// Create a new decoder.
	decoder := base64.NewDecoder(base64.StdEncoding, file)

	return NewCrypter(decoder)
}

// hmac computes and returns SHA-512 Hmac sum using the signing key.
func (c *Crypter) hmac(message []byte) []byte {
	signer := hmac.New(sha512.New, c.hmacKey)
	signer.Write(message)
	return signer.Sum(nil)
}

// encrypt encrypts a slice of bytes using the AES-256 cipher in CBC mode and
// returns an usigned sice of cipher bytes that begins with the IV.
func (c *Crypter) encrypt(plainbytes []byte) ([]byte, error) {

	// Initialize size with room for the IV.
	size := aes.BlockSize + len(plainbytes)

	// Add extra padding if the size is not a multiple of the Block size.
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

	// Create the enCrypter and crypt the plainbytes in place.
	mode := cipher.NewCBCEncrypter(c.block, iv)
	mode.CryptBlocks(cipherbytes[aes.BlockSize:], cipherbytes[aes.BlockSize:])

	return cipherbytes, nil
}

// decrypt decrypts a slice of cipherbytes using the AES-256 cipher in CBC
// mode and returns a slice of plain bytes. The first Block of the cipherbytes
// argument is expected to be the IV. It does not verify or expect a signature
// to be present in the cipherbytes argument.
func (c *Crypter) decrypt(cipherbytes []byte) ([]byte, error) {

	// We need an IV and at least one Block of cipherbytes to proceed.
	if len(cipherbytes) < aes.BlockSize*2 {
		return []byte{}, CrypterError{"cipherbytes is too short"}
	}

	// CBC mode always works in whole Blocks.
	if len(cipherbytes)%aes.BlockSize != 0 {
		return []byte{}, CrypterError{"cipherbytes is not a multiple of the block size"}
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

// Encrypt converts plainbytes to signed cipherbytes by encrypting the
// plainbytes using AES-256 and prepending a Hmac SHA-512 signature.
func (c *Crypter) Encrypt(plainbytes []byte) ([]byte, error) {

	// Encrypt the slice of plainbytes, producing cipherbytes.
	cipherbytes, err := c.encrypt(plainbytes)
	if err != nil {
		return nil, err
	}

	// Get the signatrue for the cipherbytes.
	hmacbytes := c.hmac(cipherbytes)

	// Copy all the bytes into a single byte string.
	messagebytes := make([]byte, HmacOutputLength+len(cipherbytes))
	copy(messagebytes[:len(hmacbytes)], hmacbytes)
	copy(messagebytes[len(hmacbytes):], cipherbytes)

	return messagebytes, nil
}

// EncryptString converts plaintext to signed, base 64 encoded ciphertext by
// encrypting the plaintext using AES-256 and prepending a Hmac SHA-512
// signature.
func (c *Crypter) EncryptString(plaintext string) (string, error) {

	// Convert the string to a slice of bytes.
	messagebytes, err := c.Encrypt([]byte(plaintext))
	if err != nil {
		return "", err
	}

	// Convert the result to a base 64 encoded string.
	return base64.StdEncoding.EncodeToString(messagebytes), nil
}

// Decrypt converts signed slice of cipherbytes to plainbytes by first
// validating a prepended Hmac SHA-512 signature and then decrypting the
// remaining message using AES-256.
func (c *Crypter) Decrypt(messagebytes []byte) ([]byte, error) {

	// Check that message bytes is long enough.
	if len(messagebytes) < 64 {
		return nil, CrypterError{"message signature is too short"}
	}

	// Check the signature.
	if hmac.Equal(messagebytes[:64], c.hmac(messagebytes[64:])) != true {
		return nil, CrypterError{"invalid signature"}
	}

	// Decode the encrypted bytes.
	plainbytes, err := c.decrypt(messagebytes[64:])
	if err != nil {
		return nil, err
	}

	return plainbytes, nil
}

// DecryptString converts signed, base 64 encoded ciphertext to plaintext by
// first validating a prepended Hmac SHA-512 signature and then decrypting the
// remaining message using AES-256.
func (c *Crypter) DecryptString(message string) (string, error) {

	// Decode the base 64 string.
	messagebytes, err := base64.StdEncoding.DecodeString(message)
	if err != nil {
		return "", err
	}

	plainbytes, err := c.Decrypt(messagebytes)
	if err != nil {
		return "", err
	}

	// Convert the result to a string.
	return string(plainbytes), nil
}

// writeBase64 encodes the raw key as a base64-encoded string.
func (c *Crypter) writeBase64(w io.Writer) error {

	// Create a new encoder.
	encoder := base64.NewEncoder(base64.StdEncoding, w)

	if _, err := encoder.Write(c.hmacKey); err != nil {
		return err
	}

	if _, err := encoder.Write(c.symetricKey); err != nil {
		return err
	}

	return encoder.Close()
}

// ToFile saves the Crypter's keys to disk, encoded as a base 64 string.
func (c *Crypter) ToFile(filename string) error {

	// Create a new file. This will wipe out any existing file (if we can
	// write to it) and set permissions to 666.
	out, err := os.Create(filename)
	if err != nil {
		return err
	}

	// Set more restrictive permissions on the file *before* we write to it.
	if err := out.Chmod(0600); err != nil {
		return err
	}

	// Defer the closing of the file, ignoring any error.
	defer out.Close()

	// Return the encoded string.
	return c.writeBase64(out)
}

// ToFile saves the Crypter's keys to disk, encoded as a base 64 string.
func (c *Crypter) ToString() (string, error) {
	out := bytes.NewBuffer([]byte{})
	if err := c.writeBase64(out); err != nil {
		return "", err
	}
	return out.String(), nil
}

// CrypterError represents a run-time error in a Crypter method.
type CrypterError struct {
	Err string
}

func (e CrypterError) Error() string {
	return fmt.Sprintf("crypter: %s", e.Err)
}
