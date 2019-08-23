package connection

import (
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"golang.org/x/crypto/ssh/knownhosts"
)

var Handles map[string]*Handle = map[string]*Handle{}

type Handle struct {
	*Connection
	ssh *ssh.Client
}

func NewHandle(c *Connection) *Handle {
	return &Handle{
		Connection: c,
	}
}

func sshAgent() ssh.AuthMethod {
	if sshAgent, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK")); err == nil {
		return ssh.PublicKeysCallback(agent.NewClient(sshAgent).Signers)
	}
	return nil
}

func (h *Handle) sshConnect() error {
	if h.ssh != nil {
		return nil
	}

	hostKeyCallback, err := knownhosts.New(os.Getenv("HOME") + "/.ssh/known_hosts")
	if err != nil && err != os.ErrNotExist {
		return err
	}

	config := &ssh.ClientConfig{
		User: h.SSH.User,
		Auth: []ssh.AuthMethod{
			sshAgent(),
		},
		HostKeyCallback: hostKeyCallback,
	}

	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", h.SSH.Host, h.SSH.Port), config)
	if err != nil {
		return err
	}

	h.ssh = client
	return nil
}

func (h *Handle) Connect() (*Session, error) {
	err := h.sshConnect()
	if err != nil {
		return nil, err
	}

	sess, err := h.ssh.NewSession()
	if err != nil {
		return nil, fmt.Errorf("Failed to create session: %s", err)
	}

	inr, inw := io.Pipe()
	outr, outw := io.Pipe()

	sess.Stderr = ioutil.Discard
	sess.Stdin = inr
	sess.Stdout = outw

	cmd := "bash"
	if h.SSH.Sudo {
		cmd = "sudo " + cmd
	}

	end := make(chan error, 1)
	go func() {
		end <- sess.Run(cmd)
	}()

	return &Session{
		ssh:    sess,
		Stdin:  inw,
		Stdout: outr,
		End:    end,
	}, nil
}
