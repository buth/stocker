package redis

import (
	"bytes"
	"crypto/rand"
	"io"
	"testing"
)

func TestGetSet(t *testing.T) {

	r := New("test", "tcp", "127.0.0.1:6379")

	valueBytes := make([]byte, 512)
	if _, err := io.ReadFull(rand.Reader, valueBytes); err != nil {
		t.Fatalf("%s", err)
	}

	err := r.SetVariable("group", "variable", valueString)
	if err != nil {
		t.Fatalf("%s", err)
	}

	v, err := r.GetVariable("group", "variable")
	if err != nil {
		t.Fatalf("%s", err)
	}

	if !bytes.Equal(v, valueBytes) {
		t.Errorf("\n%s\n%s\nRetrieved bytes did not match!", valueString, v)
	}
}
