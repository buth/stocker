package auth

import (
	"code.google.com/p/go.crypto/ssh"
	"fmt"
	"github.com/buth/stocker/backend"
	"github.com/buth/stocker/crypto"
	"io"
	"log"
	"net"
	"strings"
)

const (
	ACK = '\x06'
	NAK = '\x15'
	EOT = '\x04'
)

type Server struct {
	Config  *ssh.ServerConfig
	Backend backend.Backend
	Crypter crypto.Crypter
}

func NewServer(s *ssh.ServerConfig, b backend.Backend, c crypto.Crypter) *Server {
	return &Server{
		Config:  s,
		Backend: b,
		Crypter: c,
	}
}

func (s *Server) handleError(channel io.Writer, err error) {
	channel.Write([]byte{NAK})
	log.Println(err)
}

func (s *Server) handleChannel(channel io.ReadWriteCloser) {
	defer channel.Close()

	// Start with an empty group for this channel.
	group := ""

	// Create a new buffer to read into.
	input := make([]byte, 256)

	for {

		// Read a chunk of input. Right now, this is assuming that there is no
		// such thing as partial input.
		n, err := channel.Read(input)
		if err != nil {
			return
		}

		// Parse the command string.
		commandString := string(input[:n])
		commandStringChomp := strings.TrimRight(commandString, "\n")
		commands := strings.Split(commandStringChomp, "\n")

		for _, command := range commands {

			switch {

			// End of transmission.
			case command[0] == EOT:
				return

			// Set group.
			case command[0] == '@':
				group = command[1:]

			// Get variable.
			case command[0] == '$':
				variable := command[1:]

				// Pull the encrypted value from the store.
				cryptedValue, err := s.Backend.GetVariable(group, variable)
				if err != nil {
					s.handleError(channel, err)
					continue
				}

				// Attempt to decrypt it.
				value, err := s.Crypter.DecryptString(cryptedValue)
				if err != nil {
					s.handleError(channel, err)
					continue
				}

				// Write the variable to the channel.
				fmt.Fprintf(channel, "%s=%s\n", variable, value)

			// Delete variable.
			case command[0] == '!':
				variable := command[1:]
				if err := s.Backend.RemoveVariable(group, variable); err != nil {
					s.handleError(channel, err)
					continue
				}

			// Get all variables.
			case command[0] == '*':

				// Pull the encrypted values from the store.
				variables, err := s.Backend.GetGroup(group)
				if err != nil {
					s.handleError(channel, err)
					continue
				}

				for variable, cryptedValue := range variables {

					// Attempt to decrypt the encrypted value.
					value, err := s.Crypter.DecryptString(cryptedValue)
					if err != nil {
						s.handleError(channel, err)
						continue
					}

					// Write the variable to the channel.
					fmt.Fprintf(channel, "%s=%s\n", variable, value)
				}

			// Set variable.
			case strings.ContainsRune(command, '='):
				components := strings.SplitN(command, "=", 2)
				variable := components[0]
				value := components[1]

				// Attempt to encrypt the value.
				cryptedValue, err := s.Crypter.EncryptString(value)
				if err != nil {
					s.handleError(channel, err)
					continue
				}

				// Save the encrypted value in the store.
				if err := s.Backend.SetVariable(group, variable, cryptedValue); err != nil {
					s.handleError(channel, err)
					continue
				}
			}

			channel.Write([]byte{ACK})
		}
	}
}

func (s *Server) ListenAndServe(address string) error {

	// Start listening.
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}

	for {

		// Accept a new connection.
		nConn, err := listener.Accept()
		if err != nil {
			log.Println("server: failed to accept incoming connection")
		}

		_, chans, reqs, err := ssh.NewServerConn(nConn, s.Config)
		if err != nil {
			log.Println("server: failed to handshake")
			continue
		}

		// The incoming Request channel must be serviced.
		go ssh.DiscardRequests(reqs)

		// Service the incoming Channel channel.
		for newChannel := range chans {

			// Only accept sessions.
			if newChannel.ChannelType() != "session" {
				newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
				continue
			}

			// Attempt to accecpt the session channel.
			channel, requests, err := newChannel.Accept()
			if err != nil {
				log.Println("server: could not accept channel.")
				continue
			}

			go ssh.DiscardRequests(requests)
			go s.handleChannel(channel)
		}
	}

	return nil
}
