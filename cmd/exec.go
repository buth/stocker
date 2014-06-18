package cmd

import (
	"bytes"
	"code.google.com/p/go.crypto/ssh"
	"code.google.com/p/go.crypto/ssh/agent"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"strings"
	"syscall"
)

var Exec = &Command{
	UsageLine: "exec [options] command [argument...]",
	Short:     "execute a command with the given environment",
}

type StringAcumulator []string

func (s *StringAcumulator) Set(value string) error {
	*s = append(*s, value)
	return nil
}

func (s *StringAcumulator) String() string {
	return fmt.Sprintf("%s", *s)
}

var execConfig struct {
	Address, PrivateFilepath, Group, User string
}

func init() {
	Exec.Run = execRun
	Exec.Flag.StringVar(&execConfig.Address, "a", ":2022", "address of the stocker server")
	Exec.Flag.StringVar(&execConfig.Group, "g", "", "group to use for storing and retrieving data")
	Exec.Flag.StringVar(&execConfig.PrivateFilepath, "i", "", "path to an SSH private key")
	Exec.Flag.StringVar(&execConfig.User, "u", "", "user to execute the command as")
}

func execRun(cmd *Command, args []string) {

	// Find the expanded path to cmd.
	command, err := exec.LookPath(args[0])
	if err != nil {
		log.Fatalf("%s: command not found", args[0])
	}

	// Create an empty SSH configuration as we don't yet know what
	// authentication methods to use.
	config := &ssh.ClientConfig{}

	// Check if we should use an explicitly defined key on disk or consult
	// ssh-agent.
	if execConfig.PrivateFilepath != "" {

		privateBytes, err := ioutil.ReadFile(execConfig.PrivateFilepath)
		if err != nil {
			log.Fatal("Failed to read private key: " + err.Error())
		}

		private, err := ssh.ParsePrivateKey(privateBytes)
		if err != nil {
			log.Fatal("Failed to parse private key: " + err.Error())
		}

		config.Auth = []ssh.AuthMethod{
			ssh.PublicKeys(private),
		}
	} else {

		sshAuthSock := os.Getenv(`SSH_AUTH_SOCK`)

		socket, err := net.Dial("unix", sshAuthSock)
		if err != nil {
			log.Fatal("Failed to to open connection to ssh-agent: " + err.Error())
		}

		sshAgent := agent.NewClient(socket)
		signers, err := sshAgent.Signers()
		if err != nil {
			log.Fatal("Failed to retrieve signers from ssh-agent: " + err.Error())
		}

		config.Auth = []ssh.AuthMethod{
			ssh.PublicKeys(signers...),
		}
	}

	client, err := ssh.Dial("tcp", execConfig.Address, config)
	if err != nil {
		log.Fatal("Failed to dial: " + err.Error())
	}

	// Each ClientConn can support multiple interactive sessions,
	// represented by a Session.
	session, err := client.NewSession()
	if err != nil {
		log.Fatal("Failed to create session: " + err.Error())
	}
	defer session.Close()

	// Once a Session is created, you can execute a single command on
	// the remote side using the Run method.
	var b bytes.Buffer
	session.Stdout = &b
	if err := session.Run("env"); err != nil {
		log.Fatal("Failed to run: " + err.Error())
	}

	// Create a map of environment variables to be passed to cmd and
	// initialize it with the current environment.
	env := make(map[string]string)
	for _, variable := range os.Environ() {
		components := strings.Split(variable, "=")
		env[components[0]] = components[1]
	}

	// Parse the stocker environment and save it into the env.
	pairs := strings.Split(b.String(), "\n")
	for _, pair := range pairs {
		components := strings.SplitN(pair, "=", 2)
		if len(components) == 2 {
			env[components[0]] = components[1]
		}
	}

	// Create a list of environment key/value pairs and write the
	// flattened environment variables map to it.
	commandEnv := make([]string, 0, len(env))
	for key, value := range env {
		commandEnv = commandEnv[:len(commandEnv)+1]
		commandEnv[len(commandEnv)-1] = fmt.Sprintf("%s=%s", key, value)
	}

	// Handle user.
	if execConfig.User != "" {

		u, err := user.Lookup(execConfig.User)
		if err != nil {
			log.Fatal(err)
		}

		uid, err := strconv.Atoi(u.Uid)
		if err != nil {
			log.Fatal(err)
		}

		if err := syscall.Setuid(uid); err != nil {
			log.Fatal(err)
		}
	}

	// Exec the new command.
	syscall.Exec(command, args, commandEnv)
}
