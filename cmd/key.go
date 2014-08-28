package cmd

import (
	"flag"
	"fmt"
	"github.com/buth/stocker/crypto"
	"log"
)

type KeyCommand struct {
	filename string
}

// Flags defines the command-specific flags.
func (cmd *KeyCommand) Flags(fs *flag.FlagSet) *flag.FlagSet {
	fs.StringVar(&cmd.filename, "f", "", "filename to save the key to")
	return fs
}

func (cmd *KeyCommand) Run(args []string) {

	// Create a random crypter object.
	c, err := crypto.NewRandomCrypter()
	if err != nil {
		log.Fatal(err)
	}

	if cmd.filename != "" {

		// Write out the key to the given filename.
		if err := c.ToFile(cmd.filename); err != nil {
			log.Fatal(err)
		}
	} else {

		// Just get the string.
		keyString, err := c.ToString()
		if err != nil {
			log.Fatal(err)
		}

		// Print the string to STDOUT.
		fmt.Print(keyString)
	}
}
