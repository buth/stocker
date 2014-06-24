package auth

import (
	"bytes"
	"code.google.com/p/go.crypto/ssh"
	"container/list"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/buth/stocker/backend"
	"github.com/buth/stocker/crypto"
	"io"
	"log"
	"net"
	"strings"
	"sync"
)

const (
	WriterUser = `w`
	ReaderUser = `r`
)

type Server interface {
	AddReadKey(key ssh.PublicKey)
	AddWriteKey(key ssh.PublicKey)
	ListenAndServe(address string) error
	Stop() error
}

type server struct {
	backend backend.Backend
	crypter crypto.Crypter

	// SSH
	serverConfig *ssh.ServerConfig
	listeners    *list.List
	listenersMu  sync.Mutex

	// Keys.
	writeKeys, readKeys     *list.List
	writeKeysMu, readKeysMu sync.RWMutex
}

func NewServer(b backend.Backend, c crypto.Crypter, hostKey ssh.Signer) *server {

	// Initialize a new server object with the backend and crypter.
	s := &server{
		backend: b,
		crypter: c,
	}

	// Initialize the listener list.
	s.listeners = list.New()

	// Initialize the key lists.
	s.writeKeys = list.New()
	s.readKeys = list.New()

	// Build a new certificate checker.
	certChecker := &ssh.CertChecker{

		IsAuthority:     NotAnAuthority,
		UserKeyFallback: s.checkUserKey,
	}

	// An SSH server is represented by a ServerConfig, which holds certificate
	// details and handles authentication of ServerConns.
	s.serverConfig = &ssh.ServerConfig{
		Config: ssh.Config{
			Ciphers: []string{"aes256-ctr"},
			MACs:    []string{"hmac-sha1"},
		},
		PublicKeyCallback: certChecker.Authenticate,
	}

	// Add the signing private key.
	s.serverConfig.AddHostKey(hostKey)

	return s
}

// AddWriteKey adds a public key that is authorized to connect to the server
// and perform both read and write operations. If the has already been added
// to the server, this function will update its status.
func (s *server) AddWriteKey(key ssh.PublicKey) {

	// Get the write keys lock for writing.
	s.writeKeysMu.Lock()
	defer s.writeKeysMu.Unlock()

	// Serialze the key and convert it to an string (which is immutable).
	serialized := string(key.Marshal())

	// Add the key string to the write keys list.
	s.writeKeys.PushBack(serialized)
}

func matchKey(key ssh.PublicKey, keys *list.List) bool {

	// Serialze the key and convert it to an string.
	serialized := SerializeKey(key)

	// Add the key string to the write keys list.
	for e := keys.Front(); e != nil; e = e.Next() {
		if serialized == e.Value.(string) {
			return true
		}
	}

	return false
}

func (s *server) matchWriteKey(key ssh.PublicKey) bool {

	// Get the write keys lock for reading.
	s.writeKeysMu.RLock()
	defer s.writeKeysMu.RUnlock()

	// Return the result of the generic match key function.
	return matchKey(key, s.writeKeys)
}

// AddReaderKey adds a public key that is authorized to connect to the server
// and perform only read operations. If the has already been added to the
// server, this function will update its status.
// to the server, this function will update its status.
func (s *server) AddReadKey(key ssh.PublicKey) {

	// Get the read keys lock for writing.
	s.readKeysMu.Lock()
	defer s.readKeysMu.Unlock()

	// Serialze the key and convert it to an string (which is immutable).
	serialized := SerializeKey(key)

	// Add the key string to the read keys list.
	s.readKeys.PushBack(serialized)
}

func (s *server) matchReadKey(key ssh.PublicKey) bool {

	// Get the read keys lock for reading.
	s.readKeysMu.RLock()
	defer s.readKeysMu.RUnlock()

	// Return the result of the generic match key function.
	return matchKey(key, s.readKeys)
}

