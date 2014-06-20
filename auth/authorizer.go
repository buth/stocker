package auth

import (
	"code.google.com/p/go.crypto/ssh"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"
)

// Authorizer maintains a thread-safe map of serialized public keys to check
// against incomming connections. It will differentiate between keys that
// allow write access and those that do not.
type Authorizer struct {
	initOnce sync.Once
	keys     map[string]bool
	mu       sync.RWMutex
	client   *http.Client
}

// init sets values for the Authorizer's internal key map and HTTP client.
func (a *Authorizer) init() {
	a.keys = make(map[string]bool)
	a.client = &http.Client{
		Transport: &http.Transport{
			MaxIdleConnsPerHost:   1,
			ResponseHeaderTimeout: time.Minute,
		},
	}
}

// Init sets values for the Authorizer's internal key map and HTTP client.
func (a *Authorizer) Init() {
	a.initOnce.Do(a.init)
}

// NewAuthorizer is a convenience method fror creating a new Authorizer
// object.
func NewAuthorizer() *Authorizer {
	return &Authorizer{}
}

// AddKey adds a key to the Authorizer's internal key map (if it is missing)
// and sets a value indicating whether or not the key's user should be allowed
// to write; successive writes can update this value.
func (a *Authorizer) AddKey(key ssh.PublicKey, canWrite bool) {
	a.Init()

	// Get the lock for writing.
	a.mu.Lock()
	defer a.mu.Unlock()

	serialized := string(key.Marshal())
	a.keys[serialized] = canWrite
}

// AddKeysFromURL fetches a JSON array of public keys (conforming to GitHub's
// API) from a URL and adds them to the Authorizer's internal key map.
func (a *Authorizer) AddKeysFromURL(url string, canWrite bool) error {
	a.Init()

	// Fetch the public keys.
	response, err := a.client.Get(url)
	if err != nil {
		return err
	}

	// Read out the entire body.
	jsonResponse, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	// Build a raw keys object that reflects the expected structure of the JSON.
	var rawKeys []struct {
		Key string
	}

	// Try to parse the body of the response as JSON.
	if err := json.Unmarshal(jsonResponse, &rawKeys); err != nil {
		return err
	}

	// Build a new authorizer and iterate through the raw keys, parsing them
	// and then adding them.
	for _, rawKey := range rawKeys {

		// We're only interested in the key itself and whether or not there was an error.
		publicKey, _, _, _, err := ssh.ParseAuthorizedKey([]byte(rawKey.Key))
		if err != nil {
			return err
		}

		// Add the key to the authorizer.
		a.AddKey(publicKey, canWrite)
	}

	return nil
}

//
func (a *Authorizer) Authorize(key ssh.PublicKey) (bool, bool) {
	a.Init()

	// Get the lock for reading.
	a.mu.RLock()
	defer a.mu.RUnlock()

	// Serialize the key and attempt to pull it from the map.
	serialized := string(key.Marshal())
	if canWrite, ok := a.keys[serialized]; ok {
		return true, canWrite
	}

	return false, false
}

// UserKeyFallback evaluates a public key in the context of an ssh connection
// to determine whether or not the key is authorized and what permissions the
// user of the key may be entitled to.
func (a *Authorizer) UserKeyFallback(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {

	if authorized, canWrite := a.Authorize(key); authorized {
		permissions := &ssh.Permissions{}
		permissions.Extensions = make(map[string]string)

		if canWrite {
			permissions.Extensions["permit-stocker-writes"] = "Yes"
		} else {
			permissions.Extensions["permit-stocker-writes"] = "No"
		}

		return permissions, nil
	}

	return nil, UnauthorizedPublicKeyError{key}
}

type UnauthorizedPublicKeyError struct {
	Key ssh.PublicKey
}

func (u UnauthorizedPublicKeyError) Error() string {
	return fmt.Sprintf("public key (%s) is not authorized", u.Key.Type())
}
