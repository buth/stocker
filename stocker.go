package main

import (
	"encoding/base64"
	"errors"
	"flag"
	"fmt"
	"github.com/buth/stocker/stocker/crypto"
	"github.com/buth/stocker/stocker/crypto/chain"
	"github.com/coreos/go-etcd/etcd"
	"io/ioutil"
	"log"
	"os"
)

var config struct {
	GenerateKey, RunDaemon                       bool
	SecretFilepath, Crypter, EtcdURL, Key, Value string
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

// crypterFromFile instantiates a new crypter of the chosen type using the
// base 64 encoded key present in the file at filepath.
func crypterFromFile(filepath string) (crypto.Crypter, error) {

	// Attempt to read the entire content of the file at filepath.
	encodedContent, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	// Attempt to decode the encoded content into a new slice of bytes.
	content := make([]byte, len(encodedContent)*4)
	base64.StdEncoding.Decode(content, encodedContent)

	switch config.Crypter {
	case "chain":
		newChain, err := chain.New(content)
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
	flag.StringVar(&config.SecretFilepath, "s", "key.txt", "path to encryption secret")
	flag.StringVar(&config.Crypter, "m", "chain", "crypter to use")
	flag.StringVar(&config.Key, "key", "", "key")
	flag.StringVar(&config.Value, "value", "", "value")
	flag.Parse()
}

func main() {

	// Check if we are just supposed to generate a key.
	if config.GenerateKey {
		fmt.Println(base64.StdEncoding.EncodeToString(generateKey()))
		return
	}

	// Check the status of the secret file.
	stat, err := os.Stat(config.SecretFilepath)
	if err != nil {
		log.Fatalln(err)
	}

	// Only proceed if the running user is the only user that can read the
	// secret.
	if stat.Mode() != 0600 && stat.Mode() != 0400 {
		log.Fatalln("incorrect secret file permissions!")
	}

	// Attempt to load a crypter from the key provided in the secret filepath
	crypter, err := crypterFromFile(config.SecretFilepath)
	if err != nil {
		log.Fatalln(err)
	}

	// Right now, we're limited to expecting etcd to be running on localhost, so
	// we'll just use the NewClient method provided by the etcd client.
	client := etcd.NewClient()

	// Check a key was provided. Arguably, providing a value in the absense of a
	// key could be an error, but for the moment that's not implemented.
	if config.Key != "" {

		// Check if a value was provided. In that case, this is a set operation.
		if config.Value != "" {

			// Encrypt the provided value before sending it.
			encryptedValue, err := crypter.EncryptString(config.Value)
			if err != nil {
				log.Fatalln(err)
			}

			// Run the set command with the encrypted value.
			response, err := client.Set(config.Key, encryptedValue, 0)
			if err != nil {
				log.Fatalln(err)
			}
			log.Println(response)
		} else {

			// Run a get command to retrieve the encrypted value.
			responses, err := client.Get(config.Key)
			if err != nil {
				log.Fatalln(err)
			}

			// Range over the responses, throwing away the index.
			for _, response := range responses {

				// Decrypt the individual response.
				decryptedValue, err := crypter.DecryptString(response.Value)
				if err != nil {
					log.Fatalln(err)
				}
				fmt.Println(decryptedValue)
			}
		}
	}
}
