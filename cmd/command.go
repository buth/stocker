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

	// Short is the short description shown in the 'stocker help' output.
	Short, Long string

	// Flag is a set of flags specific to this command.
	Flag flag.FlagSet
}

func (c *Command) Usage(code int) {

	if c.Long != "" {
		fmt.Fprintf(os.Stderr, "%s\n\n", c.Long)
	}

	fmt.Fprintf(os.Stderr, "Usage:\n\n\tstocker %s\n\nOptions:\n", c.UsageLine)

	c.Flag.PrintDefaults()

	os.Exit(code)
}

func (c *Command) Name() string {
	name := c.UsageLine
	i := strings.Index(name, " ")
	if i >= 0 {
		name = name[:i]
	}
	return name
}

func (c *Command) Fatal(message string) {
	fmt.Fprintf(os.Stderr, "%s: %s", c.Name(), message)
	os.Exit(1)
}
