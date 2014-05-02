package cmd

import (
	"code.google.com/p/gopass"
	"fmt"
	"github.com/buth/stocker/backend"
	"github.com/buth/stocker/crypto"
	"os"
)

var Set = &Command{
	UsageLine: "set [OPTIONS] VARIABLE [VARIABLE...]",
	Short:     "set the values of the given variables",
}

var setConfig struct {
	SecretFilepath, Backend, BackendProtocol, BackendHost, Group string
}

func init() {
	Set.Run = setRun
	Set.Flag.StringVar(&setConfig.SecretFilepath, "k", "", "path to encryption key")
	Set.Flag.StringVar(&setConfig.Backend, "b", "redis", "backend to use")
	Set.Flag.StringVar(&setConfig.BackendProtocol, "t", "tcp", "backend connection protocol")
	Set.Flag.StringVar(&setConfig.BackendHost, "h", ":6379", "backend connection host (optionally including port)")
	Set.Flag.StringVar(&setConfig.Group, "g", "", "group to use for storing and retrieving data")
}

func setRun(cmd *Command, args []string) {

	// Check the number of args.
	if len(args) < 1 {
		cmd.Usage()
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

	b, err := backend.NewBackend(setConfig.Backend, setConfig.BackendProtocol, setConfig.BackendHost)
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
