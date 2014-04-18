package main

import (
	"code.google.com/p/gopass"
	"errors"
	"flag"
	"fmt"
	"github.com/buth/stocker/backend"
	"github.com/buth/stocker/backend/redis"
	"github.com/buth/stocker/crypto"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
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
	SecretFilepath, Backend, BackendProtocol, BackendHost string
	EnvVars                                               StringAcumulator
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

			// Set the key and notify any listeners.
			b.Set(backend.KeyEnv(prefix, variable), cryptedValue)
		}

	case "exec":

		// Check to make sure there is an app name and a command.
		if flag.NArg() < 3 {
			usage(1)
		}

		// Set the prefix.
		prefix := flag.Arg(1)

		// Create a new run command.
		cmd := exec.Command(flag.Arg(2), flag.Args()[3:]...)

		// Setup the stdout pipe.
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			log.Fatal(err)
		}
		go io.Copy(os.Stdout, stdout)

		// Setup the stderr pipe.
		stderr, err := cmd.StderrPipe()
		if err != nil {
			log.Fatal(err)
		}
		go io.Copy(os.Stderr, stderr)

		// Setup the stderr pipe.
		stdin, err := cmd.StdinPipe()
		if err != nil {
			log.Fatal(err)
		}
		go io.Copy(stdin, os.Stdin)

		// Create a list of environment variables for the command itself.
		cmd.Env = make([]string, len(config.EnvVars))

		// Loop through the provided environment variables, looking for values
		// first in the environment, and secondarally in the backend store.
		// All errors are fatal.
		for i, variable := range config.EnvVars {

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
			cmd.Env[i] = fmt.Sprintf("%s=%s", variable, value)
		}

		// Run the command.
		if err := cmd.Start(); err != nil {
			log.Fatal(err)
		}

		// Handle signalling.
		ch := make(chan os.Signal, 1)
		go func() {
			for {

				// Wait for a signal; exit if the channel is closed.
				sig, ok := <-ch
				if !ok {
					return
				}

				// Forward the signal to the command process.
				cmd.Process.Signal(sig)
			}
		}()

		// Set the channel for notifications. We're sending along all signals.
		signal.Notify(ch)

		// Wait for it to exit.
		cmd.Wait()

		// Stop sending signal notifications to the channel.
		signal.Stop(ch)

		// Close the channel to tell the go routine to exit.
		close(ch)
	}
}
