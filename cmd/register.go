package cmd

import (
	"flag"
	"fmt"
	"github.com/buth/stocker/auth"
	"io/ioutil"
	"log"
)

type RegisterCommand struct {
	Address, Group, PrivateFilepath string
	AllEnvVars                      bool
}

func (cmd *RegisterCommand) Flags(fs *flag.FlagSet) *flag.FlagSet {
	fs.StringVar(&cmd.Address, "a", ":2022", "address of the stocker server")
	fs.StringVar(&cmd.PrivateFilepath, "i", "", "path to an SSH private key")
	return fs
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

	// Get a new client object. If the private key is nil, the method will
	// attempt to use ssh-agent.
	client, err := auth.NewClient(auth.RegisterUser, cmd.Address, privateKey)
	if err != nil {

		log.Fatal(err.Error())
	}

	r, err := client.Run("register", nil)
	if err != nil {

		log.Fatal(err.Error())
	}

	fmt.Print(r)
	// for variable, value := range env {

	// 	// Create an environment specific to this variable.
	// 	runEnv := map[string]string{
	// 		"GROUP":  cmd.Group,
	// 		variable: value,
	// 	}

	// 	client.Run(fmt.Sprintf("export %s", variable), runEnv)
	// }
}
