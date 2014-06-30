package cmd

import (
	"fmt"
	"github.com/buth/stocker/auth"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"strings"
	"syscall"
)

var Exec = &Command{
	UsageLine: "exec [options] command [argument...]",
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
	Address, PrivateFilepath, Group, User string
}

func init() {
	Exec.Run = execRun
	Exec.Flag.StringVar(&execConfig.Address, "a", ":2022", "address of the stocker server")
	Exec.Flag.StringVar(&execConfig.Group, "g", "", "group to use for storing and retrieving data")
	Exec.Flag.StringVar(&execConfig.PrivateFilepath, "i", "", "path to an SSH private key")
	Exec.Flag.StringVar(&execConfig.User, "u", "", "user to execute the command as")
}

func execRun(cmd *Command, args []string) {

	// Check the number of args.
	if len(args) < 1 {
		cmd.Usage(2)
	}

	// Find the expanded path to cmd.
	command, err := exec.LookPath(args[0])
	if err != nil {
		cmd.Fatal(fmt.Sprintf("%s: command not found", args[0]))
	}

	// Read the private key from disk if a filepath has been provided.
	var privateKey []byte
	if execConfig.PrivateFilepath != "" {
		privateKeyBytes, err := ioutil.ReadFile(execConfig.PrivateFilepath)
		if err != nil {
			cmd.Fatal(err.Error())
		}
		privateKey = privateKeyBytes
	}

	// Get a new client object. If the private key is nil, the method will
	// attempt to use ssh-agent.
	client, err := auth.NewClient(auth.ReaderUser, execConfig.Address, privateKey)
	if err != nil {
		cmd.Fatal(err.Error())
	}

	// Create an environment specific to this variable.
	runEnv := map[string]string{
		"GROUP": execConfig.Group,
	}

	stockerEnv, err := client.Run("env", runEnv)
	if err != nil {
		cmd.Fatal(err.Error())
	}

	// Create a map of environment variables to be passed to cmd and
	// initialize it with the current environment.
	env := make(map[string]string)
	for _, variable := range os.Environ() {
		components := strings.Split(variable, "=")
		env[components[0]] = components[1]
	}

	// Parse the stocker environment and save it into the env.
	pairs := strings.Split(stockerEnv, "\n")
	for _, pair := range pairs {
		components := strings.SplitN(pair, "=", 2)
		if len(components) == 2 {
			env[components[0]] = components[1]
		}
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
			cmd.Fatal(err.Error())
		}

		uid, err := strconv.Atoi(u.Uid)
		if err != nil {
			cmd.Fatal(err.Error())
		}

		if err := syscall.Setuid(uid); err != nil {
			cmd.Fatal(err.Error())
		}
	}

	// Exec the new command.
	syscall.Exec(command, args, commandEnv)
}
