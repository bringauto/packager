package bringauto_build

import (
	"io"
)

type RemoteProcessAttribs struct {
	StdOut   io.Reader
	StdIn    io.Writer
	StdErr   io.Writer
	ExitCode int
}

type CMDLineInterface interface {
	// ConstructCMDLine
	// returns list of commands which can be executed by Bash
	ConstructCMDLine() []string
}