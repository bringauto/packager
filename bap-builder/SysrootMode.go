package main

import (
	"bringauto/modules/bringauto_context"
	"bringauto/modules/bringauto_log"
	"bringauto/modules/bringauto_package"
	"bringauto/modules/bringauto_prerequisites"
	"bringauto/modules/bringauto_repository"
	"fmt"
	"io"
	"os"
	"path"

	"github.com/mholt/archiver/v3"
)

const (
	ReleasePath = "release"
	DebugPath = "debug"
)

// CreateSysroot
// Creates new sysroot based on Context and Packages in Git Lfs.
func CreateSysroot(cmdLine *CreateSysrootCmdLineArgs, contextPath string) error {
	dirEmpty, err := isDirEmpty(*cmdLine.Sysroot)
	if err != nil {
		return err
	}
	if !dirEmpty {
		return fmt.Errorf("given sysroot directory is not empty")
	}

	repo := bringauto_repository.GitLFSRepository{
		GitRepoPath: *cmdLine.Repo,
	}
	err = bringauto_prerequisites.Initialize(&repo)
	if err != nil {
		return err
	}
	platformString, err := determinePlatformString(*cmdLine.ImageName)
	if err != nil {
		return err
	}
	contextManager := bringauto_context.ContextManager{
		ContextPath: contextPath,
	}
	logger := bringauto_log.GetLogger()
	logger.Info("Checking Git Lfs directory consistency")
	err = repo.CheckGitLfsConsistency(&contextManager, platformString, *cmdLine.ImageName)
	if err != nil {
		return err
	}
	packages, err := contextManager.GetAllPackagesStructs(platformString)
	if err != nil {
		return err
	}

	logger.Info("Creating sysroot directory from packages")
	err = unzipAllPackagesToDir(packages, &repo, *cmdLine.Sysroot)
	if err != nil {
		return err
	}

	return nil
}

// unzipAllPackagesToDir
// Unzips all given Packages in repo to specified dirPath.
func unzipAllPackagesToDir(packages []bringauto_package.Package, repo *bringauto_repository.GitLFSRepository, dirPath string) error {
	anyPackageCopied := false
	for _, pack := range packages {
		packPath := path.Join(repo.CreatePackagePath(pack), pack.GetFullPackageName() + bringauto_package.ZipExt)
		_, err := os.Stat(packPath)
		if err == nil { // Package exists in Git Lfs
			var sysrootPath string
			if pack.IsDebug {
				sysrootPath = path.Join(dirPath, DebugPath)
			} else {
				sysrootPath = path.Join(dirPath, ReleasePath)
			}

			zipArchive := archiver.Zip{
				MkdirAll:             true,
				OverwriteExisting:    false,
				SelectiveCompression: true,
			}
			err := zipArchive.Unarchive(packPath, sysrootPath)
			if err != nil {
				return err
			}
			anyPackageCopied = true
		}
	}
	if !anyPackageCopied {
		logger := bringauto_log.GetLogger()
		logger.Warn("No package from context is in Git Lfs, so nothing copied to sysroot")
	}

	return nil
}

// isDirEmpty
// Checks if the given path is empty.
func isDirEmpty(path string) (bool, error) {
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return true, nil
	} else if err != nil {
		return false, err
	}
	defer f.Close()

	_, err = f.Readdirnames(1)

	if err == io.EOF { // The directory exists, but is empty
		return true, nil
	} else if err != nil {
		return false, err
	}

	return false, nil
}
