package bringauto_build

import (
	"path/filepath"
)

const (
	// Where to clone a git repository on the remote machine
	dockerGitCloneDirConst = string(filepath.Separator) + "git"
	// Where to copy file from remote machine before the package is created
	localInstallDirNameConst = string(filepath.Separator) + "localInstall"
)
