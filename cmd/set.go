package cmd

import (
	"code.google.com/p/gopass"
	"fmt"
	"github.com/buth/stocker/auth"
	"io/ioutil"
	"os"
)

var Set = &Command{
	UsageLine: "set [options] variable [variable...]",
	Short:     "set the values of the given variables",
}

var setConfig struct {
	Address, Group, PrivateFilepath string
	AllEnvVars                      bool
}

func init() {
	Set.Run = setRun
	Set.Flag.StringVar(&setConfig.Address, "a", ":2022", "address of the stocker server")
	Set.Flag.StringVar(&setConfig.Group, "g", "", "group to use for storing and retrieving data")
	Set.Flag.StringVar(&setConfig.PrivateFilepath, "i", "", "path to an SSH private key")
	Set.Flag.BoolVar(&setConfig.AllEnvVars, "E", false, "use current environment when possible")
}

func setRun(cmd *Command, args []string) {

	// Check the number of args.
	if len(args) < 1 {
		cmd.Usage(2)
	}

	// Create an empty environment map.
	env := make(map[string]string)

	// Iterate through the variables provided.
	for _, variable := range args {

		var value string
		if envValue := os.Getenv(variable); setConfig.AllEnvVars && envValue != "" {
			value = envValue
		} else {

			// Get the value from user input.
			inputValue, err := gopass.GetPass(fmt.Sprintf("%s=", variable))
			if err != nil {
				cmd.Fatal(err.Error())
			}

			value = inputValue
		}

		// Set the variable in the env hash.
		env[variable] = value
	}

	// Read the private key from disk if a filepath has been provided.
	var privateKey []byte
	if setConfig.PrivateFilepath != "" {
		privateKeyBytes, err := ioutil.ReadFile(setConfig.PrivateFilepath)
		if err != nil {
			cmd.Fatal(err.Error())
		}
		privateKey = privateKeyBytes
	}

	// Get a new client object. If the private key is nil, the method will
	// attempt to use ssh-agent.
	client, err := auth.NewClient(auth.WriterUser, setConfig.Address, privateKey)
	if err != nil {
		cmd.Fatal(err.Error())
	}

	for variable, value := range env {

		// Create an environment specific to this variable.
		runEnv := map[string]string{
			"GROUP":  setConfig.Group,
			variable: value,
		}

		client.Run(fmt.Sprintf("export %s", variable), runEnv)
	}
}
