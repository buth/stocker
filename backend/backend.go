package backend

import (
	"strings"
)

const (
	KeyNamePrefix = "stocker"
	KeyNameEnv    = "env"
	KeySep        = "/"
)

type Backend interface {
	Get(string) (string, error)
	Set(key, value string) error
	Remove(string) error
	Subscribe(key string, process func(value string)) error
	Publish(key, message string) error
}

func key(components ...string) string {
	return strings.Join(components, KeySep)
}

func KeyEnv(prefix, variable string) string {
	return key(KeyNamePrefix, prefix, KeyNameEnv, variable)
}
