package auth

import (
	"bytes"
	"code.google.com/p/go.crypto/ssh"
	"container/list"
	"encoding/binary"
	"fmt"
	"github.com/buth/stocker/backend"
	"github.com/buth/stocker/crypto"
	"io"
	"log"
	"net"
	"strings"
)

func UnpackMessage(message []byte) ([]string, error) {

	// Create a new buffer based on the message byte slice and initialize an
	// empty list.
	buf := bytes.NewBuffer(message)
	l := list.New()

	// We need 4 bytes to define a number of bytes to read. If there are only
	// 4 bytes left we can assume that the number is zero an move on.
	for buf.Len() > 4 {

		// Create a 4-byte unsigned integer to contain the read value from the
		// message and read from the buffer.
		var n uint32
		if err := binary.Read(buf, binary.BigEndian, &n); err != nil {
			return nil, err
		}

		// Convert the unsigned 4-byte integer to an int and read out the
		// specified number of bytes. Store the resulting byte slice in the
		// list.
		l.PushBack(buf.Next(int(n)))
	}

	rval := make([]string, l.Len())
	i := 0
	for e := l.Front(); e != nil; e = e.Next() {
		rval[i] = string(e.Value.([]byte))
		i++
	}

	return rval, nil
}

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

func (s *Server) exec(stdout io.Writer, environment map[string]string, commandString string) error {

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
		variables, err := s.Backend.GetGroup(group)
		if err != nil {
			return err
		}

		for variable, cryptedValue := range variables {

			// Attempt to decrypt the encrypted value.
			value, err := s.Crypter.DecryptString(cryptedValue)
			if err != nil {
				return err
			}

			// Write the variable to the channel.
			fmt.Fprintf(stdout, "%s=%s\n", variable, value)
		}

	case "export":

		// Parse the variable name and value from the argument.
		argumentComponents := strings.SplitN(argument, `=`, 2)
		variable := argumentComponents[0]
		value := argumentComponents[1]

		// Attempt to encrypt the value.
		cryptedValue, err := s.Crypter.EncryptString(value)
		if err != nil {
			return err
		}

		// Save the encrypted value in the store.
		if err := s.Backend.SetVariable(group, variable, cryptedValue); err != nil {
			return err
		}

	case "unset":

		// Assume the argument is a variable name and remove it.
		if err := s.Backend.RemoveVariable(group, argument); err != nil {
			return err
		}
	}

	return nil
}

func (s *Server) handleRequests(channel ssh.Channel, in <-chan *ssh.Request) {

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
			if err := s.exec(channel, environment, payload[0]); err != nil {

				// Write the error message to the log.
				log.Println(err)
				binary.Write(exitStatusBuffer, binary.LittleEndian, uint32(1))
			} else {
				binary.Write(exitStatusBuffer, binary.LittleEndian, uint32(0))
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

			go s.handleRequests(channel, requests)
		}
	}

	return nil
}
