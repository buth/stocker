package resource

import (
	"log"
	// "net/url"
	"testing"
)

func TestNew(t *testing.T) {

	c, err := New("test", "-e PASSWORD=secret -p 5000 stackbrew/registry")
	log.Println(c, err)
	c.Reload()

}
