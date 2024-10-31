package bringauto_repository

import (
	"bringauto/modules/bringauto_package"
	"bringauto/modules/bringauto_prerequisites"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
)

// GitLFSRepository represents Package repository based on Git LFS
type GitLFSRepository struct {
	GitRepoPath string
	LocalInstallDir string
}

const (
	gitExecutablePath = "/usr/bin/git"
)

func (lfs *GitLFSRepository) FillDefault(args *bringauto_prerequisites.Args) error {
	return nil
}

func (lfs *GitLFSRepository) FillDynamic(args *bringauto_prerequisites.Args) error {
	return nil
}

func (lfs *GitLFSRepository) CheckPrerequisites(*bringauto_prerequisites.Args) error {
	if _, err := os.Stat(lfs.GitRepoPath); os.IsNotExist(err) {
		return fmt.Errorf("package repository '%s' does not exist", lfs.GitRepoPath)
	}
	if _, err := os.Stat(lfs.GitRepoPath + "/.git"); os.IsNotExist(err) {
		return fmt.Errorf("package repository '%s' is not a git repository", lfs.GitRepoPath)
	}

	isStatusEmpty := lfs.gitIsStatusEmpty()
	if !isStatusEmpty {
		return fmt.Errorf("sorry, but the given git root does not have empty `git status`. clean up changes and try again")
	}
	return nil
}

func (lfs *GitLFSRepository) CommitAllChanges() error {
	err := lfs.gitAddAll()
	if err != nil {
		return err
	}
	err = lfs.gitCommit()
	if err != nil {
		return err
	}

	return nil
}

func (lfs *GitLFSRepository) RestoreAllChanges() error {
	err := lfs.gitRestoreAll()
	if err != nil {
		return err
	}
	return nil
}

func (lfs *GitLFSRepository) createPackagePath(pack bringauto_package.Package) string {
	repositoryPath := path.Join(
		pack.PlatformString.String.DistroName,
		pack.PlatformString.String.DistroRelease,
		pack.PlatformString.String.Machine,
		pack.Name,
	)
	return path.Join(lfs.GitRepoPath, repositoryPath)
}

// CopyToRepository copy package to the Git LFS repository.
// Each package is stored in different directory structure represented by
//	PlatformString.DistroName / PlatformString.DistroRelease / PlatformString.Machine / <package>
func (lfs *GitLFSRepository) CopyToRepository(pack bringauto_package.Package, sourceDir string) error {
	archiveDirectory := lfs.createPackagePath(pack)

	var err error
	err = os.MkdirAll(archiveDirectory, 0755)
	if err != nil {
		return err
	}

	err = pack.CreatePackage(sourceDir, archiveDirectory)
	if err != nil {
		return err
	}

	return nil
}

func (lfs *GitLFSRepository) gitIsStatusEmpty() bool {
	var ok, buffer = lfs.prepareAndRun([]string{
		"status",
		"-s",
	},
	)
	if !ok {
		return false
	}
	if buffer.Len() != 0 {
		return false
	}
	return true
}

func (lfs *GitLFSRepository) gitAddAll() error {
	var ok, _ = lfs.prepareAndRun([]string{
		"add",
		"*",
	},
	)
	if !ok {
		return fmt.Errorf("cannot add changes")
	}
	return nil
}

func (lfs *GitLFSRepository) gitCommit() error {
	var ok, _ = lfs.prepareAndRun([]string{
		"commit",
		"-m",
		"Build packages",
	},
	)
	if !ok {
		return fmt.Errorf("cannot commit changes")
	}
	return nil
}

func (lfs *GitLFSRepository) gitRestoreAll() error {
	var ok, _ = lfs.prepareAndRun([]string{
		"restore",
		".",
	},
	)
	if !ok {
		return fmt.Errorf("cannot restore changes")
	}
	return nil
}

func (repo *GitLFSRepository) prepareAndRun(cmdline []string) (bool, *bytes.Buffer) {
	var cmd exec.Cmd
	var outBuffer bytes.Buffer
	var err error

	repoPath := repo.GitRepoPath
	if !filepath.IsAbs(repoPath) {
		workingDir, err := os.Getwd()
		if err != nil {
			return false, nil
		}
		repoPath = path.Join(workingDir, repoPath)
	}

	cmd.Dir = repoPath
	cmdArgs := append([]string{gitExecutablePath}, cmdline...)
	cmd.Args = cmdArgs
	cmd.Path = gitExecutablePath
	cmd.Stdout = &outBuffer
	err = cmd.Run()
	if err != nil {
		fmt.Printf("cannot start command - %s", err)
		return false, &outBuffer
	}
	if cmd.ProcessState.ExitCode() > 0 {
		return false, &outBuffer
	}
	return true, &outBuffer
}
