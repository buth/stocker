package cmd

import (
	"github.com/buth/stocker/crypto"
	"log"
)

var Key = &Command{
	UsageLine: "key filename",
	Short:     "create a key saved at the given filename",
}

func init() {
	Key.Run = keyRun
}

func keyRun(cmd *Command, args []string) {

	// Check the number of args.
	if len(args) != 1 {
		cmd.Usage(2)
	}

	// Set the filename.
	filename := args[0]

	// Create a random crypter object.
	c, err := crypto.NewRandomCrypter()
	if err != nil {
		log.Fatal(err)
	}

	// Write out the key to the given filename.
	if err := c.ToFile(filename); err != nil {
		log.Fatal(err)
	}
}
