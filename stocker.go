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
	cmd.Server,
}

func Usage(code int) {
	fmt.Fprint(os.Stderr, `Stocker is a tool for managing secure configuration information.

Usage:

	stocker command [arguments]

The commands are:

`)

	// Print each command and its short description.
	for _, command := range commands {
		fmt.Fprintf(os.Stderr, "    %-12s %s\n", command.Name(), command.Short)
	}

	fmt.Fprint(os.Stderr, `
Use "stocker help [topic]" for more information about that topic.

`)

	os.Exit(code)
}

func main() {
	flag.Parse()

	if flag.NArg() == 0 {
		Usage(2)
	}

	if flag.Arg(0) == "help" {
		switch flag.NArg() {
		case 1:
			Usage(0)
		case 2:
			for _, command := range commands {
				if command.Name() == flag.Arg(1) {
					command.Usage(0)
				}
			}
		}
		Usage(2)
	}

	for _, command := range commands {
		if command.Name() == flag.Arg(0) {
			command.Flag.Usage = func() { command.Usage(2) }
			command.Flag.Parse(flag.Args()[1:])
			command.Run(command, command.Flag.Args())
			return
		}
	}
	Usage(2)
}
