package group

import (
	// "exec"
	"github.com/buth/stocker/stocker/backend"
	"github.com/buth/stocker/stocker/container"
	"github.com/buth/stocker/stocker/crypto"
	// "log"
	// "strings"
	"errors"
	"sync"
)

type group struct {
	name    string
	state   map[string]container
	backend *backend.Backend
	crypter *crypto.Crypter
	mu      *sync.Mutex
}

func New(name string, b *backend.Backend, c *crypto.Crypter) *group {
	return &group{name: name, backend: b, crypter: c}
}

// add adds the named container to the state. If the named container exists,
// its configuration is updated.
func (g *group) add(name, command) error {

}

// remove removes the named container from the state, shutting it down if it
// is present.
func (g *group) remove(name) error {

	// Check if we know about this container.
	if c, ok := g.state[name]; ok {
		defer delete(g.state, name)
		return c.Stop()
	}

	return errors.New("Can't remove a container we don't know about.")
}

// update reloads the named container but does not update its configuration.
func (g *group) update(name) error {

	// Check if we know about this container.
	if c, ok := g.state[name]; ok {
		return c.Update()
	}

	return errors.New("Can't update a container we don't know about.")
}

func (g *group) handleMessage(channel, message string) error {

	components := backend.DecomposeKey(channel)

	// Every channel should have at least 3 components.
	if len(components) < 3 {
		return errors.New("Recieved a short broadcast.")
	}

	// We want to insure that all actions state modifications are atomic and
	// that we release the lock in the case of any error.
	mu.Lock()
	defer mu.Unlock()

	switch components[1] {
	case "add":
		return g.add(components[2], message)
	case "remove":
		return g.remove(components[2])
	case "update":

		if c, ok := g.state[components[2]]; ok {
			return c.Update()
		} else {
			return errors.New("Recieved a short broadcast.")
		}

		return g.update(components[2])
	}

	return errors.New("Recieved a broadcast with an unknown command.")
}

func (g *group) Subscribe() {
	g.backend.Subscribe(backend.Key("broadcast", g.name), g.handleMessage)
}

func (g *group) Unsubscribe() {
	g.backend.Unsubscribe(backend.Key("broadcast", g.name))
}
