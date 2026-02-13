// Package for collecting all constants used by more modules in one place
package constants

import (
	"path/filepath"
)

const (
	// Where to install files on the remote machine
	DockerInstallDirConst = string(filepath.Separator) + "INSTALL"
	// Default SSH port of docker container
	DefaultSSHPort = 1122
	// Name of the docker directory
	DockerDirName  = "docker"
	// Name of the package directory
	PackageDirName = "package"
	// Name of the app directory
	AppDirName = "app"
	// Empty Git commit hash
	EmptyGitCommitHash = ""
	// Absolute path of Package Repository inside docker container
	ContainerPackageRepoPath = "/lfsrepo"
	// Absolute path of sysroot inside docker container
	ContainerSysrootPath = "/sysroot"
)
