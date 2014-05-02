package main

import (
	"flag"
	"fmt"
	"github.com/buth/stocker/cmd"
	"os"
)

var commands = []*cmd.Command{
	cmd.Key,
	cmd.Set,
	cmd.Exec,
}

func main() {
	flag.Parse()

	// Check to make sure that a command has been specified.
	if flag.NArg() < 1 {
		fmt.Print("Stocker is a tool for managing secure configuration information.\n\nUsage:\n\n\tstocker COMMAND [ARG...]\n\nThe commands are:\n\n")
		for _, command := range commands {
			fmt.Printf("\t%s\t%s\n", command.Name(), command.Short)
		}
		fmt.Print("\n")
		os.Exit(2)
	}

	// Iterate through the commands.
	for _, command := range commands {
		if command.Name() == flag.Arg(0) && command.Run != nil {

			// Parse everything but the name of the command itself.
			command.Flag.Parse(flag.Args()[1:])

			// Check the args.
			args := command.Flag.Args()

			// Run the Command!
			command.Run(command, args)
			return
		}
	}
}
