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

	// Create a new key and attempt to write it to a file.
	key := crypto.NewKey()
	if err := key.ToFile(filename); err != nil {
		log.Fatal(err)
	}
}
