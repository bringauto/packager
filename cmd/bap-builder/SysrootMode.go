package main

import (
	"github.com/bacpack-system/packager/internal/constants"
	"github.com/bacpack-system/packager/internal/context"
	"github.com/bacpack-system/packager/internal/log"
	"github.com/bacpack-system/packager/internal/bacpack_package"
	"github.com/bacpack-system/packager/internal/prerequisites"
	"github.com/bacpack-system/packager/internal/repository"
	"github.com/bacpack-system/packager/internal/packager_error"
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
		return fmt.Errorf("%w - given sysroot directory is not empty", packager_error.CreatingSysrootErr)
	}

	repo := repository.GitLFSRepository{
		GitRepoPath: *cmdLine.Repo,
	}
	err = prerequisites.Initialize(&repo)
	if err != nil {
		return err
	}
	platformString, err := checkDockerEnvironmentAndDeterminePlatformString(*cmdLine.ImageName, uint16(*cmdLine.Port))
	if err != nil {
		return err
	}
	logger := log.GetLogger()

	contextManager := context.ContextManager{
		ContextPath: contextPath,
		ForPackage: true,
	}
	err = prerequisites.Initialize(&contextManager)
	if err != nil {
		logger.Error("Context consistency error - %s", err)
		return packager_error.ContextErr
	}
	logger.Info("Checking Git Lfs directory consistency")
	err = repo.CheckGitLfsConsistency(&contextManager, platformString, *cmdLine.ImageName)
	if err != nil {
		return packager_error.GitLfsErr
	}
	packages, err := contextManager.GetAllPackagesStructs(platformString)
	if err != nil {
		return fmt.Errorf("%w - %s", packager_error.CreatingSysrootErr, err)
	}

	logger.Info("Creating sysroot directory from packages")
	err = unzipAllPackagesToDir(packages, &repo, *cmdLine.Sysroot)
	if err != nil {
		return fmt.Errorf("%w - %s", packager_error.CreatingSysrootErr, err)
	}

	return nil
}

// unzipAllPackagesToDir
// Unzips all given Packages in repo to specified dirPath.
func unzipAllPackagesToDir(packages []bacpack_package.Package, repo *repository.GitLFSRepository, dirPath string) error {
	anyPackageCopied := false
	for _, pack := range packages {
		packPath := path.Join(repo.CreatePath(pack, constants.PackageDirName), pack.GetFullPackageName() + bacpack_package.ZipExt)
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
		return fmt.Errorf("no package from Context is in Git Lfs, so nothing copied to sysroot")
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
