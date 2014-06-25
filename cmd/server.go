package cmd

import (
	"code.google.com/p/go.crypto/ssh"
	"encoding/json"
	"github.com/buth/stocker/auth"
	"github.com/buth/stocker/backend"
	"github.com/buth/stocker/crypto"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

var serverConfig struct {
	SecretFilepath, PrivateFilepath, Backend, BackendNamespace, BackendProtocol, BackendAddress, Group, Address, ReadersURL, WritersURL string
}

var serverClient *http.Client

var Server = &Command{
	UsageLine: "server [options]",
	Short:     "start a stocker server",
}

func init() {
	Server.Run = serverRun
	Server.Flag.StringVar(&serverConfig.Address, "a", ":2022", "address to listen on")
	Server.Flag.StringVar(&serverConfig.Backend, "b", "redis", "backend to use")
	Server.Flag.StringVar(&serverConfig.BackendAddress, "h", ":6379", "backend address")
	Server.Flag.StringVar(&serverConfig.BackendNamespace, "n", "stocker", "backend namespace")
	Server.Flag.StringVar(&serverConfig.BackendProtocol, "t", "tcp", "backend connection protocol")
	Server.Flag.StringVar(&serverConfig.PrivateFilepath, "i", "/etc/stocker/id_rsa", "path to an ssh private key")
	Server.Flag.StringVar(&serverConfig.SecretFilepath, "k", "/etc/stocker/key", "path to encryption key")
	Server.Flag.StringVar(&serverConfig.ReadersURL, "r", "", "retrieve reader public keys from this URL")
	Server.Flag.StringVar(&serverConfig.WritersURL, "w", "", "retrieve writer public keys from this URL")

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

func serverRun(cmd *Command, args []string) {

	c, err := crypto.NewCrypterFromFile(serverConfig.SecretFilepath)
	if err != nil {
		log.Fatal(err)
	}

	b, err := backend.NewBackend(serverConfig.Backend, serverConfig.BackendNamespace, serverConfig.BackendProtocol, serverConfig.BackendAddress)
	if err != nil {
		log.Fatal(err)
	}

	privateBytes, err := ioutil.ReadFile(serverConfig.PrivateFilepath)
	if err != nil {
		log.Fatal("failed to load private key")
	}

	private, err := ssh.ParsePrivateKey(privateBytes)
	if err != nil {
		log.Fatal("failed to parse private key")
	}

	// Create a new server using the specified Backend and Crypter.
	server := auth.NewServer(b, c, private)

	// Check if a URL was provided to pull reader keys from.
	if serverConfig.ReadersURL != "" {

		// Fetch the reader keys.
		readers, err := serverFetchPublicKeys(serverConfig.ReadersURL)
		if err != nil {
			log.Fatal(err)
		}

		// Add the reader keys to the server.
		for _, reader := range readers {
			server.AddReadKey(reader)
		}
	}

	// Check if a URL was provided to pull writer keys from.
	if serverConfig.WritersURL != "" {

		// Fetch the writer keys.
		writers, err := serverFetchPublicKeys(serverConfig.WritersURL)
		if err != nil {
			log.Fatal(err)
		}

		// Add the writer keys to the server.
		for _, writer := range writers {
			server.AddWriteKey(writer)
		}
	}

	// Start the server.
	log.Fatal(server.ListenAndServe(serverConfig.Address))
}
