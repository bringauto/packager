package bringauto_repository

import (
	"bringauto/modules/bringauto_package"
	"bringauto/modules/bringauto_prerequisites"
	"bringauto/modules/bringauto_log"
	"bringauto/modules/bringauto_context"
	"bringauto/modules/bringauto_config"
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"slices"
)

// GitLFSRepository represents Package repository based on Git LFS
type GitLFSRepository struct {
	GitRepoPath string
}

const (
	gitExecutablePath = "/usr/bin/git"
	// Count of files which will be list in warnings
	listFileCount = 10
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

// RestoreAllChanges
// Restores all changes in repository and cleans all untracked changes.
func (lfs *GitLFSRepository) RestoreAllChanges() error {
	err := lfs.gitRestoreAll()
	if err != nil {
		return err
	}
	err = lfs.gitCleanAll()
	if err != nil {
		return err
	}
	return nil
}

func dividePackagesForCurrentImage(allConfigs []*bringauto_config.Config, imageName string) ([]bringauto_package.Package, []bringauto_package.Package) {
	var packagesForImage []bringauto_package.Package
	var packagesNotForImage []bringauto_package.Package

	for _, config := range allConfigs {
		if slices.Contains(config.DockerMatrix.ImageNames, imageName) {
			packagesForImage = append(packagesForImage, config.Package)
		} else {
			packagesNotForImage = append(packagesNotForImage, config.Package)
		}
	}

	return packagesForImage, packagesNotForImage
}

func (lfs *GitLFSRepository) CheckGitLfsConsistency(contextManager *bringauto_context.ContextManager, platformString *bringauto_package.PlatformString, imageName string) error {
	packConfigs, err := contextManager.GetAllPackagesConfigs(platformString)
	if err != nil {
		return err
	}

	packagesForImage, packagesNotForImage := dividePackagesForCurrentImage(packConfigs, imageName)

	var expectedPackForImagePaths, expectedPackNotForImagePaths []string
	for _, pack := range packagesForImage {
		packPath := filepath.Join(lfs.CreatePackagePath(pack) + "/" + pack.GetFullPackageName() + ".zip")
		expectedPackForImagePaths = append(expectedPackForImagePaths, packPath)
	}
	for _, pack := range packagesNotForImage {
		packPath := filepath.Join(lfs.CreatePackagePath(pack) + "/" + pack.GetFullPackageName() + ".zip")
		expectedPackNotForImagePaths = append(expectedPackNotForImagePaths, packPath)
	}

	lookupPath := filepath.Join(lfs.GitRepoPath, platformString.String.DistroName, platformString.String.DistroRelease, platformString.String.Machine)

	var errorPackPaths []string
	_, err = os.Stat(lookupPath)
	if !os.IsNotExist(err) {
		err = filepath.WalkDir(lookupPath, func(path string, d fs.DirEntry, err error) error {
			if d.Name() == ".git" && d.IsDir() {
				return filepath.SkipDir
			}
			if !d.IsDir() {
				if !slices.Contains(expectedPackForImagePaths, path) {
					errorPackPaths = append(errorPackPaths, path)
				} else {
					// Remove element from expected package paths
					index := slices.Index(expectedPackForImagePaths, path)
					expectedPackForImagePaths[index] = expectedPackForImagePaths[len(expectedPackForImagePaths) - 1]
					expectedPackForImagePaths = expectedPackForImagePaths[:len(expectedPackForImagePaths) - 1]
				}
			}
			return nil
		})
		if err != nil {
			return err
		}
	}
	err = printErrors(errorPackPaths, expectedPackForImagePaths, expectedPackNotForImagePaths)
	if err != nil {
		return err
	}
	return nil
}

func printErrors(errorPackPaths []string, expectedPackForImagePaths []string, expectedPackNotForImagePaths []string) error {
	logger := bringauto_log.GetLogger()
	if len(errorPackPaths) > 0 {
		logger.Error("%d packages are not in Json definitions but are in Git Lfs (listing first %d):", len(errorPackPaths), listFileCount)
		for i, errorPackPath := range errorPackPaths {
			if i > listFileCount - 1 {
				break
			}
			logger.ErrorIndent("%s", errorPackPath)
		}
		return fmt.Errorf("packages in Git Lfs are not subset of packages in Json definitions")
	}

	if len(expectedPackForImagePaths) > 0 {
		logger.Warn("Expected %d packages (built for target image) to be in git lfs (listing first %d):", len(expectedPackForImagePaths), listFileCount)
		for i, expectedPackForImagePath := range expectedPackForImagePaths {
			if i > listFileCount - 1 {
				break
			}
			logger.WarnIndent("%s", expectedPackForImagePath)
		}
	}
	if len(expectedPackNotForImagePaths) > 0 {
		logger.Warn("%d packages are in context but are not built for target image (listing first %d):", len(expectedPackNotForImagePaths), listFileCount)
		for i, expectedPackNotForImagePath := range expectedPackNotForImagePaths {
			if i > listFileCount - 1 {
				break
			}
			logger.WarnIndent("%s", expectedPackNotForImagePath)
		}
	}
	return nil
}

func (lfs *GitLFSRepository) CreatePackagePath(pack bringauto_package.Package) string {
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
	archiveDirectory := lfs.CreatePackagePath(pack)

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
		return fmt.Errorf("cannot add changes in Git Lfs")
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
		return fmt.Errorf("cannot commit changes in Git Lfs")
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
		return fmt.Errorf("cannot restore changes in Git Lfs")
	}
	return nil
}

func (lfs *GitLFSRepository) gitCleanAll() error {
	var ok, _ = lfs.prepareAndRun([]string{
		"clean",
		"-f",
		".",
	},
	)
	if !ok {
		return fmt.Errorf("cannot clean changes in Git Lfs")
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
		return false, &outBuffer
	}
	if cmd.ProcessState.ExitCode() > 0 {
		return false, &outBuffer
	}
	return true, &outBuffer
}
