package main

import (
	"code.google.com/p/gopass"
	"container/list"
	"encoding/base64"
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
	// "strings"
)

var config struct {
	Group, SecretFilepath, Crypter, Backend, BackendConnectionType, BackendConnectionString string
}

type Stocker struct {
	Env map[string]string
}

// newBackend instantiates a new backend of the chosen type using the
// connection information in config.BackendConnectionType and
// config.BackendConnectionString.
func newBackend() (backend.Backend, error) {
	switch config.Backend {
	case "redis":
		newRedisBackend := redis.New(config.BackendConnectionType, config.BackendConnectionString)
		return newRedisBackend, nil
	}
	return nil, errors.New("no backend selected")
}

func init() {
	flag.StringVar(&config.SecretFilepath, "secret", "", "path to encryption secret")
	flag.StringVar(&config.Backend, "backend", "redis", "backend to use")
	flag.StringVar(&config.BackendConnectionType, "backend-connection-type", "tcp", "backend connection type")
	flag.StringVar(&config.BackendConnectionString, "backend-connection-string", ":6379", "backend connection string")
}

func main() {
	flag.Parse()

	// Check to make sure that a command has been specified.
	if flag.NArg() < 2 {
		flag.Usage()
		return
	}

	// A blank key should be acceptable.
	key := []byte{}

	// Check if we should try to load a key from a file on disk.
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

	// Set the prefix.
	prefix := flag.Arg(0)

	// What are we doing here?
	switch flag.Arg(1) {

	case "set":

		variable := flag.Arg(2)

		value, err := gopass.GetPass(fmt.Sprintf("%s=", flag.Arg(2)))
		if err != nil {
			log.Fatal(err)
		}

		cryptedValue, err := c.EncryptString(value)
		if err != nil {
			log.Fatal(err)
		}

		// Set the key and notify any listeners.
		b.Set(backend.KeyEnv(prefix, variable), cryptedValue)

	case "run":

		// Set up a map for the decrypted values.

		// processedEnv := make([]string, len(config.Env))

		// Create a new run command.
		cmd := exec.Command("docker", flag.Args()[1:]...)

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

		decryptedEnv := list.New()
		// Loop through the remaining arguments looking for possible
		// environment settings.
		for i := 2; i < flag.NArg(); i++ {

			if flag.Arg(i) == "-e" && i+1 < flag.NArg() {

				variable := flag.Arg(i + 1)

				// Set the key to use with the backend.
				key := backend.KeyEnv(prefix, variable)

				cryptedValue, err := b.Get(key)
				if err != nil {
					log.Fatal(err)
				}

				decryptedValue, err := c.DecryptString(cryptedValue)
				if err != nil {
					log.Fatal(err)
				}

				// Format the statement.
				statement := fmt.Sprintf("%s=%s", variable, decryptedValue)

				// Add the statement to the list.
				decryptedEnv.PushBack(statement)
			}
		}

		// Create a slice of strings large enough to contain both the os
		// environement and the decrypted environment.
		cmd.Env = make([]string, len(os.Environ()), len(os.Environ())+decryptedEnv.Len())
		copy(cmd.Env, os.Environ())

		for e := decryptedEnv.Front(); e != nil; e = e.Next() {
			i := len(cmd.Env)
			cmd.Env = cmd.Env[:i+1]
			cmd.Env[i] = e.Value.(string)
		}

		// Run the command.
		if err := cmd.Start(); err != nil {
			log.Fatal(err)
		}

		// Handle signalling.

		ch := make(chan os.Signal)
		go func() {
			for {
				sig, ok := <-ch
				if !ok {
					return
				}

				fmt.Println("GO", sig)

				// Forward the signal to the command process.
				cmd.Process.Signal(sig)

			}
		}()

		signal.Notify(ch)

		// Wait for it to exit.
		cmd.Wait()
		signal.Stop(ch)
		close(ch)

	case "key":
		fmt.Println(base64.StdEncoding.EncodeToString(key))
	}
}
