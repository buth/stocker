package cmd

import (
	"code.google.com/p/go.crypto/ssh"
	"encoding/json"
	"flag"
	"github.com/buth/stocker/auth"
	"github.com/buth/stocker/backend"
	"github.com/buth/stocker/crypto"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

var serverClient *http.Client

type ServerCommand struct {
	SecretFilepath, PrivateFilepath, Backend, BackendNamespace, BackendProtocol, BackendAddress, Group, Address, ReadersURL, WritersURL string
}

func (cmd *ServerCommand) Flags(fs *flag.FlagSet) *flag.FlagSet {
	fs.StringVar(&cmd.Address, "p", ":2022", "address to listen on")
	fs.StringVar(&cmd.Backend, "-backend", "etcd", "backend to use")
	fs.StringVar(&cmd.BackendAddress, "-backend-address", ":4001", "backend address")
	fs.StringVar(&cmd.BackendNamespace, "-backend-namespace", "stocker", "backend namespace")
	fs.StringVar(&cmd.BackendProtocol, "-backend-protocol", "tcp", "backend connection protocol")
	fs.StringVar(&cmd.PrivateFilepath, "i", "/etc/stocker/id_rsa", "path to an ssh private (host) key")
	fs.StringVar(&cmd.SecretFilepath, "k", "/etc/stocker/key", "path to encryption key")
	fs.StringVar(&cmd.ReadersURL, "r", "", "reader public keys URL")
	fs.StringVar(&cmd.WritersURL, "w", "", "writer public keys URL")
	return fs
}

func init() {
	serverClient = &http.Client{
		Transport: &http.Transport{
			MaxIdleConnsPerHost:   1,
			ResponseHeaderTimeout: time.Minute,
		},
	}
}

func serverFetchPublicKeys(url string) ([]ssh.PublicKey, error) {

	// Fetch the public keys.
	response, err := serverClient.Get(url)
	if err != nil {
		return nil, err
	}

	// Read out the entire body.
	jsonResponse, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	// Build a raw keys object that reflects the expected structure of the JSON.
	var rawKeys []struct {
		Key string
	}

	// Try to parse the body of the response as JSON.
	if err := json.Unmarshal(jsonResponse, &rawKeys); err != nil {
		return nil, err
	}

	// Build a new authorizer and iterate through the raw keys, parsing them
	// and then adding them.
	publicKeys := make([]ssh.PublicKey, 0, len(rawKeys))
	for _, rawKey := range rawKeys {

		// We're only interested in the key itself and whether or not there was an error.
		publicKey, _, _, _, err := ssh.ParseAuthorizedKey([]byte(rawKey.Key))
		if err != nil {
			return publicKeys, err
		}

		// Add the key to the list.
		publicKeys = publicKeys[:len(publicKeys)+1]
		publicKeys[len(publicKeys)-1] = publicKey
	}

	return publicKeys, nil
}

func (cmd *ServerCommand) Run(args []string) {

	// Pull a new crypter from the path to the key.
	c, err := crypto.NewCrypterFromFile(cmd.SecretFilepath)
	if err != nil {
		log.Fatal(err)
	}

	// Create a new backend of the specified type.
	b, err := backend.NewBackend(cmd.Backend, cmd.BackendNamespace, cmd.BackendProtocol, cmd.BackendAddress)
	if err != nil {
		log.Fatal(err)
	}

	// Read the private host key from the given path.
	privateBytes, err := ioutil.ReadFile(cmd.PrivateFilepath)
	if err != nil {
		log.Fatal("failed to load private key")
	}

	// Attempt to parse the key.
	private, err := ssh.ParsePrivateKey(privateBytes)
	if err != nil {
		log.Fatal("failed to parse private key")
	}

	// Create a new server using the specified Backend and Crypter.
	server, err := auth.NewServer(b, c, private)
	if err != nil {
		log.Fatal(err)
	}

	// Check if a URL was provided to pull reader keys from.
	if cmd.ReadersURL != "" {

		// Fetch the reader keys.
		readers, err := serverFetchPublicKeys(cmd.ReadersURL)
		if err != nil {
			log.Fatal(err)
		}

		// Add the reader keys to the server.
		for _, reader := range readers {
			server.AddRegisterKey(reader)
		}
	}

	// Check if a URL was provided to pull writer keys from.
	if cmd.WritersURL != "" {

		// Fetch the writer keys.
		writers, err := serverFetchPublicKeys(cmd.WritersURL)
		if err != nil {
			log.Fatal(err)
		}

		// Add the writer keys to the server.
		for _, writer := range writers {
			server.AddWriteKey(writer)
		}
	}

	// Start the server.
	log.Fatal(server.ListenAndServe(cmd.Address))
}
