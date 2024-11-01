package bringauto_repository

import (
	"bringauto/modules/bringauto_config"
	"bringauto/modules/bringauto_package"
	"bringauto/modules/bringauto_prerequisites"
	"bringauto/modules/bringauto_log"
	"bringauto/modules/bringauto_context"
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

func (lfs *GitLFSRepository) RestoreAllChanges() error {
	err := lfs.gitRestoreAll()
	if err != nil {
		return err
	}
	return nil
}

func (lfs *GitLFSRepository) CheckGitLfsConsistency(contextManager *bringauto_context.ContextManager, platformString *bringauto_package.PlatformString) error {
	packages, err := getPackages(contextManager, platformString)

	var expectedPackPaths []string
	for _, pack := range packages {
		packPath := filepath.Join(lfs.createPackagePath(pack) + "/" + pack.GetFullPackageName() + ".zip")
		expectedPackPaths = append(expectedPackPaths, packPath)
	}

	var errorPackPaths []string
	err = filepath.WalkDir(lfs.GitRepoPath, func(path string, d fs.DirEntry, err error) error {
		if d.Name() == ".git" && d.IsDir() {
			return filepath.SkipDir
		}
		if !d.IsDir() {
			if !slices.Contains(expectedPackPaths, path) {
				errorPackPaths = append(errorPackPaths, path)
			} else {
				// Remove element from expected package paths
				index := slices.Index(expectedPackPaths, path)
				expectedPackPaths[index] = expectedPackPaths[len(expectedPackPaths) - 1]
				expectedPackPaths = expectedPackPaths[:len(expectedPackPaths) - 1]
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	err = printErrors(errorPackPaths, expectedPackPaths)
	if err != nil {
		return err
	}
	return nil
}

func getPackages(contextManager *bringauto_context.ContextManager, platformString *bringauto_package.PlatformString) ([]bringauto_package.Package, error) {
	var packConfigs []*bringauto_config.Config
	packageJsonPathMap, err := contextManager.GetAllPackagesJsonDefPaths()
	if err != nil {
		return nil, err
	}
	logger := bringauto_log.GetLogger()
	for _, packageJsonPaths := range packageJsonPathMap {
		for _, packageJsonPath := range packageJsonPaths {
			var config bringauto_config.Config
			err = config.LoadJSONConfig(packageJsonPath)
			if err != nil {
				logger.Warn("Couldn't load JSON config from %s path - %s", packageJsonPath, err)
				continue
			}
			packConfigs = append(packConfigs, &config)
		}
	}
	var packages []bringauto_package.Package
	for _, packConfig := range packConfigs {
		packConfig.Package.PlatformString = *platformString
		packages = append(packages, packConfig.Package)
	}

	return packages, nil
}

func printErrors(errorPackPaths []string, expectedPackPaths []string) error {
	logger := bringauto_log.GetLogger()
	if len(errorPackPaths) > 0 {
		logger.Error("%d packages are not in Json definitions but are in Git Lfs (listing first %d):", len(errorPackPaths), listFileCount)
		for i, errorPackPath := range errorPackPaths {
			if i > listFileCount - 1 {
				break
			}
			logger.ErrorIndent(errorPackPath)
		}
		return fmt.Errorf("packages in Git Lfs are not subset of packages in Json definitions")
	}

	if len(expectedPackPaths) > 0 {
		logger.Warn("Expected %d packages to be in git lfs (listing first %d):", len(expectedPackPaths), listFileCount)
		for i, expectedPackPath := range expectedPackPaths {
			if i > listFileCount - 1 {
				break
			}
			logger.WarnIndent(expectedPackPath)
		}
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
