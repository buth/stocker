package main

import (
	"code.google.com/p/gopass"
	"encoding/base64"
	"errors"
	"flag"
	"fmt"
	"github.com/buth/stocker/backend"
	"github.com/buth/stocker/backend/redis"
	"github.com/buth/stocker/crypto"
	"github.com/dotcloud/docker/pkg/sysinfo"
	"github.com/dotcloud/docker/runconfig"
	"log"
	"os"
	"strings"
)

var config struct {
	Group, SecretFilepath, Crypter, Backend, BackendConnectionType, BackendConnectionString string
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

	// A blank key should be acceptable.
	key := []byte{}

	// Check if we should try to load a key from a file on disk.
	if config.SecretFilepath != "" {
		if keyFromFile, err := crypto.NewKeyFromFile(config.SecretFilepath); err != nil {
			log.Fatalln(err)
		} else {
			key = keyFromFile
		}
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

	case "set":

		prefix := flag.Arg(1)
		variable := flag.Arg(2)

		value, err := gopass.GetPass(fmt.Sprintf("%s=", flag.Arg(2)))
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(value)

		cryptedValue, err := c.EncryptString(value)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(cryptedValue)

		key := fmt.Sprintf("stocker/%s/env/%s", prefix, variable)
		listener := fmt.Sprintf("stocker/%s/signal/%s", prefix, variable)

		// Set the key and notify any listeners.
		b.Set(key, cryptedValue)
		b.Publish(listener, cryptedValue)

	case "run":

		if flag.NArg() < 3 {
			log.Fatal("run requires a group and resource name!")
		}

		log.Println(c, b)

		prefix := flag.Arg(1)

		config, hostConfig, _, err := runconfig.Parse(flag.Args()[2:], sysinfo.New(true)) //(*Config, *HostConfig, *flag.FlagSet, error)

		log.Println(config, hostConfig, err)

		log.Println(config.Image)

		log.Println(config.ExposedPorts)

		log.Println(config.Cmd)

		log.Println(config.Env)

		processedEnv := make([]string, len(config.Env))

		for i, env := range config.Env {

			components := strings.Split(env, "=")
			variable := components[0]
			value := components[1]

			if value != "" {

				// A value was specified explicitly on the command line, so
				// let's just use that.
				processedEnv[i] = env
			} else if osEnvValue := os.Getenv(variable); osEnvValue != "" {

				// A value was available in the environment.
				processedEnv[i] = fmt.Sprintf("%s=%s", variable, osEnvValue)
			} else {

				// No value was given or available in the evironment so let's
				// assume that we should try to pull a secure value from the
				// store.

				key := fmt.Sprintf("stocker/%s/env/%s", prefix, variable)
				// listener := fmt.Sprintf("stocker/%s/signal/%s", prefix, variable)

				cryptedValue, err := b.Get(key)
				if err != nil {
					log.Println(err)
					// handle
				}

				decryptedValue, err := c.DecryptString(cryptedValue)
				if err != nil {
					log.Println(err)
					// handle
				}

				processedEnv[i] = fmt.Sprintf("%s=%s", variable, decryptedValue)
			}
		}

		config.Env = processedEnv

		fmt.Println(config.Env)

	case "key":
		fmt.Println(base64.StdEncoding.EncodeToString(key))
	}
}
