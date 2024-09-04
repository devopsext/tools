package vendors

import (
	"bytes"
	"fmt"
	"net"

	"golang.org/x/crypto/ssh"
)

type SSH struct {
	options SSHOptions
}
type SSHOptions struct {
	User       string
	Address    string
	PrivateKey []byte
	Command    string
}

func (s *SSH) Run(options SSHOptions) (string, error) {
	key, err := ssh.ParsePrivateKey(options.PrivateKey)
	if err != nil {
		return "", fmt.Errorf("failed to parse private key: %w", err)
	}

	config := &ssh.ClientConfig{
		User:            options.User,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(key),
		},
		/*
			Auth: []ssh.AuthMethod{
				ssh.Password("PASSWORD"),
			},
		*/
	}

	client, err := ssh.Dial("tcp", net.JoinHostPort(options.Address, "22"), config)
	if err != nil {
		return "", err
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close() // Ensure session is closed

	var b bytes.Buffer
	session.Stdout = &b

	err = session.Run(options.Command)
	return b.String(), err
}

func NewSSH(options SSHOptions) *SSH {

	ssh := &SSH{
		options: options,
	}
	return ssh
}
