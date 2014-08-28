package main

import (
	"github.com/buth/stocker/cmd"
	"github.com/rakyll/command"
)

func main() {
	command.On("key", "create a new private key", &cmd.KeyCommand{}, []string{})
	command.On("server", "start a stocker daemon", &cmd.ServerCommand{}, []string{})
	command.On("set", "set a group's environment", &cmd.SetCommand{}, []string{})
	command.On("exec", "run a command with a group's environment", &cmd.ExecCommand{}, []string{})
	command.On("register", "run a command with a group's environment", &cmd.RegisterCommand{}, []string{})

	command.Parse()
	command.Run()
}
