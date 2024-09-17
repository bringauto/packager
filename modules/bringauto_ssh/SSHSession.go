package bringauto_ssh

import (
	"bringauto/modules/bringauto_log"
	"bringauto/modules/bringauto_prerequisites"
	"fmt"
	"golang.org/x/crypto/ssh"
	"io"
	"net"
	"strconv"
	"time"
)

// SSHCredentials is used as endpoint credentials of the remote server
//
type SSHCredentials struct {
	IPAddress string
	Port      uint16
	Username  string
	Password  string
}

// SSHSession represents standard SSH Session needed for each SSH "Connection"
//
type SSHSession struct {
	StdIn      io.Reader
	StdOut     io.Writer
	StdErr     io.Writer
	sshClient  *ssh.Client
	sshSession *ssh.Session
}

func (cred *SSHCredentials) FillDefault(*bringauto_prerequisites.Args) error {
	*cred = SSHCredentials{
		IPAddress: "127.0.0.1",
		Port:      1122,
		Username:  "root",
		Password:  "1234",
	}
	return nil
}

func (cred *SSHCredentials) FillDynamic(*bringauto_prerequisites.Args) error {
	return nil
}

func (cred *SSHCredentials) CheckPrerequisites(*bringauto_prerequisites.Args) error {
	return nil
}

func (session *SSHSession) GetSSHSession() *ssh.Session {
	return session.sshSession
}

// LoginMultipleAttempts calls Login a maximum of N times.
// It calls Login, if the Login fails it waits M seconds and then try it again.
// If the Login failed N times then error is returned.
//
func (session *SSHSession) LoginMultipleAttempts(credentials SSHCredentials) error {
	var err error
	numberOfAttempts := 0
	for {
		err = session.Login(credentials)
		if err == nil {
			break
		}

		if numberOfAttempts >= numberObConnectionAttemptsConst {
			return fmt.Errorf("cannot connect to docker container over ssh - %s", err)
		} else {
			numberOfAttempts += 1
			time.Sleep(waitingInSecondsBeforeRetryConst * time.Second)
		}
	}
	return nil
}

// Login tries to login to the remote server. It creates an SSH Session
// if succeed the nil is returned. If not succeed the valid error is returned
//
func (session *SSHSession) Login(credentials SSHCredentials) error {
	sshConfig := &ssh.ClientConfig{
		User: credentials.Username,
		Auth: []ssh.AuthMethod{
			ssh.Password(credentials.Password),
		},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
		BannerCallback: func(message string) error {
			return nil
		},
	}

	IPAndPort := credentials.IPAddress + ":" + strconv.Itoa(int(credentials.Port))
	sshClient, err := ssh.Dial("tcp", IPAndPort, sshConfig)
	if err != nil {
		return fmt.Errorf("cannot connect to server")
	}

	sshSession, err := sshClient.NewSession()
	if err != nil {
		return fmt.Errorf("cannot create new SSH session")
	}

	modes := ssh.TerminalModes{
		ssh.ECHO:          0,     // disable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}

	if err := sshSession.RequestPty("xterm", 80, 40, modes); err != nil {
		err = sshSession.Close()
		if err != nil {
			return fmt.Errorf("tty request failed + cannot close session")
		}
		return fmt.Errorf("cannot request tty")
	}

	if session.StdIn != nil {
		stdin, err := sshSession.StdinPipe()
		if err != nil {
			sshSession.Close()
			return fmt.Errorf("unable to setup stdin for session")
		}
		go io.Copy(stdin, session.StdIn)
	}

	if session.StdOut != nil {
		stdout, err := sshSession.StdoutPipe()
		if err != nil {
			sshSession.Close()
			return fmt.Errorf("unable to setup stdout for session")
		}
		go io.Copy(session.StdOut, stdout)
	}

	if session.StdErr != nil {
		stderr, err := sshSession.StderrPipe()
		if err != nil {
			sshSession.Close()
			return fmt.Errorf("unable to setup stderr for session")
		}
		go io.Copy(session.StdErr, stderr)
	}

	session.sshSession = sshSession
	session.sshClient = sshClient
	return nil
}

// Logout from session if the session is active.
//
func (session *SSHSession) Logout() {
	if session.sshSession == nil {
		return
	}
	session.sshSession.Close()
	time.Sleep(250 * time.Millisecond) // wait for buffer sync. TODO better!
	session.sshSession = nil
}

// IsLoggedIn Check if the session is logged in
//
func (session *SSHSession) IsLoggedIn() bool {
	return session != nil && session.sshSession != nil
}

// SetEnvironment sets the Environment variables for the given session.
//
func (session *SSHSession) SetEnvironment(envMap map[string]string) error {
	if !session.IsLoggedIn() {
		return fmt.Errorf("cannot set environment for not active session")
	}
	for key, value := range envMap {
		err := session.sshSession.Setenv(key, value)
		if err != nil {
			return err
		}
	}
	return nil
}

// Start
// It starts a given command on the remote machine.
// User must call 'Wait' function to wait until command returns.
//
func (session *SSHSession) Start(command string) error {
	err := session.sshSession.Start(command)
	if err != nil {
		return fmt.Errorf("problem while executing program %s", err)
	}
	return nil
}

// Wait
// Wait until command started by 'Start' function ends
//
func (session *SSHSession) Wait() error {
	if !session.IsLoggedIn() {
		return fmt.Errorf("cannot wait for not active session")
	}
	err := session.sshSession.Wait()
	if err != nil {
		return fmt.Errorf("invalid wait - %s", err)
	}
	return nil
}
