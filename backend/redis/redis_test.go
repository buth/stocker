package redis

import (
	"crypto/rand"
	"encoding/base64"
	"io"
	"testing"
)

func TestGetSet(t *testing.T) {

	r := New("test", "tcp", "127.0.0.1:6379")

	valueBytes := make([]byte, 512)
	if _, err := io.ReadFull(rand.Reader, valueBytes); err != nil {
		t.Errorf("%s", err)
	}

	valueString := base64.StdEncoding.EncodeToString(valueBytes)

	err := r.SetVariable("group", "variable", valueString)
	if err != nil {
		t.Errorf("%s", err)
	}

	v, err := r.GetVariable("group", "variable")
	if err != nil {
		t.Errorf("%s", err)
	}

	if v != valueString {
		t.Errorf("\n%s\n%s\nRetrieved text did not match!", valueString, v)
	}
}
