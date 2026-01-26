package main

import (
	"github.com/bacpack-system/packager/internal/build"
	"github.com/bacpack-system/packager/internal/config"
	"github.com/bacpack-system/packager/internal/constants"
	"github.com/bacpack-system/packager/internal/context"
	"github.com/bacpack-system/packager/internal/docker"
	"github.com/bacpack-system/packager/internal/log"
	"github.com/bacpack-system/packager/internal/bacpack_package"
	"github.com/bacpack-system/packager/internal/prerequisites"
	"github.com/bacpack-system/packager/internal/process"
	"github.com/bacpack-system/packager/internal/repository"
	"github.com/bacpack-system/packager/internal/ssh"
	"github.com/bacpack-system/packager/internal/sysroot"
	"github.com/bacpack-system/packager/internal/packager_error"
	"fmt"
	"strconv"
	"slices"
	"time"
)

type buildDepList struct {
	dependsMap map[string]*map[string]bool
}

func removeDuplicates(configList *[]config.Config) []config.Config {
	var newConfigList []config.Config
	packageMap := make(map[string]bool)
	for _, cconfig := range *configList {
		packageName := cconfig.Package.Name + ":" + strconv.FormatBool(cconfig.Package.IsDebug)
		exist, _ := packageMap[packageName]
		if exist {
			continue
		}
		packageMap[packageName] = true
		newConfigList = append(newConfigList, cconfig)
	}
	return newConfigList
}

func (list *buildDepList) TopologicalSort(buildMap context.ConfigMapType) ([]config.Config, error) {

	// Map represents 'PackageName: []DependsOnPackageNames'
	var dependsMap map[string]*map[string]bool
	var allDependencies map[string]bool

	dependsMap, allDependencies, err := context.CreateDependsMap(&buildMap)
	if err != nil {
		return []config.Config{}, err
	}

	dependsMapCopy := make(map[string]*map[string]bool, len(dependsMap))
	for key, value := range dependsMap {
		dependsMapCopy[key] = value
	}
	var rootList []string
	for dependencyName, _ := range allDependencies {
		delete(dependsMapCopy, dependencyName)
	}
	for key, _ := range dependsMapCopy {
		rootList = append(rootList, key)
	}

	var sortedDependencies []string
	for key, _ := range dependsMapCopy {
		sortedDependencies = append(sortedDependencies, *list.sortDependencies(key, &dependsMap)...)
	}

	sortedLen := len(sortedDependencies)
	sortedReverse := make([]string, sortedLen)
	for i := sortedLen - 1; i >= 0; i-- {
		sortedReverse[i] = sortedDependencies[sortedLen-i-1]
	}

	var sortedDependenciesConfig []config.Config
	for _, packageName := range sortedReverse {
		sortedDependenciesConfig = append(sortedDependenciesConfig, buildMap[packageName]...)
	}

	return removeDuplicates(&sortedDependenciesConfig), nil
}

func (list *buildDepList) sortDependencies(rootName string, dependsMap *map[string]*map[string]bool) *[]string {
	sorted := []string{rootName}
	rootDeps, found := (*dependsMap)[rootName]

	if !found || len(*rootDeps) == 0 {
		return &sorted
	}

	for packageName, _ := range *rootDeps {
		packageDeps := list.sortDependencies(packageName, dependsMap)
		sorted = append(sorted, *packageDeps...)
	}

	return &sorted
}

// performPreBuildChecks
// Performs Git lfs and sysroot consistency checks. This should be called before builds.
func performPreBuildChecks(
	repo           *repository.GitLFSRepository,
	contextManager *context.ContextManager,
	platformString *bacpack_package.PlatformString,
	imageName      string,
) error {
	logger := log.GetLogger()
	logger.Info("Checking Git Lfs directory consistency")
	err := repo.CheckGitLfsConsistency(contextManager, platformString, imageName)
	if err != nil {
		logger.Error("Git Lfs consistency error - %s", err)
		return packager_error.GitLfsErr
	}
	logger.Info("Checking Sysroot directory consistency")
	err = checkSysrootDirs(platformString)
	if err != nil {
		return err
	}
	return nil
}

