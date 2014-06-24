package auth

import (
	"bytes"
	"code.google.com/p/go.crypto/ssh"
	"code.google.com/p/go.crypto/ssh/agent"
	"net"
	"os"
)

type Client interface {
	Run(command string) (string, error)
	Close() error
}

type client struct {
	client *ssh.Client
}

func NewClient(user, address string, privateKey []byte) (*client, error) {
	c := &client{}

	config := &ssh.ClientConfig{
		User: user,
	}

	// Check if we've been given a byte slice from which to parse a key.
	if privateKey != nil {

		privateKeyParsed, err := ssh.ParsePrivateKey(privateKey)
		if err != nil {
			return c, err
		}

		config.Auth = []ssh.AuthMethod{
			ssh.PublicKeys(privateKeyParsed),
		}
	} else {

		sshAuthSock := os.Getenv(`SSH_AUTH_SOCK`)

		socket, err := net.Dial("unix", sshAuthSock)
		if err != nil {
			return c, err
		}

		sshAgent := agent.NewClient(socket)
		signers, err := sshAgent.Signers()
		if err != nil {
			return c, err
		}

		config.Auth = []ssh.AuthMethod{
			ssh.PublicKeys(signers...),
		}
	}

	client, err := ssh.Dial("tcp", address, config)
	if err != nil {
		return c, err
	}

	c.client = client

	return c, nil
}

func (c *client) Run(command string, env map[string]string) (string, error) {

	// Create a new session in which to run the command.
	session, err := c.client.NewSession()
	if err != nil {
		return "", err
	}

	// Defer the sessions closing, ignoring any error.
	defer session.Close()

	// Once a Session is created, you can execute a single command on
	// the remote side using the Run method.
	var buf bytes.Buffer
	session.Stdout = &buf

	// Set the environment.
	for variable, value := range env {
		if err := session.Setenv(variable, value); err != nil {
			return "", err
		}
	}

	if err := session.Run(command); err != nil {
		return "", err
	}

	// Return the buffer as a string.
	return buf.String(), nil
}

func (c *client) Close() error {
	return c.client.Close()
}
