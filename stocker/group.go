package stocker

import (
	"errors"
	"github.com/buth/stocker/stocker/backend"
	"github.com/buth/stocker/stocker/crypto"
	"github.com/dotcloud/docker"
	dockerclient "github.com/fsouza/go-dockerclient"
	"log"
	"os/exec"
	"strings"
	"sync"
)

type group struct {
	name        string
	resourcesMu sync.Mutex
	resources   map[string]*sync.Mutex
	backend     backend.Backend
	crypter     crypto.Crypter
	client      *dockerclient.Client
}

func NewGroup(name string, b backend.Backend, c crypto.Crypter) (*group, error) {
	g := &group{
		name:      name,
		backend:   b,
		crypter:   c,
		resources: make(map[string]*sync.Mutex),
	}

	if client, err := dockerclient.NewClient("unix:///var/run/docker.sock"); err != nil {
		return g, err
	} else {
		g.client = client
	}

	names, err := g.backend.List(backend.Key("conf", g.name, "resources"))
	if err != nil {
		return g, err
	}

	for _, name := range names {
		if err := g.reloadResource(name); err != nil {
			return g, err
		}
	}

	return g, nil
}

func (g *group) reloadResource(name string) error {
	log.Printf("i\t%s\t%s\tReloading\n", g.name, name)

	// Get the resources lock and, check if we need to create a new mutex for
	// this resource.
	g.resourcesMu.Lock()
	if _, ok := g.resources[name]; !ok {
		g.resources[name] = &sync.Mutex{}
	}
	g.resourcesMu.Unlock()

	// Get the lock for this specific resource.
	g.resources[name].Lock()
	defer g.resources[name].Unlock()

	// Try to get the command from the backend.
	command, err := g.backend.Get(backend.Key("conf", g.name, "resource", name))
	if err != nil {
		return err
	}

	// Try to parse the given configuration.
	config, host, _, err := docker.ParseRun(strings.Split(command, " "), nil)
	if err != nil {
		return err
	}

	// TODO: Replace this hack with the client code that follows it.
	if err := exec.Command("docker", "pull", config.Image).Run(); err != nil {
		return err
	}

	// This is the prefered method, but it isn't working at the moment because
	// the client does not support this operation when using unix socket
	// connectiong.

	// if err := c.client.PullImage(dockerclient.PullImageOptions{Repository: config.Image}, nil); err != nil {
	//  return err
	// }

	// Check the current status of the container that corresponds to this
	// resource, using the resouce name as the container ID.
	if container, err := g.client.InspectContainer(name); err != nil {

		// The only error that's not a problem is NoSuchContainer. Anything else
		// and we should quit.
		if _, ok := err.(*dockerclient.NoSuchContainer); !ok {
			return err
		}

		// Now we are assuming the container does not exist and we need to create
		// it. This should insure that the resulting container is as up to date as
		// possible.

		// Create a new container with the given configuration. This doesn't start
		// it up yet.
		log.Printf("i\t%s\t%s\tCreating a new container\n", g.name, name)
		if _, err := g.client.CreateContainer(dockerclient.CreateContainerOptions{Name: name}, config); err != nil {
			return err
		}
	} else {

		// Now we are assuming the container does exist and we need to check if it
		// is up to date.
		image, err := g.client.InspectImage(config.Image)
		if err != nil {
			return err
		}

		if ConfigsEqual(container.Config, config) && container.Image == image.ID {
			log.Printf("i\t%s\t%s\tContainer is already up to date\n", g.name, name)

			// Check if our work is done.
			if container.State.Running {
				log.Printf("i\t%s\t%s\tContainer already running\n", g.name, name)
				return nil
			}
		} else {

			// Check if the container is running.
			if container.State.Running {

				// Try and stop the container.
				log.Printf("i\t%s\t%s\tStopping container\n", g.name, name)
				if err := g.client.StopContainer(name, 60); err != nil {
					return err
				}
			}

			// Try and remove the resource.
			log.Printf("i\t%s\t%s\tRemoving container\n", g.name, name)
			if err := g.client.RemoveContainer(name); err != nil {
				return err
			}

			// Create a new container with the given configuration. This doesn't start
			// it up yet.
			log.Printf("i\t%s\t%s\tCreating a new container\n", g.name, name)
			if _, err := g.client.CreateContainer(dockerclient.CreateContainerOptions{Name: name}, config); err != nil {
				return err
			}
		}
	}

	// Start the container we just created.
	log.Printf("i\t%s\t%s\tStarting container\n", g.name, name)
	if err := g.client.StartContainer(name, host); err != nil {
		return err
	}

	return nil
}

func (g *group) handleMessage(channel, message string) error {

	components := backend.DecomposeKey(channel)

	// Every channel should have at least 3 componentg.
	if len(components) < 3 {
		return errors.New("Recieved a short broadcast.")
	}

	name := components[2]

	switch message {

	case "remove":
		// return g.remove(components[2])

	case "reload":

		if err := g.reloadResource(name); err != nil {
			return err
		}

	default:
		return errors.New("Recieved a broadcast with an unknown command.")
	}

	return nil
}

func (g *group) Run() error {
	return g.backend.Subscribe(backend.Key("cast", g.name, "*"), g.handleMessage)
}
