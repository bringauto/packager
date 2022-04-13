package bringauto_ssh

import (
	"io"
	"os"
)

// ShellEvaluator run commands by the bash thru the SSH.
//
type ShellEvaluator struct {
	// Commands to execute
	Commands []string
	// Environments variables to set
	Env map[string]string
	//
	StdOut io.Writer
}

// RunOverSSH
// Runs command over SSH.
//
// All commands specified in Commands are run be Bash by a one Bash session
//
// All environment variables are preserved across command run and can be used by other
// subsequent commands.
func (shell *ShellEvaluator) RunOverSSH(credentials SSHCredentials) error {
	var err error
	pipeReader, pipeWriter := io.Pipe()
	session := SSHSession{
		StdOut: shell.StdOut,
		StdErr: os.Stderr,
		StdIn:  pipeReader,
	}

	err = session.LoginMultipleAttempts(credentials)
	if err != nil {
		return err
	}

	err = session.Start("bash")
	if err != nil {
		return err
	}

	// We cannot use SSHSession/SSHConnection setenv function
	// for Env. setting because SetEnv must be configured at the Server side
	var env []string
	for envName, envValue := range shell.Env {
		env = append(env, "export "+envName+"="+escapeVariableValue(envValue))
	}

	commands := shell.Commands
	commands = append(env, commands...)
	commands = append(commands, "exit")

	for _, value := range commands {
		_, err = pipeWriter.Write([]byte(value + "\n\r"))
		if err != nil {
			return err
		}
	}
	err = session.Wait()
	if err != nil {
		return err
	}
	return nil
}

func escapeVariableValue(varValue string) string {
	return "\"" + varValue + "\""
}