// BuildPackage
// process Package mode of the program
func BuildPackage(cmdLine *BuildPackageCmdLineArgs, contextPath string) error {
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
		ForPackage: true,
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
		return buildAllPackages(cmdLine, &contextManager, platformString, repo)
	} else {
		return buildSinglePackage(cmdLine, &contextManager, platformString, repo)
	}
}

// buildAllPackages
// Builds all packages specified in contextPath. Also takes care of building all deps for all
// packages in correct order. It returns nil if everything is ok, or not nil in case of error.
func buildAllPackages(
	cmdLine        *BuildPackageCmdLineArgs,
	contextManager *context.ContextManager,
	platformString *bacpack_package.PlatformString,
	repo           repository.GitLFSRepository,
) error {
	configMap := contextManager.GetAllConfigsMap()

	depsList := buildDepList{}
	configList, err := depsList.TopologicalSort(configMap)
	if err != nil {
		return err
	}

	count := int32(0)
	for _, config := range configList {
		buildConfigs, err := config.GetBuildStructure(
			*cmdLine.DockerImageName,
			platformString,
			uint16(*cmdLine.Port),
			false,
			"",
		)
		if err != nil {
			return err
		}
		if len(buildConfigs) == 0 {
			continue
		}
		count++
		err = buildAndCopyPackage(&buildConfigs, platformString, repo, constants.PackageDirName)
		if err != nil {
			return fmt.Errorf("cannot build package '%s' - %w", config.Package.Name, err)
		}
	}
	if count == 0 {
		return fmt.Errorf("no Packages to build for %s image", *cmdLine.DockerImageName)
	}

	return nil
}

// prepareConfigs
// Returns sorted packageConfigs list.
func sortConfigs(packageConfigs []config.Config) ([]config.Config, error) {
	var configList []config.Config
	defsMap := make(context.ConfigMapType)
	addConfigsToDefsMap(&defsMap, packageConfigs)
	depList := buildDepList{}
	configList, err := depList.TopologicalSort(defsMap)
	if err != nil {
		return []config.Config{}, err
	}
	return configList, nil
}

// prepareConfigsNoBuildDeps
// Returns Config structures only for given Package or App (depends on packageOrApp).
func prepareConfigsNoBuildDeps(
	packageName    string,
	contextManager *context.ContextManager,
	platformString *bacpack_package.PlatformString,
	packageOrApp   string,
) ([]config.Config, error) {
	logger := log.GetLogger()

	if packageOrApp == constants.PackageDirName {
		value, err := isPackageDepsInSysroot(packageName, contextManager, platformString, false)
		if err != nil {
			return []config.Config{}, err
		}
		if !value {
			logger.Error("Package dependencies are not in sysroot")
			return []config.Config{}, packager_error.PackageMissingDependencyErr
		}
	}

	packageConfigs, err := contextManager.GetPackageConfigs(packageName)
	if err != nil {
		return []config.Config{}, err
	}

	return packageConfigs, nil
}

// prepareConfigsBuildDepsOrBuildDepsOn
// Returns Config structures based on --build-deps and --build-deps-on flags.
func prepareConfigsBuildDepsOrBuildDepsOn(
	cmdLine        *BuildPackageCmdLineArgs,
	packageName    string,
	contextManager *context.ContextManager,
	platformString *bacpack_package.PlatformString,
) ([]config.Config, error) {
	var packageConfigs []config.Config

	logger := log.GetLogger()

	if *cmdLine.BuildDeps {
		configs, err := contextManager.GetPackageWithDepsConfigs(packageName)
		if err != nil {
			return []config.Config{}, err
		}
		packageConfigs = append(packageConfigs, configs...)
	} else if *cmdLine.BuildDepsOn || *cmdLine.BuildDepsOnRecursive {
		value, err := isPackageDepsInSysroot(packageName, contextManager, platformString, true)
		if err != nil {
			return []config.Config{}, err
		}
		if !value {
			logger.Error("--build-deps-on(-recursive) set but base package or its dependencies are not in sysroot")
			return []config.Config{}, packager_error.PackageMissingDependencyErr
		}
	}
	if *cmdLine.BuildDepsOn || *cmdLine.BuildDepsOnRecursive {
		configs, err := contextManager.GetPackageWithDepsOnConfigs(packageName, *cmdLine.BuildDepsOnRecursive)
		if err != nil {
			return []config.Config{}, err
		}
		if len(configs) == 0 {
			logger.Warn("No package depends on %s", packageName)
		}
		packageConfigs = append(packageConfigs, configs...)
	}
	return sortConfigs(packageConfigs)
}

