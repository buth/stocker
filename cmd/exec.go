package cmd

import (
	"flag"
	"fmt"
	"github.com/buth/stocker/auth"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"strings"
	"syscall"
)

type ExecCommand struct {
	Address, PrivateFilepath, Group, User string
}

func (cmd *ExecCommand) Flags(fs *flag.FlagSet) *flag.FlagSet {
	fs.StringVar(&cmd.Address, "a", ":2022", "address of the stocker server")
	fs.StringVar(&cmd.Group, "g", "default", "group to use for storing and retrieving data")
	fs.StringVar(&cmd.PrivateFilepath, "i", "", "path to an SSH private key")
	fs.StringVar(&cmd.User, "u", "", "user to execute the command as")
	return fs
}

func (cmd *ExecCommand) Run(args []string) {

	// Check the number of args.
	if len(args) < 1 {
		log.Fatal("no command specified")
	}

	// Find the expanded path to cmd.
	command, err := exec.LookPath(args[0])
	if err != nil {
		log.Fatal(fmt.Sprintf("%s: command not found", args[0]))
	}

	// Read the private key from disk if a filepath has been provided.
	var privateKey []byte
	if cmd.PrivateFilepath != "" {
		privateKeyBytes, err := ioutil.ReadFile(cmd.PrivateFilepath)
		if err != nil {
			log.Fatal(err.Error())
		}
		privateKey = privateKeyBytes
	}

	log.Println("PATH", cmd.PrivateFilepath)

	// Get a new client object. If the private key is nil, the method will
	// attempt to use ssh-agent.
	client, err := auth.NewClient(auth.ReaderUser, cmd.Address, privateKey)
	if err != nil {
		log.Fatal(err.Error())
	}

	// Create an environment specific to this variable.
	runEnv := map[string]string{
		"GROUP": cmd.Group,
	}

	stockerEnv, err := client.Run("env", runEnv)
	if err != nil {
		log.Fatal(err.Error())
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
	if cmd.User != "" {

		u, err := user.Lookup(cmd.User)
		if err != nil {
			log.Fatal(err.Error())
		}

		uid, err := strconv.Atoi(u.Uid)
		if err != nil {
			log.Fatal(err.Error())
		}

		if err := syscall.Setuid(uid); err != nil {
			log.Fatal(err.Error())
		}
	}

	// Exec the new command.
	syscall.Exec(command, args, commandEnv)
}
