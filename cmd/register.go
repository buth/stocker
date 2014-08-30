package cmd

import (
	"flag"
	"fmt"
	"github.com/buth/stocker/auth"
	"io/ioutil"
	"log"
	"math/rand"
	"time"
)

type RegisterCommand struct {
	Address, Group, PrivateFilepath, Filename string
	AllEnvVars                                bool
	Splay                                     int
}

func (cmd *RegisterCommand) Flags(fs *flag.FlagSet) *flag.FlagSet {
	fs.StringVar(&cmd.Address, "a", ":2022", "address of the stocker server")
	fs.StringVar(&cmd.PrivateFilepath, "i", "", "path to an SSH private key")
	fs.StringVar(&cmd.Filename, "f", "", "filename to save SSH private key to")
	fs.IntVar(&cmd.Splay, "s", 0, "splay for continuous updates")
	return fs
}

func register(user, address string, privateKey []byte) ([]byte, error) {

	// Get a new client object. If the private key is nil, the method will
	// attempt to use ssh-agent.
	client, err := auth.NewClient(user, address, privateKey)
	if err != nil {
		return nil, err
	}

	// Defer closing the client.
	defer client.Close()

	// Send the register command. The expected response is an SSH private key.
	key, err := client.Run("register", nil)
	if err != nil {
		return nil, err
	}

	return []byte(key), nil
}

func (cmd *RegisterCommand) Run(args []string) {

	// Read the private key from disk if a filepath has been provided.
	var privateKey []byte
	if cmd.PrivateFilepath != "" {
		privateKeyBytes, err := ioutil.ReadFile(cmd.PrivateFilepath)
		if err != nil {
			log.Fatal(err.Error())
		}
		privateKey = privateKeyBytes
	}

	// Send the register command. The expected response is an SSH private key.
	key, err := register(auth.RegisterUser, cmd.Address, privateKey)
	if err != nil {
		log.Fatal(err)
	}

	if cmd.Filename != "" {

		// Try to write a file.
		if err := ioutil.WriteFile(cmd.Filename, key, 0600); err != nil {
			log.Fatal(err)
		}

		if cmd.Splay > 0 {

			rand.Seed(time.Now().UnixNano())

			for {

				time.Sleep(time.Second * time.Duration(rand.Intn(cmd.Splay)))

				newKey, err := register(auth.ReaderUser, cmd.Address, key)
				if err != nil {
					log.Fatal(err)
				}

				if err := ioutil.WriteFile(cmd.Filename, newKey, 0600); err != nil {
					log.Fatal(err)
				}

				key = newKey
			}
		}
	} else {

		// Print the string to STDOUT.
		fmt.Printf("%s", key)
	}
}
