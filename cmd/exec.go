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
	SecretFilepath, Backend, BackendNamespace, BackendProtocol, BackendAddress, Group, User string
	EnvVars                                                                                 StringAcumulator
	AllEnvVars                                                                              bool
}

func init() {
	Exec.Run = execRun
	Exec.Flag.StringVar(&execConfig.Backend, "b", "redis", "backend to use")
	Exec.Flag.StringVar(&execConfig.BackendAddress, "h", ":6379", "backend address")
	Exec.Flag.StringVar(&execConfig.BackendNamespace, "n", "stocker", "backend namespace")
	Exec.Flag.StringVar(&execConfig.BackendProtocol, "t", "tcp", "backend connection protocol")
	Exec.Flag.StringVar(&execConfig.Group, "g", "", "group to use for storing and retrieving data")
	Exec.Flag.StringVar(&execConfig.SecretFilepath, "k", "/etc/stocker/key", "path to encryption key")
	Exec.Flag.StringVar(&execConfig.User, "u", "", "user to execute the command as")
	Exec.Flag.Var(&execConfig.EnvVars, "e", "environment variable to fetch")
	Exec.Flag.BoolVar(&execConfig.AllEnvVars, "E", false, "fetch all environment variables")
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

	b, err := backend.NewBackend(execConfig.Backend, execConfig.BackendNamespace, execConfig.BackendProtocol, execConfig.BackendAddress)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Find the expanded path to cmd.
	command, err := exec.LookPath(args[0])
	if err != nil {
		log.Fatalf("%s: command not found", args[0])
	}

	// Create a map of environment variables to be passed to cmd and
	// initialize it with the current environment.
	env := make(map[string]string)
	for _, variable := range os.Environ() {
		components := strings.Split(variable, "=")
		env[components[0]] = components[1]
	}

	// Create a new map to store crypted values to later write into the
	// evironment variables map.
	cryptedEnv := make(map[string]string)

	// Check if we are supposed to pull all availble environement variables
	// for the given group.
	if execConfig.AllEnvVars {

		variables, err := b.GetGroup(execConfig.Group)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		for variable, cryptedValue := range variables {
			cryptedEnv[variable] = cryptedValue
		}
	}

	// Handle individually specified environment variables. This can be done
	// in conjunction with the all environment variables in order to garauntee
	// the presense of certain values.
	for _, variable := range execConfig.EnvVars {

		if _, ok := cryptedEnv[variable]; !ok {

			cryptedValue, err := b.GetVariable(execConfig.Group, variable)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			cryptedEnv[variable] = cryptedValue
		}
	}

	// Decode the crypted values map.
	for variable, cryptedValue := range cryptedEnv {

		// Try to decrypt the value.
		decryptedValue, err := c.DecryptString(cryptedValue)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Set the decrypted value overwriting any existing value.
		env[variable] = decryptedValue
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
	syscall.Exec(command, args, commandEnv)
}
