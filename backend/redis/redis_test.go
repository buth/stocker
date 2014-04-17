package redis

import (
	"crypto/rand"
	"encoding/base64"
	"io"
	"sync"
	"testing"
)

func TestGetSet(t *testing.T) {

	r := New("tcp", "127.0.0.1:6379")

	valueBytes := make([]byte, 512)
	if _, err := io.ReadFull(rand.Reader, valueBytes); err != nil {
		t.Errorf("%s", err)
	}

	valueString := base64.StdEncoding.EncodeToString(valueBytes)

	err := r.Set("test", valueString)
	if err != nil {
		t.Errorf("%s", err)
	}

	v, err := r.Get("test")
	if err != nil {
		t.Errorf("%s", err)
	}

	if v != valueString {
		t.Errorf("\n%s\n%s\nRetrieved text did not match!", valueString, v)
	}
}

func TestPubSub(t *testing.T) {

	r := New("tcp", "127.0.0.1:6379")

	var wg sync.WaitGroup

	processMessage := func(message string) {
		wg.Done()
	}

	r.Subscribe("test/*", processMessage)

	wg.Add(5)
	r.Publish("test/key1", "something")
	r.Publish("test/key2", "nothing")
	r.Publish("test/key3", "nothing")
	r.Publish("test/key4", "nothing")
	r.Publish("test/", "nothing")
	wg.Wait()
}
