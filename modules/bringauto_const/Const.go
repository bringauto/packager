// Package for collecting all constants used by more modules in one place
package bringauto_const

import (
	"path/filepath"
)

const (
	// Where to install files on the remote machine
	DockerInstallDirConst = string(filepath.Separator) + "INSTALL"
	// Default SSH port of docker container
	DefaultSSHPort = 1122
)
