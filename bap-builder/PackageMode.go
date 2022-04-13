package main

import (
	"bringauto/modules/bringauto_build"
	"bringauto/modules/bringauto_config"
	"bringauto/modules/bringauto_prerequisites"
	"bringauto/modules/bringauto_repository"
	"bringauto/modules/bringauto_sysroot"
	"fmt"
	"log"
	"os"
)

type (
	dependsMapType      map[string]*map[string]bool
	allDependenciesType map[string]bool
	ConfigMapType       map[string][]*bringauto_config.Config
)

type buildDepList struct {
	dependsMap map[string]*map[string]bool
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

	return sortedDependenciesConfig
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
	packagesDefs, err := contextManager.GetAllPackagesJsonDefPaths()
	if err != nil {
		return err
	}

	defsList := make(map[string][]*bringauto_config.Config)
	for _, packageJsonDef := range packagesDefs {
		for _, defdef := range packageJsonDef {
			var config bringauto_config.Config
			err = config.LoadJSONConfig(defdef)
			packageName := config.Package.Name
			_, found := defsList[packageName]
			if !found {
				defsList[packageName] = []*bringauto_config.Config{}
			}
			defsList[packageName] = append(defsList[packageName], &config)
		}

	}
	depsList := buildDepList{}
	configList := depsList.TopologicalSort(defsList)

	for _, config := range configList {
		buildConfigs := config.GetBuildStructure(*cmdLine.DockerImageName)
		if len(buildConfigs) == 0 {
			continue
		}
		log.Println("Build %s", buildConfigs[0].Package.CreatePackageName())
		err = buildAndCopyPackage(cmdLine, &buildConfigs)
		if err != nil {
			panic(fmt.Errorf("cannot build package '%s' - %s", config.Package.Name, err))
		}
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
	packageJsonDefsList, err := contextManager.GetPackageJsonDefPaths(packageName)
	if err != nil {
		return err
	}

	for _, packageJsonDef := range packageJsonDefsList {
		var config bringauto_config.Config
		err = config.LoadJSONConfig(packageJsonDef)
		if err != nil {
			fmt.Fprintf(os.Stderr, "package '%s' JSON config def problem - %s\n", packageName, err)
			continue
		}

		buildConfigs := config.GetBuildStructure(*cmdLine.DockerImageName)
		err = buildAndCopyPackage(cmdLine, &buildConfigs)
		if err != nil {
			fmt.Fprintf(os.Stderr, "cannot build package '%s' - %s\n", packageName, err)
			continue
		}
	}
	return nil
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

	for _, buildConfig := range *build {
		sysroot := bringauto_sysroot.Sysroot{
			IsDebug:        buildConfig.Package.IsDebug,
			PlatformString: &buildConfig.Package.PlatformString,
		}
		err = bringauto_prerequisites.Initialize(&sysroot)

		buildConfig.SetSysroot(&sysroot)
		err = buildConfig.RunBuild()
		if err != nil {
			return err
		}

		err = repo.CopyToRepository(*buildConfig.Package, buildConfig.GetLocalInstallDirPath())
		if err != nil {
			return err
		}

		err = sysroot.CopyToSysroot(buildConfig.GetLocalInstallDirPath())
		if err != nil {
			return err
		}

		err = buildConfig.CleanUp()
		if err != nil {
			return err
		}
	}
	return nil
}
