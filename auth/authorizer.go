package auth

import (
	"bytes"
	"code.google.com/p/go.crypto/ssh"
	"container/list"
	"sync"
)

// Authorizer maintains a list of serialized public keys to check against
// incomming connections.
type Authorizer struct {
	initOnce sync.Once
	keys     *list.List
	mu       sync.RWMutex
}

func (a *Authorizer) init() {
	a.keys = list.New()
}

func (a *Authorizer) Init() {
	a.initOnce.Do(a.init)
}

func NewAuthorizer() *Authorizer {
	return &Authorizer{}
}

func (a *Authorizer) AddKey(key ssh.PublicKey) {
	a.Init()

	// Get the lock for writing.
	a.mu.Lock()
	defer a.mu.Unlock()

	serialized := key.Marshal()
	a.keys.PushBack(serialized)
}

func (a *Authorizer) Authorize(key ssh.PublicKey) bool {
	a.Init()

	// Get the lock for reading.
	a.mu.RLock()
	defer a.mu.RUnlock()

	serialized := key.Marshal()
	for e := a.keys.Front(); e != nil; e = e.Next() {
		if bytes.Equal(e.Value.([]byte), serialized) {
			return true
		}
	}
	return false
}
