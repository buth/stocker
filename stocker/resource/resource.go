package resource

import (
	// "bytes"
	"github.com/dotcloud/docker"
	dockerclient "github.com/fsouza/go-dockerclient"
	// "log"
	"strings"
	"sync"
	// "time"
)

type Resource interface {
	SetCommand(command string) error
	Reload() error
}

type resource struct {
	name, command string
	client        *dockerclient.Client
	signalMu      sync.Mutex
}

func New(name, command string) (*resource, error) {

	newResource := &resource{
		name:    name,
		command: command,
	}

	if client, err := dockerclient.NewClient("unix:///var/run/docker.sock"); err != nil {
		return newResource, err
	} else {
		newResource.client = client
	}

	return newResource, nil
}

func (c *resource) SetCommand(command string) error {

	if c.command != command {
		c.command = command
		return c.Reload()
	}

	return nil
}

// Reload will update the image corresponding to a run command,
// elimnate the related resource – if it exists – and start a new one.
func (c *resource) Reload() error {

	c.signalMu.Lock()
	defer c.signalMu.Unlock()

	// Check the current status of the resource.
	if container, err := c.client.InspectContainer(c.name); err != nil {

		// The only error that's not a problem is NoSuchContainer. Anything else
		// and we should quit.
		if _, ok := err.(*dockerclient.NoSuchContainer); !ok {
			return err
		}
	} else {

		// Check if the container is running.
		if container.State.Running {

			// Try and stop the container.
			if err := c.client.StopContainer(c.name, 60); err != nil {
				return err
			}
		}

		// Try and remove the resource.
		if err := c.client.RemoveContainer(c.name); err != nil {
			return err
		}
	}

	// Try to parse the given configuration.
	config, hostConfig, _, err := docker.ParseRun(strings.Split(c.command, " "), nil)
	if err != nil {
		return err
	}

	// Note: This would be where we would pull any updates to the conatiner
	// image, but as this isn't working with unix socket client connections at
	// the moment, it's left out.

	if _, err := c.client.CreateContainer(dockerclient.CreateContainerOptions{Name: c.name}, config); err != nil {
		return err
	}

	if err := c.client.StartContainer(c.name, hostConfig); err != nil {
		return err
	}

	return nil
}
