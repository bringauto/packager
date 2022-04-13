package bringauto_ssh

import (
	"bytes"
	"os"
)

type Command struct {
	Command string
}

// RunCommandOverSSH
// Run  command over SSH
//
func (command *Command) RunCommandOverSSH(credentials SSHCredentials) (string, error) {
	var err error
	stdOut := bytes.Buffer{}
	session := SSHSession{
		StdOut: &stdOut,
		StdErr: os.Stderr,
		StdIn:  os.Stdin,
	}

	err = session.LoginMultipleAttempts(credentials)
	if err != nil {
		return "", err
	}
	defer session.Logout()

	err = session.Start(command.Command)
	if err != nil {
		return "", err
	}
	err = session.Wait()
	if err != nil {
		return "", err
	}
	session.Logout() // buffer sync
	return stdOut.String(), nil
}
