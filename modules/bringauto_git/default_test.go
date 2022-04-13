package bringauto_git_test

import (
	"bringauto/modules/bringauto_git"
	"reflect"
	"strings"
	"testing"
)

func TestGitClone_ConstructCMDLine(t *testing.T) {
	git := bringauto_git.GitClone{}
	git.URI = "TestUri"
	git.Revision = "master"
	git.ClonePath = "local"
	cmdLine := git.ConstructCMDLine()
	validCmdLine := []string{
		bringauto_git.GitExecutablePath,
		"clone",
		"--recursive",
		git.URI,
		git.ClonePath,
	}
	cmdLineValid := reflect.DeepEqual(cmdLine[0], strings.Join(validCmdLine, " "))
	if !cmdLineValid {
		t.Errorf("git clone CMD line is not valid!")
	}
}

func TestGitCheckout_ConstructCMDLine(t *testing.T) {
	git := bringauto_git.GitCheckout{}
	git.URI = "TestUri"
	git.Revision = "master"
	git.ClonePath = "local"
	cmdLine := git.ConstructCMDLine()
	gitCmdLine := []string{
		bringauto_git.GitExecutablePath,
		"checkout",
		git.Revision,
	}
	validCmdLine := []string{
		"pushd " + git.ClonePath,
		strings.Join(gitCmdLine, " "),
		"popd",
	}
	cmdLineValid := reflect.DeepEqual(cmdLine, validCmdLine)
	if !cmdLineValid {
		t.Errorf("git checkout CMD line is not valid!")
	}
}

func TestGitSubmoduleUpdate_ConstructCMDLine(t *testing.T) {
	git := bringauto_git.GitSubmoduleUpdate{}
	git.URI = "TestUri"
	git.Revision = "master"
	git.ClonePath = "local"
	cmdLine := git.ConstructCMDLine()
	gitCmdLine := []string{
		bringauto_git.GitExecutablePath,
		"submodule",
		"update",
		"--init",
		"--recursive",
	}
	validCmdLine := []string{
		"pushd " + git.ClonePath,
		strings.Join(gitCmdLine, " "),
		"popd",
	}
	cmdLineValid := reflect.DeepEqual(cmdLine, validCmdLine)
	if !cmdLineValid {
		t.Errorf("git update CMD line is not valid!")
	}
}
