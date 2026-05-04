package integration

import "golang.org/x/crypto/ssh"

const SSHKey = "framework.ssh"

type SSHService interface {
	Client(hostName string) (SSHClient, error)
}

type SSHClient interface {
	NewSession() (SSHSession, error)
	Close() error
}

type SSHSession interface {
	CombinedOutput(cmd string) ([]byte, error)
	Run(cmd string) error
	Close() error
}

type NativeSSHClient interface {
	NativeSSHClient() *ssh.Client
}