// buildSinglePackage
// Builds single package specified by name in cmdLine. Also takes care of building all deps for
// given package in correct order. It returns nil if everything is ok, or not nil in case of error.
func buildSinglePackage(
	cmdLine        *BuildPackageCmdLineArgs,
	contextManager *context.ContextManager,
	platformString *bacpack_package.PlatformString,
	repo           repository.GitLFSRepository,
) error {
	packageName := *cmdLine.Name
	var err error
	var configList []config.Config

	if *cmdLine.BuildDeps || *cmdLine.BuildDepsOn || *cmdLine.BuildDepsOnRecursive {
		configList, err = prepareConfigsBuildDepsOrBuildDepsOn(cmdLine, packageName, contextManager, platformString)
	} else {
		configList, err = prepareConfigsNoBuildDeps(packageName, contextManager, platformString, constants.PackageDirName)
	}
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
			false,
			"",
		)
		if err != nil {
			return err
		}
		err = buildAndCopyPackage(&buildConfigs, platformString, repo, constants.PackageDirName)
		if err != nil {
			return fmt.Errorf("cannot build package '%s' - %w", config.Package.Name, err)
		}
	}
	return nil
}

// addConfigsToDefsMap
// Adds Configs in packageConfigs to defsMap.
func addConfigsToDefsMap(defsMap *context.ConfigMapType, packageConfigs []config.Config) {
	for _, cfg := range packageConfigs {
		packageName := cfg.Package.Name
		_, found := (*defsMap)[packageName]
		if !found {
			(*defsMap)[packageName] = []config.Config{}
		}
		(*defsMap)[packageName] = append((*defsMap)[packageName], cfg)
	}
}

// buildAndCopyPackage
// Builds single Package or App (depends on packageOrApp), takes care of every step of build for
// single package.
func buildAndCopyPackage(
	build          *[]build.Build,
	platformString *bacpack_package.PlatformString,
	repo           repository.GitLFSRepository,
	packageOrApp   string,
) error {
	var err error
	var removeHandler func()

	logger := log.GetLogger()

	for _, buildConfig := range *build {
		logger.Info("Build %s", buildConfig.Package.GetFullPackageName())

		sysroot := sysroot.Sysroot{
			IsDebug:        buildConfig.Package.IsDebug,
			PlatformString: platformString,
		}
		err = prerequisites.Initialize(&sysroot)
		buildConfig.SetSysroot(&sysroot)

		removeHandler = process.SignalHandlerAddHandler(buildConfig.CleanUp)
		var buildPerformed bool
		err, buildPerformed = buildConfig.RunBuild()
		if err != nil {
			return packager_error.BuildErr
		}

		if buildPerformed {
			logger.InfoIndent("Copying to local sysroot directory")
			err = sysroot.CopyToSysroot(buildConfig.GetLocalInstallDirPath(), *buildConfig.BuiltPackage)
			if err != nil {
				break
			}
			
			logger.InfoIndent("Copying to Git repository")
			err = repo.CopyToRepository(*buildConfig.Package, buildConfig.GetLocalInstallDirPath(), packageOrApp)
			if err != nil {
				break
			}
		}

		removeHandler()
		removeHandler = nil
		logger.InfoIndent("Build OK")
	}
	if removeHandler != nil {
		removeHandler()
	}
	return err
}

// checkDockerEnvironment
// Checks if the Docker environment is valid. Checks for read/write permissions and volume read
// permissions.
func checkDockerEnvironment(credentials ssh.SSHCredentials) error {
	testDir := "/packager-test-dir-" + strconv.Itoa(time.Now().Nanosecond())

	logger := log.GetLogger()

	shellEvaluator := ssh.ShellEvaluator{
		Commands: []string{"mkdir " + testDir + " && test -r " + testDir + " && test -w " + testDir + " && rmdir " + testDir},
	}	

	err := shellEvaluator.RunOverSSH(credentials)
	if err != nil {
		logger.ErrorIndent("Cannot create directories, or read or write to them inside Docker container")
		return fmt.Errorf("invalid Docker environment")
	}

	shellEvaluator.Commands = []string{"test -r " + constants.ContainerSysrootPath}

	err = shellEvaluator.RunOverSSH(credentials)
	if err != nil {
		logger.ErrorIndent("Cannot read volume directory inside Docker container")
		return fmt.Errorf("invalid Docker environment")
	}

	return nil
}

