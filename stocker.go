package main

import (
	"encoding/base64"
	"errors"
	"flag"
	"fmt"
	"github.com/buth/stocker/stocker"
	"github.com/buth/stocker/stocker/backend"
	"github.com/buth/stocker/stocker/backend/redis"
	"github.com/buth/stocker/stocker/crypto"
	"github.com/buth/stocker/stocker/crypto/chain"
	"log"
	"strings"
)

var config struct {
	Group, SecretFilepath, Crypter, Backend, BackendConnectionType, BackendConnectionString string
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

// newCrypter instantiates a new crypter of the chosen type using a new key or
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

// newBackend instantiates a new backend of the chosen type using the
// connection information in config.BackendConnectionType and
// config.BackendConnectionString.
func newBackend() (backend.Backend, error) {
	switch config.Backend {
	case "redis":
		newRedisBackend := redis.New(config.BackendConnectionType, config.BackendConnectionString)
		return newRedisBackend, nil
	}
	return nil, errors.New("no backend selected")
}

func init() {

	flag.StringVar(&config.Crypter, "stocker-crypter", "chain", "crypter to use")
	flag.StringVar(&config.SecretFilepath, "stocker-secret", "", "path to encryption secret")
	flag.StringVar(&config.Backend, "stocker-backend", "redis", "backend to use")
	flag.StringVar(&config.BackendConnectionType, "stocker-backend-connection-type", "tcp", "backend connection type")
	flag.StringVar(&config.BackendConnectionString, "stocker-backend-connection-string", "127.0.0.1:6379", "backend connection string")
}

func main() {
	flag.Parse()

	// A blank key should be acceptable.
	key := []byte{}

	// Check if we should try to load a key from a file on disk.
	if config.SecretFilepath != "" {
		if keyFromFile, err := crypto.KeyFromFile(config.SecretFilepath); err != nil {
			log.Fatalln(err)
		} else {
			key = keyFromFile
		}
	} else {
		key = generateKey()
	}

	// Attempt to load a crypter.
	c, err := newCrypter(key)
	if err != nil {
		log.Fatalln(err)
	}

	// Attempt to load a backend.
	b, err := newBackend()
	if err != nil {
		log.Fatalln(err)
	}

	// What are we doing here?
	switch flag.Arg(0) {

	case "daemon":

		if flag.NArg() < 2 {
			log.Fatal("daemon requires a groupname!")
		}

		if s, err := stocker.New(flag.Arg(1), b, c); err != nil {

			log.Fatal(err)
		} else {
			log.Println(s.Run())
		}

	case "run":

		if flag.NArg() < 3 {
			log.Fatal("run requires a group and resource name!")
		}

		group := flag.Arg(1)
		name := flag.Arg(2)
		message := strings.Join(flag.Args()[3:], " ")

		// Save the new configuration.
		if err := b.Set(backend.Key("conf", group, "resource", name), message); err != nil {
			log.Fatal(err)
		}

		// Add the resource to the list for this group.
		if err := b.Add(backend.Key("conf", group, "resources"), name); err != nil {
			log.Fatal(err)
		}

		// Signal listeners for this group to reload the resource.
		if err := b.Publish(backend.Key("cast", group, name), "reload"); err != nil {
			log.Fatal(err)
		}

	case "key":
		fmt.Println(base64.StdEncoding.EncodeToString(key))
	}
}
