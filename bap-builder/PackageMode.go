package main

import (
	"bringauto/modules/bringauto_log"
	"bringauto/modules/bringauto_build"
	"bringauto/modules/bringauto_config"
	"bringauto/modules/bringauto_package"
	"bringauto/modules/bringauto_prerequisites"
	"bringauto/modules/bringauto_repository"
	"bringauto/modules/bringauto_sysroot"
	"fmt"
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

func (list *buildDepList) TopologicalSort(buildMap ConfigMapType) []*bringauto_config.Config {

	// Map represents 'PackageName: []DependsOnPackageNames'
	var dependsMap map[string]*map[string]bool
	var allDependencies map[string]bool

	dependsMap, allDependencies = list.createDependsMap(&buildMap)

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

	return removeDuplicates(&sortedDependenciesConfig)
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

// BuildPackage
// process Package mode of the program
func BuildPackage(cmdLine *BuildPackageCmdLineArgs, contextPath string) error {
	buildAll := cmdLine.All
	if *buildAll {
		return buildAllPackages(cmdLine, contextPath)
	}
	return buildSinglePackage(cmdLine, contextPath)
}

// buildAllPackages
// builds all docker images in the given contextPath.
// It returns nil if everything is ok, or not nil in case of error
func buildAllPackages(cmdLine *BuildPackageCmdLineArgs, contextPath string) error {
	contextManager := ContextManager{
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
	configList := depsList.TopologicalSort(defsMap)

	logger := bringauto_log.GetLogger()

	count := int32(0)
	for _, config := range configList {
		buildConfigs := config.GetBuildStructure(*cmdLine.DockerImageName)
		if len(buildConfigs) == 0 {
			continue
		}
		count++
		logger.Info("Build %s", buildConfigs[0].Package.GetFullPackageName())
		err = buildAndCopyPackage(cmdLine, &buildConfigs)
		if err != nil {
			logger.Fatal("cannot build package '%s' - %s", config.Package.Name, err)
		}
	}
	if count == 0 {
		logger.Warn("Nothing to build. Did you enter correct image name?")
	}

	return nil
}

// buildSinglePackage
// build single package specified by a name
// It returns nil if everything is ok, or not nil in case of error
func buildSinglePackage(cmdLine *BuildPackageCmdLineArgs, contextPath string) error {
	contextManager := ContextManager{
		ContextPath: contextPath,
	}
	packageName := *cmdLine.Name
	var err error
	logger := bringauto_log.GetLogger()

	var configList []*bringauto_config.Config

	if *cmdLine.BuildDeps {
		packageJsonPathList, err := contextManager.GetPackageWithDepsJsonDefPaths(packageName)
		if err != nil {
			return err
		}
		defsMap := make(ConfigMapType)
		addConfigsToDefsMap(&defsMap, packageJsonPathList)
		depList := buildDepList{}
		configList = depList.TopologicalSort(defsMap)
	} else {
		packageJsonPathList, err := contextManager.GetPackageJsonDefPaths(packageName)
		if err != nil {
			return err
		}
		for _, packageJsonPath := range packageJsonPathList {
			var config bringauto_config.Config
			err = config.LoadJSONConfig(packageJsonPath)
			if err != nil {
				logger.Warn("Couldn't load JSON config from %s path - %s", packageJsonPath, err)
				continue
			}
			configList = append(configList, &config)
		}
	}

	for _, config := range configList {
		buildConfigs := config.GetBuildStructure(*cmdLine.DockerImageName)
		err = buildAndCopyPackage(cmdLine, &buildConfigs)
		if err != nil {
			logger.Error("cannot build package '%s' - %s\n", packageName, err)
			continue
		}
	}
	return nil
}

func addConfigsToDefsMap(defsMap *ConfigMapType, packageJsonPathList []string) {
	for _, packageJsonPath := range packageJsonPathList {
		var config bringauto_config.Config
		err := config.LoadJSONConfig(packageJsonPath)
		if err != nil {
			logger := bringauto_log.GetLogger()
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

func buildAndCopyPackage(cmdLine *BuildPackageCmdLineArgs, build *[]bringauto_build.Build) error {
	if *cmdLine.OutputDirMode != OutputDirModeGitLFS {
		return fmt.Errorf("invalid OutputDirmode. Only GitLFS is supported")
	}

	var err error

	repo := bringauto_repository.GitLFSRepository{
		GitRepoPath: *cmdLine.OutputDir,
	}
	err = bringauto_prerequisites.Initialize(&repo)
	if err != nil {
		return err
	}
	logger := bringauto_log.GetLogger()

	for _, buildConfig := range *build {
		platformString, err := determinePlatformString(&buildConfig)
		if err != nil {
			return err
		}

		logger.Info("Build %s", buildConfig.Package.GetFullPackageName())

		sysroot := bringauto_sysroot.Sysroot{
			IsDebug:        buildConfig.Package.IsDebug,
			PlatformString: platformString,
		}
		err = bringauto_prerequisites.Initialize(&sysroot)
		if !sysroot.IsSysrootDirectoryEmpty() {
			logger.WarnIndent("Sysroot directory is not empty - the package build may fail")
		}

		buildConfig.SetSysroot(&sysroot)

		logger.InfoIndent("Run build inside container")
		err = buildConfig.RunBuild()
		if err != nil {
			return err
		}

		logger.InfoIndent("Copy to Git repository")
		err = repo.CopyToRepository(*buildConfig.Package, buildConfig.GetLocalInstallDirPath())
		if err != nil {
			return err
		}

		logger.InfoIndent("Copy to local sysroot directory")
		err = sysroot.CopyToSysroot(buildConfig.GetLocalInstallDirPath())
		if err != nil {
			return err
		}

		err = buildConfig.CleanUp()
		if err != nil {
			return err
		}
		logger.InfoIndent("Build OK")
	}
	return nil
}

// determinePlatformString will construct platform string suitable
// for sysroot.
// For example: the any_machine platformString must be copied to all machine-specific sysroot for
// a given image.
func determinePlatformString(build *bringauto_build.Build) (*bringauto_package.PlatformString, error) {
	platformStringSpecialized := build.Package.PlatformString
	if build.Package.PlatformString.Mode == bringauto_package.ModeAnyMachine {
		platformStringStruct := bringauto_package.PlatformString{
			Mode: bringauto_package.ModeAuto,
		}
		platformStringStruct.Mode = bringauto_package.ModeAuto
		err := bringauto_prerequisites.Initialize[bringauto_package.PlatformString](&platformStringStruct,
			build.SSHCredentials, build.Docker,
		)
		if err != nil {
			return nil, err
		}
		platformStringSpecialized.String.Machine = platformStringStruct.String.Machine
	}
	return &platformStringSpecialized, nil
}
