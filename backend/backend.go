package backend

import (
	"strings"
)

type Backend interface {
	Get(key string) (string, error)
	Set(key, value string) error
	Remove(string) error
	Subscribe(key string, process func(value string)) error
}

func Key(components ...string) string {
	return strings.Join(components, "/")
}

func DecomposeKey(key string) []string {
	return strings.Split(key, "/")
}
