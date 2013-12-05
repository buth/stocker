package stocker

import (
	"errors"
	"github.com/buth/stocker/stocker/backend"
	"github.com/buth/stocker/stocker/crypto"
	"github.com/buth/stocker/stocker/resource"
)

type stocker struct {
	group   string
	state   map[string]resource.Resource
	backend backend.Backend
	crypter crypto.Crypter
}

func New(group string, b backend.Backend, c crypto.Crypter) (*stocker, error) {
	s := &stocker{
		group:   group,
		backend: b,
		crypter: c,
		state:   make(map[string]resource.Resource),
	}

	names, err := s.backend.List(backend.Key("conf", s.group, "resources"))
	if err != nil {
		return s, err
	}

	for _, name := range names {
		if err := s.setResource(name); err != nil {
			return s, err
		}
	}

	return s, nil
}

func (s *stocker) setResource(name string) error {

	if cmd, err := s.backend.Get(backend.Key("conf", s.group, "resource", name)); err != nil {
		return err
	} else {
		if r, ok := s.state[name]; ok {
			r.SetCommand(cmd)
		} else {
			if r, err := resource.New(name, cmd); err != nil {
				return err
			} else {
				s.state[name] = r
			}
		}
	}

	return nil
}

func (s *stocker) reloadResource(name string) error {

	if r, ok := s.state[name]; ok {
		return r.Reload()
	}

	return nil
}

func (s *stocker) handleMessage(channel, message string) error {

	components := backend.DecomposeKey(channel)

	// Every channel should have at least 3 components.
	if len(components) < 3 {
		return errors.New("Recieved a short broadcast.")
	}

	switch message {

	case "remove":
		// return g.remove(components[2])

	case "reload":

		// Update the resource in the state.
		if err := s.setResource(components[2]); err != nil {
			return err
		}

		// Check if we know about this resource. If it exists, we'll reload it,
		// but otherwise this is an error.
		if err := s.reloadResource(components[2]); err != nil {
			return err
		}

	default:
		return errors.New("Recieved a broadcast with an unknown command.")
	}

	return nil
}

func (s *stocker) Run() error {
	return s.backend.Subscribe(backend.Key("cast", s.group, "*"), s.handleMessage)
}
