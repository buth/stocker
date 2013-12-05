package stocker

import (
	"log"
	// "net/url"
	"github.com/buth/stocker/stocker/backend/redis"
	"github.com/buth/stocker/stocker/crypto/chain"
	"testing"
)

func TestNew(t *testing.T) {

	r := redis.New("tcp", "127.0.0.1:6379")

	a, err := chain.New(chain.GenerateKey())

	c, err := New("testgroup", r, a)
	log.Println(c, err)

}
