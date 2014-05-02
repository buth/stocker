package backend

import (
	"fmt"
	"github.com/buth/stocker/backend/redis"
	"strings"
)

const (
	KeyNamePrefix = "stocker"
	KeyNameEnv    = "env"
	KeySep        = "/"
)

type Backend interface {
	GetVariable(group, variable string) (string, error)
	SetVariable(group, variable, value string) error
	RemoveVariable(group, variable string) error
	RemoveGroup(group string) error
}

func NewBackend(kind, protocol, host string) (Backend, error) {

	// Select a backend based on kind.
	switch kind {
	case "redis":
		backend := redis.New(protocol, host)
		return backend, nil
	}

	// Assuming no backend is implemented for kind.
	return nil, NoBackendError{kind}
}

func key(components ...string) string {
	return strings.Join(components, KeySep)
}

func KeyEnv(prefix, variable string) string {
	return key(KeyNamePrefix, prefix, KeyNameEnv, variable)
}

type NoBackendError struct {
	Kind string
}

func (e NoBackendError) Error() string {
	return fmt.Sprintf("backend: Backend \"%s\" has not been implemented", e.Kind)
}
