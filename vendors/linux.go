package vendors

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"time"

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
	Timeout    int
}

func (s *SSH) Run(options SSHOptions) ([]byte, error) {
	key, err := ssh.ParsePrivateKey(options.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
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
	defer client.Close()
	if err != nil {
		return nil, err
	}

	session, err := client.NewSession()
	defer session.Close()
	if err != nil {
		return nil, err
	}

	var b bytes.Buffer
	session.Stdout = &b

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(options.Timeout)*time.Second)
	defer cancel()

	done := make(chan error, 1)
	go func() {

		done <- session.Run(options.Command)
	}()

	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("SSH command timed out after %d seconds", options.Timeout)
	case err := <-done:
		if err != nil {
			return nil, fmt.Errorf("SSH command failed: %w", err)
		}
	}

	return b.Bytes(), err
}

func NewSSH(options SSHOptions) *SSH {

	ssh := &SSH{
		options: options,
	}
	return ssh
}
