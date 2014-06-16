package cmd

import (
	"code.google.com/p/gopass"
	"fmt"
	"github.com/buth/stocker/backend"
	"github.com/buth/stocker/crypto"
	"os"
)

var Set = &Command{
	UsageLine: "set [options] variable [variable...]",
	Short:     "set the values of the given variables",
}

var setConfig struct {
	SecretFilepath, Backend, BackendNamespace, BackendProtocol, BackendAddress, Group string
	AllEnvVars                                                                        bool
}

func init() {
	Set.Run = setRun
	Set.Flag.StringVar(&setConfig.Backend, "b", "redis", "backend to use")
	Set.Flag.StringVar(&setConfig.BackendAddress, "h", ":6379", "backend address")
	Set.Flag.StringVar(&setConfig.BackendNamespace, "n", "stocker", "backend namespace")
	Set.Flag.StringVar(&setConfig.BackendProtocol, "t", "tcp", "backend connection protocol")
	Set.Flag.StringVar(&setConfig.Group, "g", "", "group to use for storing and retrieving data")
	Set.Flag.StringVar(&setConfig.SecretFilepath, "k", "/etc/stocker/key", "path to encryption key")
	Set.Flag.BoolVar(&setConfig.AllEnvVars, "E", false, "use current environment when possible")
}

func setRun(cmd *Command, args []string) {

	// Check the number of args.
	if len(args) < 1 {
		cmd.Usage(2)
	}

	key, err := crypto.NewKeyFromFile(setConfig.SecretFilepath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	c, err := crypto.NewCrypter(key)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	b, err := backend.NewBackend(setConfig.Backend, setConfig.BackendNamespace, setConfig.BackendProtocol, setConfig.BackendAddress)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Iterate through the variables provided.
	for _, variable := range args {

		value, err := gopass.GetPass(fmt.Sprintf("%s=", variable))
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		cryptedValue, err := c.EncryptString(value)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Set the key and notify any listeners.
		b.SetVariable(setConfig.Group, variable, cryptedValue)
	}
}
