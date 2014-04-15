package crypto

import (
	"testing"
)

func TestNewKey(t *testing.T) {
	key := NewKey()

	if !key.Valid() {
		t.Error("new key was not valid!")
	}
}

func TestEmptyKey(t *testing.T) {
	key := Key{}

	if key.Valid() {
		t.Error("empty key was valid!")
	}
}
