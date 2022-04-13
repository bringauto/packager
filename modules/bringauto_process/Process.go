package bringauto_process

import (
	"fmt"
	"io"
	"os/exec"
)

// Process represents standard process
type Process struct {
	// StdOut writer to capture stdout of the process
	StdOut io.Writer
	// StdErr writer to capture stdout of the process
	StdErr io.Writer
	// StdIn writer to capture stdout of the process
	StdIn io.Reader
	// Process CMD line arguments.
	Args ProcessArgs
	// Absolute path od co executable to be run
	CommandAbsolutePath string
}

// ProcessArgs
// Represents Arguments passed to process.
// Exactly one field of this struct can be filled up.
//
// There is no need to specify executable path as the first element of the argument
// array.
type ProcessArgs struct {
	// CmdLineHandler represents a handler that can construct CMDline arguments for the process
	CmdLineHandler CmdLineHandlerInterface
	// ExtraArgs are CMD line arguments specified as list of strings.
	ExtraArgs *[]string
}

// Run the process. It returns nil in case of success
// or not-nit in case of error
func (process *Process) Run() error {
	var cmd exec.Cmd
	var cmdArgs []string

	if process.Args.ExtraArgs != nil && process.Args.CmdLineHandler != nil {
		return fmt.Errorf("process - ExtraArgs and CmdLineHandler cannot be set at one time")
	}

	if process.Args.ExtraArgs != nil {
		cmdArgs = *process.Args.ExtraArgs
	} else if process.Args.CmdLineHandler != nil {
		cmdArgs, _ = process.Args.CmdLineHandler.GenerateCmdLine()
	}

	cmdArgs = append([]string{process.CommandAbsolutePath}, cmdArgs...)
	cmd.Args = cmdArgs
	cmd.Path = process.CommandAbsolutePath
	cmd.Stderr = process.StdErr
	cmd.Stdout = process.StdOut
	cmd.Stdin = process.StdIn
	err := cmd.Run()
	if err != nil {
		return err
	}
	if cmd.ProcessState.ExitCode() > 0 {
		return err
	}
	return nil
}
