package main

import (
	"bringauto/modules/bringauto_build"
	"bringauto/modules/bringauto_config"
	"bringauto/modules/bringauto_const"
	"bringauto/modules/bringauto_context"
	"bringauto/modules/bringauto_docker"
	"bringauto/modules/bringauto_log"
	"bringauto/modules/bringauto_package"
	"bringauto/modules/bringauto_prerequisites"
	"bringauto/modules/bringauto_process"
	"bringauto/modules/bringauto_repository"
	"bringauto/modules/bringauto_ssh"
	"bringauto/modules/bringauto_sysroot"
	"fmt"
	"io/fs"
	"path/filepath"
	"strconv"
)

type (
	dependsMapType      map[string]*map[string]bool
	allDependenciesType map[string]bool
	ConfigMapType       map[string][]*bringauto_config.Config
)

type buildDepList struct {
	dependsMap map[string]*map[string]bool
}

func removeDuplicates(configList *[]*bringauto_config.Config) []*bringauto_config.Config {
	var newConfigList []*bringauto_config.Config
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

// checkForCircularDependency
// Checks for circular dependency in defsMap. If there is one, returns error with message
// and problematic packages, else returns nil.
func checkForCircularDependency(dependsMap map[string]*map[string]bool) error {
	visited := make(map[string]bool)

	for packageName := range dependsMap {
		cycleDetected, cycleString := detectCycle(packageName, dependsMap, visited)
		if cycleDetected {
			return fmt.Errorf("circular dependency detected - %s", packageName + " -> " + cycleString)
		}
		// Clearing recursion stack after one path through graph was checked
		for visitedPackage := range visited {
			visited[visitedPackage] = false
		}
	}
	return nil
}

// detectCycle
// Detects cycle between package dependencies in one path through graph. visited is current
// recursion stack and dependsMap is whole graph representation. packageName is root node where
// cycle detection should start.
func detectCycle(packageName string, dependsMap map[string]*map[string]bool, visited map[string]bool) (bool, string) {
	visited[packageName] = true
	depsMap, found := dependsMap[packageName]
	if found {
		for depPackageName := range *depsMap {
			if visited[depPackageName] {
				return true, depPackageName
			} else {
				cycleDetected, cycleString := detectCycle(depPackageName, dependsMap, visited)
				if cycleDetected {
					return cycleDetected, depPackageName + " -> " + cycleString
				}
			}
		}
	}
	visited[packageName] = false
	return false, ""
}

func (list *buildDepList) TopologicalSort(buildMap ConfigMapType) ([]*bringauto_config.Config, error) {

	// Map represents 'PackageName: []DependsOnPackageNames'
	var dependsMap map[string]*map[string]bool
	var allDependencies map[string]bool

	dependsMap, allDependencies = list.createDependsMap(&buildMap)
	err := checkForCircularDependency(dependsMap)
	if err != nil  {
		return []*bringauto_config.Config{}, err
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

	var sortedDependenciesConfig []*bringauto_config.Config
	for _, packageName := range sortedReverse {
		sortedDependenciesConfig = append(sortedDependenciesConfig, buildMap[packageName]...)
	}

	return removeDuplicates(&sortedDependenciesConfig), nil
}

func (list *buildDepList) createDependsMap(buildMap *ConfigMapType) (dependsMapType, allDependenciesType) {
	allDependencies := make(map[string]bool)
	dependsMap := make(map[string]*map[string]bool)

	for _, configArray := range *buildMap {
		if len(configArray) == 0 {
			panic("invalid entry in dependency map")
		}
		packageName := configArray[0].Package.Name
		item, found := dependsMap[packageName]
		if !found {
			item = &map[string]bool{}
			dependsMap[packageName] = item
		}
		for _, config := range configArray {
			if len(config.DependsOn) == 0 {
				continue
			}
			for _, v := range config.DependsOn {
				(*item)[v] = true
				allDependencies[v] = true
			}
		}
	}
	return dependsMap, allDependencies
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


// checkContextDirConsistency
// Checks if all directories in contextPath have same name as Package names from JSON definitions
// inside this directory. If not, returns error with description, else returns nil. Also returns error
// if the Package JSON definition can't be loaded.
func checkContextDirConsistency(contextPath string) error {
	packageContextPath := filepath.Join(contextPath, bringauto_const.PackageDirName)
	err := filepath.WalkDir(packageContextPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			var config bringauto_config.Config
			err = config.LoadJSONConfig(path)
			if err != nil {
				return fmt.Errorf("couldn't load JSON config from %s path", path)
			}
			dirName := filepath.Base(filepath.Dir(path))
			if config.Package.Name != dirName {
				return fmt.Errorf("directory name (%s) is different from package name (%s)", dirName, config.Package.Name)
			}
		}
		return nil
	})

	return err
}

func performPreBuildChecks(contextPath string, repo *bringauto_repository.GitLFSRepository, platformString *bringauto_package.PlatformString) error {
	logger := bringauto_log.GetLogger()
	logger.Info("Checking context directory (%s) consistency", contextPath)
	err := checkContextDirConsistency(contextPath)
	if err != nil {
		return fmt.Errorf("package context directory consistency check failed: %s", err)
	}
	contextManager := bringauto_context.ContextManager{
		ContextPath: contextPath,
	}
	logger.Info("Checking Git Lfs directory consistency")
	err = repo.CheckGitLfsConsistency(&contextManager, platformString)
	if err != nil {
		return err
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
	platformString, err := determinePlatformString(*cmdLine.DockerImageName)
	if err != nil {
		return err
	}
	repo := bringauto_repository.GitLFSRepository{
		GitRepoPath: *cmdLine.OutputDir,
	}
	err = bringauto_prerequisites.Initialize(&repo)
	if err != nil {
		return err
	}
	err = performPreBuildChecks(contextPath, &repo, platformString)
	if err != nil {
		return err
	}

	handleRemover := bringauto_process.SignalHandlerAddHandler(repo.RestoreAllChanges)
	if *cmdLine.All {
		err = buildAllPackages(cmdLine, contextPath, platformString, repo)
	} else {
		err = buildSinglePackage(cmdLine, contextPath, platformString, repo)
	}
	if err != nil {
		handleRemover()
		return err
	}
	repo.CommitAllChanges()
	handleRemover()
	return nil
}

// buildAllPackages
// Builds all packages specified in contextPath. Also takes care of building all deps for all
// packages in correct order. It returns nil if everything is ok, or not nil in case of error.
func buildAllPackages(
	cmdLine        *BuildPackageCmdLineArgs,
	contextPath    string,
	platformString *bringauto_package.PlatformString,
	repo           bringauto_repository.GitLFSRepository,
) error {
	contextManager := bringauto_context.ContextManager{
		ContextPath: contextPath,
	}
	packageJsonPathMap, err := contextManager.GetAllPackagesJsonDefPaths()
	if err != nil {
		return err
	}

	defsMap := make(ConfigMapType)
	for _, packageJsonPathList := range packageJsonPathMap {
		addConfigsToDefsMap(&defsMap, packageJsonPathList)
	}
	depsList := buildDepList{}
	configList, err := depsList.TopologicalSort(defsMap)
	if err != nil {
		return err
	}

	logger := bringauto_log.GetLogger()

	count := int32(0)
	for _, config := range configList {
		buildConfigs := config.GetBuildStructure(*cmdLine.DockerImageName, platformString)
		if len(buildConfigs) == 0 {
			continue
		}
		count++
		err = buildAndCopyPackage(cmdLine, &buildConfigs, platformString, repo)
		if err != nil {
			return fmt.Errorf("cannot build package '%s' - %s", config.Package.Name, err)
		}
	}
	if count == 0 {
		logger.Warn("Nothing to build. Did you enter correct image name?")
	}

	return nil
}

// prepareConfigs
// Returns Config structures list based on given jsonPaths.
func prepareConfigs(packageJsonPaths []string) ([]*bringauto_config.Config, error) {
	var configList []*bringauto_config.Config
	defsMap := make(ConfigMapType)
	addConfigsToDefsMap(&defsMap, packageJsonPaths)
	depList := buildDepList{}
	configList, err := depList.TopologicalSort(defsMap)
	if err != nil {
		return []*bringauto_config.Config{}, err
	}
	return configList, nil
}

// prepareConfigsNoBuildDeps
// Returns Config structures only for given package.
func prepareConfigsNoBuildDeps(packageName string, contextManager *bringauto_context.ContextManager) ([]*bringauto_config.Config, error) {
	var configList []*bringauto_config.Config
	packageJsonPaths, err := contextManager.GetPackageJsonDefPaths(packageName)
	if err != nil {
		return []*bringauto_config.Config{}, err
	}
	for _, packageJsonPath := range packageJsonPaths {
		var config bringauto_config.Config
		err = config.LoadJSONConfig(packageJsonPath)
		if err != nil {
			logger := bringauto_log.GetLogger()
			logger.Warn("Couldn't load JSON config from %s path - %s", packageJsonPath, err)
			continue
		}
		configList = append(configList, &config)
	}
	return configList, nil
}

// prepareConfigsBuildDepsOrBuildDepsOn
// Returns Config structures based on --build-deps and --build-deps-on flags.
func prepareConfigsBuildDepsOrBuildDepsOn(
	cmdLine        *BuildPackageCmdLineArgs,
	packageName    string,
	contextManager *bringauto_context.ContextManager,
	platformString *bringauto_package.PlatformString,
) ([]*bringauto_config.Config, error) {
	var packageJsonPaths []string
	if *cmdLine.BuildDeps {
		paths, err := contextManager.GetPackageWithDepsJsonDefPaths(packageName)
		if err != nil {
			return []*bringauto_config.Config{}, err
		}
		packageJsonPaths = append(packageJsonPaths, paths...)
	} else if *cmdLine.BuildDepsOn || *cmdLine.BuildDepsOnRecursive {
		value, err := isPackageWithDepsInSysroot(packageName, contextManager, platformString)
		if err != nil {
			return []*bringauto_config.Config{}, err
		}
		if !value {
			err = fmt.Errorf("--build-deps-on(-recursive) set but base package or its dependencies are not in sysroot")
			return []*bringauto_config.Config{}, err
		}
	}
	if *cmdLine.BuildDepsOn || *cmdLine.BuildDepsOnRecursive {
		paths, err := contextManager.GetDepsOnJsonDefPaths(packageName, *cmdLine.BuildDepsOnRecursive)
		if err != nil {
			return []*bringauto_config.Config{}, err
		}
		packageJsonPaths = append(packageJsonPaths, paths...)
	}
	return prepareConfigs(packageJsonPaths)
}

// buildSinglePackage
// Builds single package specified by name in cmdLine. Also takes care of building all deps for
// given package in correct order. It returns nil if everything is ok, or not nil in case of error.
func buildSinglePackage(
	cmdLine        *BuildPackageCmdLineArgs,
	contextPath    string,
	platformString *bringauto_package.PlatformString,
	repo           bringauto_repository.GitLFSRepository,
) error {
	contextManager := bringauto_context.ContextManager{
		ContextPath: contextPath,
	}
	packageName := *cmdLine.Name
	var err error
	var configList []*bringauto_config.Config

	if *cmdLine.BuildDeps || *cmdLine.BuildDepsOn || *cmdLine.BuildDepsOnRecursive {
		configList, err = prepareConfigsBuildDepsOrBuildDepsOn(cmdLine, packageName, &contextManager, platformString)
	} else {
		configList, err = prepareConfigsNoBuildDeps(packageName, &contextManager)
	}
	if err != nil {
		return err
	}

	for _, config := range configList {
		buildConfigs := config.GetBuildStructure(*cmdLine.DockerImageName, platformString)
		err = buildAndCopyPackage(cmdLine, &buildConfigs, platformString, repo)
		if err != nil {
			return fmt.Errorf("cannot build package '%s' - %s", packageName, err)
		}
	}
	return nil
}

// addConfigsToDefsMap
// Adds all configs in packageJsonPathList to defsMap.
func addConfigsToDefsMap(defsMap *ConfigMapType, packageJsonPathList []string) {
	logger := bringauto_log.GetLogger()
	for _, packageJsonPath := range packageJsonPathList {
		var config bringauto_config.Config
		err := config.LoadJSONConfig(packageJsonPath)
		if err != nil {
			logger.Error("Couldn't load JSON config from %s path - %s", packageJsonPath, err)
			continue
		}
		packageName := config.Package.Name
		_, found := (*defsMap)[packageName]
		if !found {
			(*defsMap)[packageName] = []*bringauto_config.Config{}
		}
		(*defsMap)[packageName] = append((*defsMap)[packageName], &config)
	}
}

// buildAndCopyPackage
// Builds single package, takes care of every step of build for single package.
func buildAndCopyPackage(
	cmdLine *BuildPackageCmdLineArgs,
	build *[]bringauto_build.Build,
	platformString *bringauto_package.PlatformString,
	repo bringauto_repository.GitLFSRepository,
) error {
	if *cmdLine.OutputDirMode != OutputDirModeGitLFS {
		return fmt.Errorf("invalid OutputDirmode. Only GitLFS is supported")
	}

	var err error
	var removeHandler func()

	logger := bringauto_log.GetLogger()

	for _, buildConfig := range *build {
		logger.Info("Build %s", buildConfig.Package.GetFullPackageName())

		sysroot := bringauto_sysroot.Sysroot{
			IsDebug:        buildConfig.Package.IsDebug,
			PlatformString: platformString,
		}
		err = bringauto_prerequisites.Initialize(&sysroot)
		buildConfig.SetSysroot(&sysroot)

		logger.InfoIndent("Run build inside container")
		removeHandler = bringauto_process.SignalHandlerAddHandler(buildConfig.CleanUp)
		err = buildConfig.RunBuild()
		if err != nil {
			return err
		}

		logger.InfoIndent("Copying to Git repository")

		err = repo.CopyToRepository(*buildConfig.Package, buildConfig.GetLocalInstallDirPath())
		if err != nil {
			break
		}

		logger.InfoIndent("Copying to local sysroot directory")
		err = sysroot.CopyToSysroot(buildConfig.GetLocalInstallDirPath(), buildConfig.Package.GetShortPackageName())
		if err != nil {
			break
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

// determinePlatformString
// Will construct platform string suitable for sysroot.
func determinePlatformString(dockerImageName string) (*bringauto_package.PlatformString, error) {
	defaultDocker := bringauto_prerequisites.CreateAndInitialize[bringauto_docker.Docker](dockerImageName)
	defaultDocker.ImageName = dockerImageName

	sshCreds := bringauto_prerequisites.CreateAndInitialize[bringauto_ssh.SSHCredentials]()

	platformString := bringauto_package.PlatformString{
		Mode: bringauto_package.ModeAuto,
	}

	err := bringauto_prerequisites.Initialize[bringauto_package.PlatformString](&platformString, sshCreds, defaultDocker)
	return &platformString, err
}

// checkSysrootDirs
// Checks if sysroot release and debug directories are empty. If not, prints a warning.
func checkSysrootDirs(platformString *bringauto_package.PlatformString) (error) {
	sysroot := bringauto_sysroot.Sysroot{
		IsDebug:        false,
		PlatformString: platformString,
	}
	err := bringauto_prerequisites.Initialize(&sysroot)
	if err != nil {
		return err
	}

	logger := bringauto_log.GetLogger()
	if !sysroot.IsSysrootDirectoryEmpty() {
		logger.Warn("Sysroot release directory is not empty - the package build may fail")
	}
	sysroot.IsDebug = true
	if !sysroot.IsSysrootDirectoryEmpty() {
		logger.Warn("Sysroot debug directory is not empty - the package build may fail")
	}
	return nil
}

// arePackagesInSysroot
// Returns true if packageName an its dependencies are in sysroot, else returns false.
func isPackageWithDepsInSysroot(packageName string, contextManager *bringauto_context.ContextManager, platformString *bringauto_package.PlatformString) (bool, error) {
	packageJsonPaths, err := contextManager.GetPackageWithDepsJsonDefPaths(packageName)
	if err != nil {
		return false, err
	}
	configList, err := prepareConfigs(packageJsonPaths)
	if err != nil {
		return false, err
	}

	sysroot := bringauto_sysroot.Sysroot{
		IsDebug:        false,
		PlatformString: platformString,
	}
	err = bringauto_prerequisites.Initialize(&sysroot)

	for _, config := range configList {
		packName := config.Package.GetShortPackageName()
		if !sysroot.IsPackageInSysroot(packName) {
			return false, nil
		}
	}

	return true, nil
}
