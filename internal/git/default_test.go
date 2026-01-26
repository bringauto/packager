package git_test

import (
	"github.com/bacpack-system/packager/internal/git"
	"reflect"
	"strings"
	"testing"
)

func TestGitClone_ConstructCMDLine(t *testing.T) {
	gitClone := git.GitClone{}
	gitClone.URI = "TestUri"
	gitClone.Revision = "master"
	gitClone.ClonePath = "local"
	cmdLine := gitClone.ConstructCMDLine()
	validCmdLine := []string{
		git.GitExecutablePath,
		"clone",
		"--no-checkout",
		gitClone.URI,
		gitClone.ClonePath,
	}
	cmdLineValid := reflect.DeepEqual(cmdLine[0], strings.Join(validCmdLine, " "))
	if !cmdLineValid {
		t.Errorf("git clone CMD line is not valid!")
	}
}

func TestGitCheckout_ConstructCMDLine(t *testing.T) {
	gitCheckout := git.GitCheckout{}
	gitCheckout.URI = "TestUri"
	gitCheckout.Revision = "master"
	gitCheckout.ClonePath = "local"
	cmdLine := gitCheckout.ConstructCMDLine()
	gitCmdLine := []string{
		git.GitExecutablePath,
		"checkout",
		gitCheckout.Revision,
	}
	validCmdLine := []string{
		"pushd " + gitCheckout.ClonePath,
		strings.Join(gitCmdLine, " "),
		"popd",
	}
	cmdLineValid := reflect.DeepEqual(cmdLine, validCmdLine)
	if !cmdLineValid {
		t.Errorf("git checkout CMD line is not valid!")
	}
}

func TestGitSubmoduleUpdate_ConstructCMDLine(t *testing.T) {
	gitSU := git.GitSubmoduleUpdate{}
	gitSU.URI = "TestUri"
	gitSU.Revision = "master"
	gitSU.ClonePath = "local"
	cmdLine := gitSU.ConstructCMDLine()
	gitCmdLine := []string{
		git.GitExecutablePath,
		"submodule",
		"update",
		"--init",
		"--recursive",
	}
	validCmdLine := []string{
		"pushd " + gitSU.ClonePath,
		strings.Join(gitCmdLine, " "),
		"popd",
	}
	cmdLineValid := reflect.DeepEqual(cmdLine, validCmdLine)
	if !cmdLineValid {
		t.Errorf("git update CMD line is not valid!")
	}
}
