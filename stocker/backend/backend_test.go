package backend

import (
	"crypto/rand"
	"encoding/base64"
	"github.com/garyburd/redigo/redis"
	"io"
	"testing"
	"time"
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

func TestGetSetTTL(t *testing.T) {

	r := New("tcp", "127.0.0.1:6379")

	valueBytes := make([]byte, 512)
	if _, err := io.ReadFull(rand.Reader, valueBytes); err != nil {
		t.Errorf("%s", err)
	}

	valueString := base64.StdEncoding.EncodeToString(valueBytes)

	err := r.SetWithTTL("test", valueString, 1)
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

	time.Sleep(2 * time.Second)

	_, err = r.Get("test")
	if err != redis.ErrNil {
		t.Errorf("\n%s\nExpected a nil value error!", err)
	}
}