// prepareDockerEnvironmentCheck
// Prepares Docker environment for check and returns Docker struct to be used for check.
func prepareDockerEnvironmentCheck(dockerImageName string, dockerPort uint16) (*docker.Docker, error) {
	defaultDocker, err := prerequisites.CreateAndInitialize[docker.Docker](dockerImageName, dockerPort)
	if err != nil {
		return nil, err
	}

	err = sysroot.CreateBaseSysrootDir()
	if err != nil {
		return nil, err
	}

	sysrootPath, err := sysroot.GetBaseSysrootPath()
	if err != nil {
		return nil, err
	}

	err = defaultDocker.SetVolume(sysrootPath, constants.ContainerSysrootPath)
	if err != nil {
		return nil, err
	}
	
	return defaultDocker, nil
}

// checkDockerEnvironmentAndDeterminePlatformString
// Checks if the Docker environment is valid and constructs platform string suitable for sysroot.
func checkDockerEnvironmentAndDeterminePlatformString(dockerImageName string, dockerPort uint16) (*bacpack_package.PlatformString, error) {
	logger := log.GetLogger()
	logger.Info("Checking Docker environment")

	defaultDocker, err := prepareDockerEnvironmentCheck(dockerImageName, dockerPort)
	if err != nil {
		return nil, err
	}

	dockerRun := (*docker.DockerRun)(defaultDocker)
	err = dockerRun.Run()
	if err != nil {
		return nil, err
	}
	removeHandler := dockerRun.GetUndoHandler()
	defer removeHandler()

	sshCreds, err := prerequisites.CreateAndInitialize[ssh.SSHCredentials]()
	if err != nil {
		return nil, err
	}
	sshCreds.Port = uint16(defaultDocker.Port)

	err = checkDockerEnvironment(*sshCreds)
	if err != nil {
		return nil, err
	}

	platformString := bacpack_package.PlatformString{
		Mode: bacpack_package.ModeAuto,
	}

	err = prerequisites.Initialize(&platformString, sshCreds, defaultDocker)
	return &platformString, err
}

// checkSysrootDirs
// Checks if sysroot release and debug directories are empty. If not, prints a warning.
func checkSysrootDirs(platformString *bacpack_package.PlatformString) (error) {
	sysroot := sysroot.Sysroot{
		IsDebug:        false,
		PlatformString: platformString,
	}
	err := prerequisites.Initialize(&sysroot)
	if err != nil {
		return err
	}

	logger := log.GetLogger()
	if !sysroot.IsSysrootDirectoryEmpty() {
		logger.Warn("Sysroot release directory is not empty - the package build may fail")
	}
	sysroot.IsDebug = true
	if !sysroot.IsSysrootDirectoryEmpty() {
		logger.Warn("Sysroot debug directory is not empty - the package build may fail")
	}
	return nil
}

// isPackageWithDepsInSysroot
// Returns true if packageName an its dependencies are in sysroot, else returns false. If the
// checkPackageItself is true, it also checks for presence of Package itself. 
func isPackageDepsInSysroot(
	packageName        string,
	contextManager     *context.ContextManager,
	platformString     *bacpack_package.PlatformString,
	checkPackageItself bool,
) (bool, error) {
	configMap, err := contextManager.GetPackageWithDepsConfigs(packageName)
	if err != nil {
		return false, err
	}

	sysrt := sysroot.Sysroot{
		IsDebug:        false,
		PlatformString: platformString,
	}
	err = prerequisites.Initialize(&sysrt)
	if err != nil {
		return false, err
	}

	for _, config := range configMap {
		if !checkPackageItself && config.Package.Name == packageName {
			continue
		}
		sysrt.IsDebug = config.Package.IsDebug
		builtPackage := sysroot.BuiltPackage {
			Name: config.Package.GetShortPackageName(),
			DirName: sysrt.GetDirNameInSysroot(),
			GitUri: config.Git.URI,
			GitCommitHash: constants.EmptyGitCommitHash,
		}
		if !sysrt.IsPackageInSysroot(builtPackage) {
			return false, nil
		}
	}

	return true, nil
}
