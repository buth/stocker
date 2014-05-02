package cmd

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

// This is a modified verson of the Comand type from the Go source code.
// http://golang.org/src/cmd/go/main.go
type Command struct {
	// Run runs the command.
	// The args are the arguments after the command name.
	Run func(cmd *Command, args []string)

	// UsageLine is the one-line usage message.
	// The first word in the line is taken to be the command name.
	UsageLine string

	// Short is the short description shown in the 'go help' output.
	Short string

	// Flag is a set of flags specific to this command.
	Flag flag.FlagSet
}

func (c *Command) Usage() {
	fmt.Fprintf(os.Stderr, "Usage:\n\n\tstocker %s\n\n", c.UsageLine)
	if c.Flag.NFlag() > 0 {
		c.Flag.Usage()
	} else {
		os.Exit(2)
	}
}

func (c *Command) Name() string {
	name := c.UsageLine
	i := strings.Index(name, " ")
	if i >= 0 {
		name = name[:i]
	}
	return name
}
