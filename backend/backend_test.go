package backend

import (
	"bytes"
	"testing"
)

var testBackends = []struct {
	Kind, Namespace, Protocol, Address string
}{
	{"redis", "redistest", "tcp", ":6379"},
	{"etcd", "etcdtest", "tcp", ":4001"},
}

func backendPairs() map[string][]byte {
	testBackendsPairs := make(map[string][]byte)
	testBackendsPairs["TESTVARIABLE1"] = []byte("test value #1")
	testBackendsPairs["TESTVARIABLE2"] = []byte("test value #2")
	testBackendsPairs["TESTVARIABLE3"] = []byte("test value #3")
	return testBackendsPairs
}

func TestBackend(t *testing.T) {
	testBackendsPairs := backendPairs()
	for _, b := range testBackends {

		backend, err := NewBackend(b.Kind, b.Namespace, b.Protocol, b.Address)
		if err != nil {
			t.Fatal(err)
		}

		for variable, value := range testBackendsPairs {

			if err := backend.SetVariable("testgroup", variable, value); err != nil {
				t.Fatal(err)
			}
		}

		for variable, value := range testBackendsPairs {
			v, err := backend.GetVariable("testgroup", variable)
			if err != nil {
				t.Error(err)
			} else if !bytes.Equal(v, value) {
				t.Errorf("expected value %s for %s but found %s!", value, variable, v)
			}
		}

		if err := backend.RemoveGroup("testgroup"); err != nil {
			t.Fatal(err)
		}
	}
}

func TestBackendGetAll(t *testing.T) {
	testBackendsPairs := backendPairs()
	for _, b := range testBackends {

		backend, err := NewBackend(b.Kind, b.Namespace, b.Protocol, b.Address)
		if err != nil {
			t.Fatal(err)
		}

		for variable, value := range testBackendsPairs {
			if err := backend.SetVariable("testgroup", variable, value); err != nil {
				t.Fatal(err)
			}
		}

		variables, err := backend.GetGroup("testgroup")
		if err != nil {
			t.Fatal(err)
		}

		for variable, value := range testBackendsPairs {
			v, ok := variables[variable]
			if !ok {
				t.Errorf("no value returned for %s!", variable)
			} else if !bytes.Equal(v, value) {
				t.Errorf("expected value \"%s\" for %s but found \"%s\"!", value, variable, v)
			}
		}

		if err := backend.RemoveGroup("testgroup"); err != nil {
			t.Fatal(err)
		}
	}
}