// checkUserKey determines whether or not the given public key is present for
// the user indicated in the SSH connection meta-data.
func (s *server) checkUserKey(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {

	if u := conn.User(); (u == ReaderUser && s.matchReadKey(key)) || (u == WriterUser && s.matchWriteKey(key)) {
		return &ssh.Permissions{}, nil
	}

	// The default case is to return an error.
	return nil, errors.New("unauthorized")
}

func (s *server) exec(stdout io.Writer, canWrite bool, environment map[string]string, commandString string) error {

	// Try to pull the group from the environment.
	var group string
	if environmentGroup, ok := environment["GROUP"]; ok {
		group = environmentGroup
	}

	// The first (and only) payload value should be a string.
	components := strings.SplitN(commandString, ` `, 2)
	command := components[0]
	argument := ""
	if len(components) == 2 {
		argument = components[1]
	}

	switch command {
	case "env":

		// Pull the encrypted values from the store.
		variables, err := s.backend.GetGroup(group)
		if err != nil {
			return err
		}

		for variable, cryptedValue := range variables {

			// Attempt to decrypt the encrypted value.
			value, err := s.crypter.DecryptString(cryptedValue)
			if err != nil {
				return err
			}

			// Write the variable to the channel.
			fmt.Fprintf(stdout, "%s=%s\n", variable, value)
		}

	case "export":

		// Check for write permission.
		if !canWrite {
			return ServerError{"unauthorized"}
		}

		// Parse the variable name and value from the argument.
		argumentComponents := strings.SplitN(argument, `=`, 2)
		variable := argumentComponents[0]
		value := ""

		// We may need to check the environment for the value.
		if len(argumentComponents) == 2 {
			value = argumentComponents[1]
		} else if environmentValue, ok := environment[variable]; ok {
			value = environmentValue
		}

		// Attempt to encrypt the value.
		cryptedValue, err := s.crypter.EncryptString(value)
		if err != nil {
			return err
		}

		// Save the encrypted value in the store.
		if err := s.backend.SetVariable(group, variable, cryptedValue); err != nil {
			return err
		}

	case "unset":

		// Check for write permission.
		if !canWrite {
			return ServerError{"unauthorized"}
		}

		// Assume the argument is a variable name and remove it.
		if err := s.backend.RemoveVariable(group, argument); err != nil {
			return err
		}
	}

	return nil
}

func (s *server) handleRequests(channel ssh.Channel, canWrite bool, in <-chan *ssh.Request) {

	// Close the connection when we return.
	defer channel.Close()

	// Maintain a group state for this channel.
	environment := make(map[string]string)

	// Pull requests off the incoming channel.
	for request := range in {

		// Assume that this request is invalid.
		ok := false

		// Switch on the request type.
		switch request.Type {
		case "env":

			// Try to parse the payload. If we can't then there's nothing we
			// can do with this request.
			if payload, err := UnpackMessage(request.Payload); err != nil {

				// Write the error message to the log.
				log.Println(err)
			} else {

				// Write the payload slice into the environment map.
				for i := 0; i < len(payload)/2; i++ {
					environment[payload[i*2]] = payload[i*2+1]
				}

				// Indicate success.
				ok = true
			}

		case "exec":

			// Try to parse the payload. If we can't we don't want to
			// proceed.
			payload, err := UnpackMessage(request.Payload)
			if err != nil {

				// Notify the caller that we couldn't run the command.
				request.Reply(false, nil)
				return
			}

			// Indicate that we have started running the command.
			request.Reply(true, nil)

			// The exit status will be reported as a 4-byte, little-endian integer.
			exitStatusBuffer := bytes.NewBuffer([]byte{})

			// Run the command, reporting any error as a failure.
			if err := s.exec(channel, canWrite, environment, payload[0]); err != nil {

				// Write the error message to the log.
				log.Println(err)
				binary.Write(exitStatusBuffer, binary.BigEndian, uint32(1))
			} else {
				binary.Write(exitStatusBuffer, binary.BigEndian, uint32(0))
			}

			// Write the exit status.
			channel.SendRequest("exit-status", false, exitStatusBuffer.Bytes())

			// Only one exec command can be handled per channel, so we're done.
			return
		}

		// If requested, reply with the status.
		if request.WantReply {
			request.Reply(ok, nil)
		}
	}
}

func (s *server) handleChannels(canWrite bool, in <-chan ssh.NewChannel) {

	// Pull channels off the incoming channel.
	for newChannel := range in {

		// Only accept sessions.
		if newChannel.ChannelType() != "session" {
			newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")

			continue
		}

		// Attempt to accecpt the session channel.
		channel, requests, err := newChannel.Accept()
		if err != nil {
			log.Printf("server: could not accept channel: %s\n", err.Error())
			continue
		}

		go s.handleRequests(channel, canWrite, requests)
	}
}

// ListenAndServe starts a new SSH server listening on the given address.
func (s *server) ListenAndServe(address string) error {

	// Get the listeners lock.
	s.listenersMu.Lock()

	// Start listening.
	listener, err := net.Listen("tcp", address)
	if err != nil {
		s.listenersMu.Unlock()
		return err
	}

	// Add the listener to the list and release the lock.
	s.listeners.PushBack(listener)
	s.listenersMu.Unlock()

	for {

		// Accept a new connection.
		nConn, err := listener.Accept()
		if err != nil {
			return err
		}

		sConn, chans, reqs, err := ssh.NewServerConn(nConn, s.serverConfig)
		if err != nil {
			log.Println("server: failed to handshake")
			continue
		}

		// Determine whether or not the user can write.
		canWrite := false
		if sConn.User() == WriterUser {
			canWrite = true
		}

		// The incoming Request channel must be serviced.
		go ssh.DiscardRequests(reqs)

		// Service the incoming Channel channel.
		go s.handleChannels(canWrite, chans)
	}

	return nil
}

func (s *server) Stop() error {

	// Get the listeners lock and defer its closing.
	s.listenersMu.Lock()
	defer s.listenersMu.Unlock()

	for e := s.listeners.Front(); e != nil; e = e.Next() {

		if err := e.Value.(net.Listener).Close(); err != nil {
			return err
		}

		s.listeners.Remove(e)
	}

	return nil
}

type ServerError struct {
	Err string
}

func (e ServerError) Error() string {
	return fmt.Sprintf("server: %s", e.Err)
}
