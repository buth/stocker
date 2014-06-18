package cmd

import (
	"code.google.com/p/go.crypto/ssh"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/buth/stocker/auth"
	"github.com/buth/stocker/backend"
	"github.com/buth/stocker/crypto"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

var serverConfig struct {
	SecretFilepath, PrivateFilepath, Backend, BackendNamespace, BackendProtocol, BackendAddress, Group, Address string
}

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
}

func serverRun(cmd *Command, args []string) {

	// Build a new HTTP client and transport. We're only going to use it to
	// make a single request, so we don't need keep-alive, etc.
	client := &http.Client{
		Transport: &http.Transport{
			DisableKeepAlives:     true,
			MaxIdleConnsPerHost:   1,
			ResponseHeaderTimeout: time.Minute,
		},
	}

	// Fetch the public keys.
	response, err := client.Get("https://s3.amazonaws.com/newsdev-ops/keys.json")
	if err != nil {
		log.Fatal(err)
	}

	// Read out the entire body.
	jsonResponse, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}

	// Build a raw keys object that reflects the expected structure of the JSON.
	var rawKeys []struct {
		Key string
	}

	// Try to parse the body of the response as JSON.
	if err := json.Unmarshal(jsonResponse, &rawKeys); err != nil {
		log.Fatal(err)
	}

	// Build a new authorizer and iterate through the raw keys, parsing them
	// and then adding them.
	authorizer := auth.NewAuthorizer()
	for _, rawKey := range rawKeys {

		// We're only interested in the key itself and whether or not there was an error.
		publicKey, _, _, _, err := ssh.ParseAuthorizedKey([]byte(rawKey.Key))
		if err != nil {
			log.Fatal(err)
		}

		// Add the key to the authorizer.
		authorizer.AddKey(publicKey)
	}

	certChecker := &ssh.CertChecker{
		IsAuthority: func(auth ssh.PublicKey) bool { return false },
		UserKeyFallback: func(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {

			if authorizer.Authorize(key) {
				return &ssh.Permissions{}, nil
			}

			return nil, errors.New("key not found")
		},
	}

	// An SSH server is represented by a ServerConfig, which holds
	// certificate details and handles authentication of ServerConns.
	config := &ssh.ServerConfig{
		PublicKeyCallback: certChecker.Authenticate,
	}

	privateBytes, err := ioutil.ReadFile(PrivateFilepath)
	if err != nil {
		log.Fatal("Failed to load private key")
	}

	private, err := ssh.ParsePrivateKey(privateBytes)
	if err != nil {
		log.Fatal("Failed to parse private key")
	}

	config.AddHostKey(private)

	key, err := crypto.NewKeyFromFile(serverConfig.SecretFilepath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	c, err := crypto.NewCrypter(key)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	b, err := backend.NewBackend(serverConfig.Backend, serverConfig.BackendNamespace, serverConfig.BackendProtocol, serverConfig.BackendAddress)
	if err != nil {
		log.Fatal(err)
	}

	server := auth.NewServer(config, b, c)

	log.Fatal(server.ListenAndServe(serverConfig.Address))

}
