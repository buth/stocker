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
	Get(group, variable, string) (string, error)
	Set(group, variable, value string) error
	RemoveVariable(group, variable string) error
	RemoveGroup(group, variable string) error
}

func key(components ...string) string {
	return strings.Join(components, KeySep)
}

func KeyEnv(prefix, variable string) string {
	return key(KeyNamePrefix, prefix, KeyNameEnv, variable)
}
