package cmd

import (
	"fmt"
	"github.com/buth/stocker/backend"
	"github.com/buth/stocker/crypto"
	"log"
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"strings"
	"syscall"
)

var Exec = &Command{
	UsageLine: "exec [OPTIONS] COMMAND [ARG...]",
	Short:     "execute a command with the given environment",
}

type StringAcumulator []string

func (s *StringAcumulator) Set(value string) error {
	*s = append(*s, value)
	return nil
}

func (s *StringAcumulator) String() string {
	return fmt.Sprintf("%s", *s)
}

var execConfig struct {
	SecretFilepath, Backend, BackendProtocol, BackendHost, Group, User string
	EnvVars                                                            StringAcumulator
}

func init() {
	Exec.Run = execRun
	Exec.Flag.StringVar(&execConfig.SecretFilepath, "k", "", "path to encryption key")
	Exec.Flag.StringVar(&execConfig.Backend, "b", "redis", "backend to use")
	Exec.Flag.StringVar(&execConfig.BackendProtocol, "t", "tcp", "backend connection protocol")
	Exec.Flag.StringVar(&execConfig.BackendHost, "h", ":6379", "backend connection host (optionally including port)")
	Exec.Flag.StringVar(&execConfig.Group, "g", "", "group to use for storing and retrieving data")
	Exec.Flag.StringVar(&execConfig.User, "u", "", "user to execute the command as")
	Exec.Flag.Var(&execConfig.EnvVars, "e", "environment variables")
}

func execRun(cmd *Command, args []string) {

	// Check the number of args.
	if len(args) < 1 {
		cmd.Usage()
	}

	key, err := crypto.NewKeyFromFile(execConfig.SecretFilepath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	c, err := crypto.NewCrypter(key)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	b, err := backend.NewBackend(execConfig.Backend, execConfig.BackendProtocol, execConfig.BackendHost)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Find the expanded path to cmd.
	command, err := exec.LookPath(args[0])
	if err != nil {
		log.Fatalf("%s: command not found", args[0])
	}

	// Set the args.
	commandArgs := args[1:]

	// Create a map of environment variables to be passed to cmd and
	// initialize it with the current environment.
	env := make(map[string]string)
	for _, variable := range os.Environ() {
		components := strings.Split(variable, "=")
		env[components[0]] = components[1]
	}

	// Loop through the provided environment variables, looking for values
	// first in the environment, and secondarally in the backend store.
	// All errors are fatal.
	for _, variable := range execConfig.EnvVars {

		// Set the key to use with the backend.
		value := os.Getenv(variable)

		// Check if we should search for a value.
		if value == "" {

			cryptedValue, err := b.GetVariable(execConfig.Group, variable)
			if err != nil {
				log.Fatalf("%s: %s", variable, err)
			}

			decryptedValue, err := c.DecryptString(cryptedValue)
			if err != nil {
				log.Fatalf("%s: %s", variable, err)
			}

			value = decryptedValue
		}

		// Format the statement.
		env[variable] = value
	}

	// Create a list of environment key/value pairs and write the
	// flattened environment variables map to it.
	commandEnv := make([]string, 0, len(env))
	for key, value := range env {
		commandEnv = commandEnv[:len(commandEnv)+1]
		commandEnv[len(commandEnv)-1] = fmt.Sprintf("%s=%s", key, value)
	}

	// Handle user.
	if execConfig.User != "" {

		u, err := user.Lookup(execConfig.User)
		if err != nil {
			log.Fatal(err)
		}

		uid, err := strconv.Atoi(u.Uid)
		if err != nil {
			log.Fatal(err)
		}

		if err := syscall.Setuid(uid); err != nil {
			log.Fatal(err)
		}
	}

	// Exec the new command.
	syscall.Exec(command, commandArgs, commandEnv)
}
