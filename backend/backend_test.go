package backend

import (
	"testing"
)

var testBackends = []struct {
	Kind, Namespace, Protocol, Address string
}{
	{"redis", "redistest", "tcp", ":6379"},
}

var testBackendsPairs = map[string]string{
	"TESTVARIABLE1": "TESTVALUE1",
	"TESTVARIABLE2": "TESTVALUE3",
	"TESTVARIABLE3": "TESTVALUE3",
}

func TestBackend(t *testing.T) {
	for _, b := range testBackends {

		backend, err := NewBackend(b.Kind, b.Namespace, b.Protocol, b.Address)
		if err != nil {
			t.Fatal(err)
		}

		for variable, value := range testBackendsPairs {
			if err := backend.SetVariable("testgroup", variable, value); err != nil {
				t.Error(err)
			}
		}

		for variable, value := range testBackendsPairs {
			v, err := backend.GetVariable("testgroup", variable)
			if err != nil {
				t.Error(err)
			} else if v != value {
				t.Errorf("expected value %s for %s but found %s!", value, variable, v)
			}
		}

		if err := backend.RemoveGroup("testgroup"); err != nil {
			t.Fatal(err)
		}
	}
}

func TestBackendGetAll(t *testing.T) {
	for _, b := range testBackends {

		backend, err := NewBackend(b.Kind, b.Namespace, b.Protocol, b.Address)
		if err != nil {
			t.Fatal(err)
		}

		for variable, value := range testBackendsPairs {
			if err := backend.SetVariable("testgroup", variable, value); err != nil {
				t.Error(err)
			}
		}

		variables, err := backend.GetGroup("testgroup")
		if err != nil {
			t.Error(err)
		}

		for variable, value := range testBackendsPairs {
			v, ok := variables[variable]
			if !ok {
				t.Errorf("no value returned for %s!", variable)
			} else if v != value {
				t.Errorf("expected value %s for %s but found %s!", value, variable, v)
			}
		}

		if err := backend.RemoveGroup("testgroup"); err != nil {
			t.Fatal(err)
		}
	}
}
