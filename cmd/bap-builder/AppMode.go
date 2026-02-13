package main

import (
	"github.com/bacpack-system/packager/internal/packager_error"
	"github.com/bacpack-system/packager/internal/constants"
	"github.com/bacpack-system/packager/internal/context"
	"github.com/bacpack-system/packager/internal/log"
	"github.com/bacpack-system/packager/internal/bacpack_package"
	"github.com/bacpack-system/packager/internal/prerequisites"
	"github.com/bacpack-system/packager/internal/process"
	"github.com/bacpack-system/packager/internal/repository"
	"github.com/bacpack-system/packager/internal/sysroot"
	"fmt"
	"slices"
)

// BuildApp
func BuildApp(cmdLine *BuildAppCmdLineArgs, contextPath string) error {
	platformString, err := checkDockerEnvironmentAndDeterminePlatformString(*cmdLine.DockerImageName, uint16(*cmdLine.Port))
	if err != nil {
		return err
	}
	repo := repository.GitLFSRepository{
		GitRepoPath: *cmdLine.OutputDir,
	}
	err = prerequisites.Initialize(&repo)
	if err != nil {
		return err
	}
	contextManager := context.ContextManager{
		ContextPath: contextPath,
		ForPackage: false,
	}
	err = prerequisites.Initialize(&contextManager)
	if err != nil {
		logger := log.GetLogger()
		logger.Error("Context consistency error - %s", err)
		return packager_error.ContextErr
	}
	err = performPreBuildChecks(&repo, &contextManager, platformString, *cmdLine.DockerImageName)
	if err != nil {
		return err
	}

	handleRemover := process.SignalHandlerAddHandler(repo.RestoreAllChanges)
	defer handleRemover()

	if *cmdLine.All {
		return buildAllApps(cmdLine, &contextManager, platformString, repo)
	} else {
		return buildSingleApp(cmdLine, &contextManager, platformString, repo)
	}
}

// buildAllApps
// Builds all Apps specified in contextPath. Returns nil if everything is ok, else returns error.
func buildAllApps(
	cmdLine        *BuildAppCmdLineArgs,
	contextManager *context.ContextManager,
	platformString *bacpack_package.PlatformString,
	repo           repository.GitLFSRepository,
) error {
	configMap := contextManager.GetAllConfigsMap()

	count := int32(0)
	for appName := range configMap {
		for _, config := range configMap[appName] {
			buildConfigs, err := config.GetBuildStructure(
				*cmdLine.DockerImageName,
				platformString,
				uint16(*cmdLine.Port),
				*cmdLine.UseLocalRepo,
				repo.GitRepoPath,
			)
			if err != nil {
				return err
			}
			if len(buildConfigs) == 0 {
				continue
			}
			count++
			err = buildAndCopyPackage(&buildConfigs, platformString, repo, constants.AppDirName)
			if err != nil {
				return fmt.Errorf("cannot build App '%s' - %w", config.Package.Name, err)
			}
		}
		err := sysroot.RemoveInstallSysroot()
		if err != nil {
			return fmt.Errorf("cannot remove install sysroot directory")
		}
	}
	if count == 0 {
		return fmt.Errorf("no Apps to build for %s image", *cmdLine.DockerImageName)
	}

	return nil
}

// buildSingleApp
// Builds single App specified by name in cmdLine. Returns nil if everything is ok, else returns error.
func buildSingleApp(
	cmdLine        *BuildAppCmdLineArgs,
	contextManager *context.ContextManager,
	platformString *bacpack_package.PlatformString,
	repo           repository.GitLFSRepository,
) error {
	configList, err := prepareConfigsNoBuildDeps(*cmdLine.Name, contextManager, platformString, constants.AppDirName)
	if err != nil {
		return err
	}
	if len(configList) == 0 {
		return fmt.Errorf("nothing to build")
	}
	for _, config := range configList {
		if !slices.Contains(config.DockerMatrix.ImageNames, *cmdLine.DockerImageName) {
			return fmt.Errorf("'%s' does not support %s image", config.Package.Name, *cmdLine.DockerImageName)
		}
		buildConfigs, err := config.GetBuildStructure(
			*cmdLine.DockerImageName,
			platformString,
			uint16(*cmdLine.Port),
			*cmdLine.UseLocalRepo,
			repo.GitRepoPath,
		)
		if err != nil {
			return err
		}
		err = buildAndCopyPackage(&buildConfigs, platformString, repo, constants.AppDirName)
		if err != nil {
			return fmt.Errorf("cannot build App '%s' - %w", *cmdLine.Name, err)
		}
	}
	return nil
}
