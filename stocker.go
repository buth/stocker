package main

import (
	"code.google.com/p/gopass"
	"errors"
	"flag"
	"fmt"
	"github.com/buth/stocker/backend"
	"github.com/buth/stocker/backend/redis"
	"github.com/buth/stocker/crypto"
	"log"
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"strings"
	"syscall"
)

type StringAcumulator []string

func (s *StringAcumulator) Set(value string) error {
	*s = append(*s, value)
	return nil
}

func (s *StringAcumulator) String() string {
	return fmt.Sprintf("%s", *s)
}

var config struct {
	SecretFilepath, Backend, BackendProtocol, BackendHost, User, Group string
	EnvVars                                                            StringAcumulator
}

// newBackend instantiates a new backend of the chosen type using the
// connection information in config.BackendProtocol and
// config.BackendHost.
func newBackend() (backend.Backend, error) {
	switch config.Backend {
	case "redis":
		newRedisBackend := redis.New(config.BackendProtocol, config.BackendHost)
		return newRedisBackend, nil
	}
	return nil, errors.New("no backend selected")
}

// usage prints a usage statment for stocker.
func usage(code int) {
	fmt.Println("Usage: stocker [OPTIONS] COMMAND NAME [ARG...]")
	flag.PrintDefaults()
	os.Exit(code)
}

func init() {
	flag.StringVar(&config.SecretFilepath, "k", "", "path to encryption key")
	flag.StringVar(&config.Backend, "b", "redis", "backend to use")
	flag.StringVar(&config.BackendProtocol, "t", "tcp", "backend connection protocol")
	flag.StringVar(&config.BackendHost, "h", ":6379", "backend connection host (optionally including port)")
	flag.StringVar(&config.User, "u", "", "username of the user to run the command as")
	flag.StringVar(&config.Group, "g", "", "group to run command as")
	flag.Var(&config.EnvVars, "e", "environment variables")
}

func main() {
	flag.Parse()

	// Check to make sure that a command has been specified.
	if flag.NArg() < 1 {
		usage(1)
	}

	// Check if we should try to load a key from a file on disk. If a path was
	// not provided, generate a new key.
	var key crypto.Key
	if config.SecretFilepath != "" {

		// Try to create a new key from the given file path.
		keyFromFile, err := crypto.NewKeyFromFile(config.SecretFilepath)
		if err != nil {
			log.Fatalln(err)
		}

		key = keyFromFile
	} else {
		key = crypto.NewKey()
	}

	// Attempt to load a crypter.
	c, err := crypto.NewCrypter(key)
	if err != nil {
		log.Fatalln(err)
	}

	// Attempt to load a backend.
	b, err := newBackend()
	if err != nil {
		log.Fatalln(err)
	}

	// What are we doing here?
	switch flag.Arg(0) {

	default:
		usage(1)

	case "help":
		usage(0)

	case "key":
		fmt.Print(key)

	case "set":

		// Check to make sure there is a prefix.
		if flag.NArg() != 2 {
			usage(1)
		}

		// Set the prefix.
		prefix := flag.Arg(1)

		// Iterate through the variables provided.
		for _, variable := range config.EnvVars {

			value, err := gopass.GetPass(fmt.Sprintf("%s=", variable))
			if err != nil {
				log.Fatal(err)
			}

			cryptedValue, err := c.EncryptString(value)
			if err != nil {
				log.Fatal(err)
			}

			fmt.Println(backend.KeyEnv(prefix, variable))

			// Set the key and notify any listeners.
			b.Set(backend.KeyEnv(prefix, variable), cryptedValue)
		}

	case "exec":

		// Check to make sure there is an app name and a command.
		if flag.NArg() < 3 {
			usage(1)
		}

		// Find the expanded path to cmd.
		cmd, err := exec.LookPath(flag.Arg(2))
		if err != nil {
			log.Fatalf("%s: command not found", flag.Arg(2))
		}

		// Set the prefix.
		prefix := flag.Arg(1)

		// Set the args.
		args := flag.Args()[2:]

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
		for _, variable := range config.EnvVars {

			fmt.Println(backend.KeyEnv(prefix, variable))

			// Set the key to use with the backend.
			key := backend.KeyEnv(prefix, variable)
			value := os.Getenv(variable)

			// Check if we should search for a value.
			if value == "" {

				cryptedValue, err := b.Get(key)
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
		envv := make([]string, 0, len(env))
		for key, value := range env {
			envv = envv[:len(envv)+1]
			envv[len(envv)-1] = fmt.Sprintf("%s=%s", key, value)
		}

		// Handle user.
		if config.User != "" {

			u, err := user.Lookup(config.User)
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
		syscall.Exec(cmd, args, envv)
	}
}
