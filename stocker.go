package main

import (
	"encoding/base64"
	"errors"
	"flag"
	"fmt"
	// "github.com/buth/stocker/stocker/backend"
	"github.com/buth/stocker/stocker/crypto"
	"github.com/buth/stocker/stocker/crypto/chain"
	"log"
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
func newCrypter(key []byte) (crypto.Crypter, error) {
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

	// A blank key should be acceptable.
	key := []byte{}

	// Check if we should try to load a key from a file on disk.
	if config.SecretFilepath != "" {
		if keyFromFile, err := crypto.KeyFromFile(config.SecretFilepath); err != nil {
			log.Fatalln(err)
		} else {
			key = keyFromFile
		}
	}

	// Attempt to load a crypter.
	crypter, err := newCrypter(key)
	if err != nil {
		log.Fatalln(err)
	}

	// Attempt to load a backend.
	log.Println(crypter)
}
