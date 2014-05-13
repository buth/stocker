package backend

import (
	"testing"
)

var testBackends = []struct {
	Kind, Namespace, Protocol, Address string
}{
	{"redis", "redistest", "tcp", ":6379"},
}

func TestBackend(t *testing.T) {
	for _, b := range testBackends {

		backend, err := NewBackend(b.Kind, b.Namespace, b.Protocol, b.Address)
		if err != nil {
			t.Fatal(err)
		}

		if err := backend.SetVariable("testgroup", "TESTVARIABLE", "testvalue"); err != nil {
			t.Fatal(err)
		}

		value, err := backend.GetVariable("testgroup", "TESTVARIABLE")
		if err != nil {
			t.Error(err)
		}

		if value != "testvalue" {
			t.Error("value did not match!")
		}

		if err := backend.RemoveGroup("testgroup"); err != nil {
			t.Fatal(err)
		}
	}
}
