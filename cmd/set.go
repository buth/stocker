package cmd

import (
	"code.google.com/p/gopass"
	"flag"
	"fmt"
	"github.com/buth/stocker/auth"
	"io/ioutil"
	"log"
	"os"
)

type SetCommand struct {
	Address, Group, PrivateFilepath string
	AllEnvVars                      bool
}

func (cmd *SetCommand) Flags(fs *flag.FlagSet) *flag.FlagSet {
	fs.StringVar(&cmd.Address, "a", ":2022", "address of the stocker server")
	fs.StringVar(&cmd.Group, "g", "default", "group to use for storing and retrieving data")
	fs.StringVar(&cmd.PrivateFilepath, "i", "", "path to an SSH private key")
	fs.BoolVar(&cmd.AllEnvVars, "E", false, "use current environment when possible")
	return fs
}

func (cmd *SetCommand) Run(args []string) {

	// Check the number of args.
	if len(args) < 1 {
		log.Fatal("Specify at least one variable")
	}

	// Create an empty environment map.
	env := make(map[string]string)

	// Iterate through the variables provided.
	for _, variable := range args {

		var value string
		if envValue := os.Getenv(variable); cmd.AllEnvVars && envValue != "" {
			value = envValue
		} else {

			// Get the value from user input.
			inputValue, err := gopass.GetPass(fmt.Sprintf("%s=", variable))
			if err != nil {
				log.Fatal(err.Error())
			}

			value = inputValue
		}

		// Set the variable in the env hash.
		env[variable] = value
	}

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
	client, err := auth.NewClient(auth.WriterUser, cmd.Address, privateKey)
	if err != nil {

		log.Fatal(err.Error())
	}

	for variable, value := range env {

		// Create an environment specific to this variable.
		runEnv := map[string]string{
			"GROUP":  cmd.Group,
			variable: value,
		}

		client.Run(fmt.Sprintf("export %s", variable), runEnv)
	}
}
