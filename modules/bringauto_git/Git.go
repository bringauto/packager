package bringauto_git

import (
	"fmt"
	"strings"
)

type Git struct {
	URI       string
	Revision  string
	ClonePath string `json:"-"`
}

type GitClone struct {
	Git
}

type GitCheckout struct {
	Git
}

type GitSubmoduleUpdate struct {
	Git
}

const (
	GitExecutablePath = "git"
)

func (args *GitClone) ConstructCMDLine() []string {
	validateGITPath(args.ClonePath)
	cmd := []string{
		GitExecutablePath,
		"clone",
		"--recursive",
		args.URI,
		args.ClonePath,
	}
	return []string{strings.Join(cmd, " ")}
}

func (args *GitCheckout) ConstructCMDLine() []string {
	validateGITPath(args.ClonePath)
	cmd := []string{
		GitExecutablePath,
		"checkout",
		args.Revision,
	}
	return []string{
		"pushd " + args.ClonePath,
		strings.Join(cmd, " "),
		"popd",
	}
}

func (args *GitSubmoduleUpdate) ConstructCMDLine() []string {
	validateGITPath(args.ClonePath)
	cmd := []string{
		GitExecutablePath,
		"submodule",
		"update",
		"--init",
		"--recursive",
	}
	return []string{
		"pushd " + args.ClonePath,
		strings.Join(cmd, " "),
		"popd",
	}
}

func validateGITPath(path string) {
	if path == "" {
		panic(fmt.Errorf("git path is empty"))
	}
}
