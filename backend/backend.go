package backend

import (
	"fmt"
	"github.com/buth/stocker/backend/redis"
)

type Backend interface {
	GetVariable(group, variable string) (string, error)
	SetVariable(group, variable, value string) error
	RemoveVariable(group, variable string) error
	GetGroup(group string) (map[string]string, error)
	RemoveGroup(group string) error
}

func NewBackend(kind, namespace, protocol, address string) (Backend, error) {

	// Select a backend based on kind.
	switch kind {
	case "redis":
		backend := redis.New(namespace, protocol, address)
		return backend, nil
	}

	// Assuming no backend is implemented for kind.
	return nil, NoBackendError{kind}
}

type NoBackendError struct {
	Kind string
}

func (e NoBackendError) Error() string {
	return fmt.Sprintf("backend: Backend \"%s\" has not been implemented", e.Kind)
}
