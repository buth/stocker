package main

import (
	"encoding/base64"
	"errors"
	"flag"
	"fmt"
	"github.com/buth/stocker/stocker/backend"
	"github.com/buth/stocker/stocker/backend/redisc"
	"github.com/buth/stocker/stocker/crypto"
	"github.com/buth/stocker/stocker/crypto/chain"
	"io/ioutil"
	"log"
	"os"
)

var config struct {
	GenerateKey, RunDaemon                                                           bool
	SecretFilepath, Crypter, Backend, BackendConnectionType, BackendConnectionString string
}

// genereateKey creates and returns a new key for the chosen crypter as a
// slice of bytes.
func generateKey() []byte {
	switch config.Crypter {
	case "chain":
		return chain.GenerateKey()
	}
	return []byte{}
}

// crypter instantiates a new crypter of the chosen type using a new key or
// the base 64 encoded key present in the file at config.Secret.
func newCrypter() (crypto.Crypter, error) {

	var key []byte

	if config.SecretFilepath != "" {

		// Check the status of the secret file.
		stat, err := os.Stat(config.SecretFilepath)
		if err != nil {
			return nil, err
		}

		// Only proceed if the running user is the only user that can read the
		// secret.
		if stat.Mode() != 0600 && stat.Mode() != 0400 {
			return nil, errors.New("incorrect secret file permissions!")
		}

		// Attempt to read the entire content of the secret file.
		encodedKey, err := ioutil.ReadFile(config.SecretFilepath)
		if err != nil {
			return nil, err
		}

		// Attempt to decode the encoded content into a new slice of bytes.
		key = make([]byte, len(encodedKey)*4)
		base64.StdEncoding.Decode(key, encodedKey)
	} else {
		key = generateKey()
	}

	switch config.Crypter {
	case "chain":
		newChain, err := chain.New(key)
		if err != nil {
			return nil, err
		}
		return newChain, nil
	}

	return nil, errors.New("no crypter selected")
}

func newBackend() (backend.Backend, error) {

	switch config.Backend {
	case "redis":
		newRedisc := redisc.New(config.BackendConnectionType, config.BackendConnectionString)
		return newRedisc, nil
	}

	return nil, errors.New("no backend selected")
}

func init() {
	flag.BoolVar(&config.GenerateKey, "k", false, "generate key")
	flag.BoolVar(&config.RunDaemon, "d", false, "run daemon")
	flag.StringVar(&config.Crypter, "crypter", "chain", "crypter to use")
	flag.StringVar(&config.SecretFilepath, "secret", "", "path to encryption secret")
	flag.StringVar(&config.Backend, "backend", "redis", "backend to use")
	flag.StringVar(&config.BackendConnectionType, "t", "tcp", "backend connection type")
	flag.StringVar(&config.BackendConnectionString, "s", "127.0.0.1:6379", "backend connection string")
	flag.Parse()
}

func main() {

	// Check if we are just supposed to generate a key.
	if config.GenerateKey {
		fmt.Println(base64.StdEncoding.EncodeToString(generateKey()))
		return
	}

	// Attempt to load a crypter.
	crypter, err := newCrypter()
	if err != nil {
		log.Fatalln(err)
	}

	// Attempt to load a backend.
	backend, err := newBackend()
	if err != nil {
		log.Fatalln(err)
	}

	log.Println(crypter, backend)
}
