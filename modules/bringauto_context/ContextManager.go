package bringauto_context

import (
	"bringauto/modules/bringauto_config"
	"bringauto/modules/bringauto_const"
	"bringauto/modules/bringauto_log"
	"bringauto/modules/bringauto_package"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
)

// ContextManager
// Manages all operations on the given Context.
type ContextManager struct {
	ContextPath string
}

// GetAllPackagesJsonDefPaths
// Returns all Package Configs in the context directory.
func (context *ContextManager) GetAllPackagesJsonDefPaths() (map[string][]string, error) {
	var err error
	err = context.validateContextPath()
	if err != nil {
		return nil, err
	}

	packageDir := path.Join(context.ContextPath, bringauto_const.PackageDirName)

	reg, err := regexp.CompilePOSIX("^.*\\.json$")
	if err != nil {
		return nil, fmt.Errorf("cannot compile regext for .json extension")
	}

	packageJsonList, err := getAllFilesInSubdirByRegexp(packageDir, reg)
	return packageJsonList, err
}

// GetAllPackagesConfigs
// Returns Config structs of all Package Configs. If platformString is not nil, it is added to all
// Packages.
func (context *ContextManager) GetAllPackagesConfigs(platformString *bringauto_package.PlatformString) ([]*bringauto_config.Config, error) {
	var packConfigs []*bringauto_config.Config
	packageJsonPathMap, err := context.GetAllPackagesJsonDefPaths()
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
			if platformString != nil {
				config.Package.PlatformString = *platformString
			}
			packConfigs = append(packConfigs, &config)
		}
	}
	return packConfigs, nil
}

// GetAllPackagesConfigs
// Returns Package structs of all Package Configs. If platformString is not nil, it is added to all
// Packages.
func (context *ContextManager) GetAllPackagesStructs(platformString *bringauto_package.PlatformString) ([]bringauto_package.Package, error) {
	packConfigs, err := context.GetAllPackagesConfigs(platformString)
	if err != nil {
		return []bringauto_package.Package{}, err
	}

	var packages []bringauto_package.Package
	for _, packConfig := range packConfigs {
		packages = append(packages, packConfig.Package)
	}

	return packages, nil
}

// GetAllImagesDockerfilePaths
// Returns all Dockerfile paths located in the Context directory.
func (context *ContextManager) GetAllImagesDockerfilePaths() (map[string][]string, error) {
	var err error
	err = context.validateContextPath()
	if err != nil {
		return nil, err
	}

	imageDir := path.Join(context.ContextPath, bringauto_const.DockerDirName)

	reg, err := regexp.CompilePOSIX("^Dockerfile$")
	if err != nil {
		return nil, fmt.Errorf("cannot compile regexp for matchiing Dockerfile")
	}

	dockerfileList, err := getAllFilesInSubdirByRegexp(imageDir, reg)

	return dockerfileList, err
}

// GetPackageJsonDefPaths
// Returns all Config for given Package.
func (context *ContextManager) GetPackageJsonDefPaths(packageName string) ([]string, error) {
	var err error
	err = context.validateContextPath()
	if err != nil {
		return []string{}, err
	}
	packageBasePath := path.Join(context.ContextPath, bringauto_const.PackageDirName, packageName)

	packageBasePathStat, err := os.Stat(packageBasePath)
	if os.IsNotExist(err) {
		return []string{}, fmt.Errorf("package does not exist, please check the name")
	}
	if !packageBasePathStat.IsDir() {
		return []string{}, fmt.Errorf("package does not exist. It seems like an ordinary file")
	}

	reg, err := regexp.CompilePOSIX("^.*\\.json$")
	if err != nil {
		return []string{}, fmt.Errorf("cannot compile regexp for .json extension")
	}

	packageDefs, err := getAllFilesInDirByRegexp(packageBasePath, reg)
	if err != nil {
		return []string{}, fmt.Errorf("cannot get definitions for package '%s'", packageName)
	}

	return packageDefs, nil
}

// getAllDepsJsonPaths
// Returns all Config paths for given Package (specified with packageJsonPath) and all Configs for
// its dependencies recursively. For tracking of circular dependencies, the visited map must be
// initialized before function call.
func (context *ContextManager) getAllDepsJsonPaths(packageJsonPath string, visited map[string]struct{}) ([]string, error) {
	var config bringauto_config.Config
	err := config.LoadJSONConfig(packageJsonPath)
	if err != nil {
		return []string{}, fmt.Errorf("couldn't load JSON config from %s path - %s", packageJsonPath, err)
	}
	visited[packageJsonPath] = struct{}{}
	addedPackages := 0
	var jsonPathListWithDeps []string
	for _, packageDep := range config.DependsOn {
		packageDepsJsonPaths, err := context.GetPackageJsonDefPaths(packageDep)
		if err != nil {
			return []string{}, fmt.Errorf("couldn't get Json Path of %s package", packageDep)
		}
		var depConfig bringauto_config.Config
		for _, packageDepJsonPath := range packageDepsJsonPaths {
			err := depConfig.LoadJSONConfig(packageDepJsonPath)
			if err != nil {
				return []string{}, fmt.Errorf("couldn't load JSON config from %s path - %s", packageDepJsonPath, err)
			}
			if depConfig.Package.IsDebug != config.Package.IsDebug {
				continue
			}
			addedPackages++
			_, packageVisited := visited[packageDepJsonPath]
			if packageVisited {
				continue
			}
			jsonPathListWithDeps = append(jsonPathListWithDeps, packageDepJsonPath)
			jsonPathListWithDepsTmp, err := context.getAllDepsJsonPaths(packageDepJsonPath, visited)
			if err != nil {
				return []string{}, err
			}
			jsonPathListWithDeps = append(jsonPathListWithDeps, jsonPathListWithDepsTmp...)
		}
	}

	if addedPackages < len(config.DependsOn) {
		return []string{}, fmt.Errorf("package %s dependencies do not have package with same build type", config.Package.Name)
	}

	return jsonPathListWithDeps, nil
}

// getAllDepsOnJsonPaths
// Returns all Config paths of Packages which depends on Package specified with config. If
// recursively is set to true, it is done recursively. For tracking of circular dependencies,
// the visited map must be initialized before function call.
func (context *ContextManager) getAllDepsOnJsonPaths(config bringauto_config.Config, visited map[string]struct{}, recursively bool) ([]string, error) {
	packConfigs, err := context.GetAllPackagesConfigs(nil)
	if err != nil {
		return []string{}, err
	}
	visited[config.Package.Name] = struct{}{}
	var packsToBuild []string
	for _, packConfig := range packConfigs {
		if (packConfig.Package.Name == config.Package.Name ||
	 	  	packConfig.Package.IsDebug != config.Package.IsDebug){
			continue
		}
		for _, dep := range packConfig.DependsOn {
			if dep == config.Package.Name {
				_, packageVisited := visited[packConfig.Package.Name]
				if packageVisited {
					break
				}
				context.addDependsOnPackagesToBuild(&packsToBuild, packConfig, visited, recursively)
				break
			}
		}
	}

	return packsToBuild, nil
}

func (context *ContextManager) addDependsOnPackagesToBuild(packsToBuild *[]string, packConfig *bringauto_config.Config, visited map[string]struct{}, recursively bool) error {
	packWithDeps, err := context.GetPackageWithDepsJsonDefPaths(packConfig.Package.Name)
	if err != nil {
		return err
	}
	*packsToBuild = append(*packsToBuild, packWithDeps...)
	if recursively {
		packsDepsOnRecursive, err := context.getAllDepsOnJsonPaths(*packConfig, visited, true)
		if err != nil {
			return err
		}
		*packsToBuild = append(*packsToBuild, packsDepsOnRecursive...)
	}
	return nil
}

// GetPackageWithDepsJsonDefPaths
// Returns all Config paths for given Package and all its dependencies Config paths recursively.
func (context *ContextManager) GetPackageWithDepsJsonDefPaths(packageName string) ([]string, error) {
	packageDefs, err := context.GetPackageJsonDefPaths(packageName)
	if err != nil {
		return []string{}, fmt.Errorf("cannot get config paths for package '%s' - %s", packageName, err)
	}
	var packageDeps []string
	visitedPackages := make(map[string]struct{})
	for _, packageDef := range packageDefs {
		packageDepsTmp, err := context.getAllDepsJsonPaths(packageDef, visitedPackages)
		if err != nil {
			return []string{}, err
		}
		packageDeps = append(packageDeps, packageDepsTmp...)
	}

	packageDefs = append(packageDefs, packageDeps...)

	return packageDefs, nil
}

// GetPackageWithDepsOnJsonDefPaths
// Returns all Config paths which depends on given Package and all its dependencies Config paths
// without package (packageName) itself and its dependencies. If recursively is set to true, it is
// done recursively.
func (context *ContextManager) GetDepsOnJsonDefPaths(packageName string, recursively bool) ([]string, error) {
	packageDefs, err := context.GetPackageJsonDefPaths(packageName)
	if err != nil {
		return []string{}, err
	}
	var packsToBuild []string
	visitedPackages := make(map[string]struct{})
	for _, packageDef := range packageDefs {
		var config bringauto_config.Config
		err := config.LoadJSONConfig(packageDef)
		if err != nil {
			return []string{}, fmt.Errorf("couldn't load JSON config from %s path - %s", packageDef, err)
		}
		packageDepsTmp, err := context.getAllDepsOnJsonPaths(config, visitedPackages, recursively)
		if err != nil {
			return []string{}, err
		}
		packsToBuild = append(packsToBuild, packageDepsTmp...)
	}

	packsToRemove, err := context.GetPackageWithDepsJsonDefPaths(packageName)
	if err != nil {
		return []string{}, err
	}
	packsToBuild = removeStrings(packsToBuild, packsToRemove)
	return removeDuplicates(packsToBuild), nil
}

// removeStrings
// Removes strList2 strings from strList1.
func removeStrings(strList1 []string, strList2 []string) []string {
	for _, str2 := range strList2 {
		strList1 = removeString(strList1, str2)
	}
	return strList1
}

// removeString
// Removes str string from strList1.
func removeString(strList1 []string, str string) []string {
	i := 0
	for _, str1 := range strList1 {
		if str1 != str {
			strList1[i] = str1
			i++
		}
	}
	return strList1[:i]
}

// removeDuplicates
// Removes duplicate entries in strList.
func removeDuplicates(strList []string) []string {
	keys := make(map[string]struct{})
    list := []string{}
    for _, item := range strList {
    	_, value := keys[item]
        if !value {
            keys[item] = struct{}{}
            list = append(list, item)
        }
    }
    return list
}

// GetImageDockerfilePath
// Returns Dockerfile path for the given Image locate in the given Context.
func (context *ContextManager) GetImageDockerfilePath(imageName string) (string, error) {
	var err error
	err = context.validateContextPath()
	if err != nil {
		return "", err
	}
	dockerImageBasePath := path.Join(context.ContextPath, bringauto_const.DockerDirName, imageName)

	dockerImageBasePathStat, err := os.Stat(dockerImageBasePath)
	if os.IsNotExist(err) {
		return "", fmt.Errorf("docker image definition does not exist, please check the name")
	}
	if !dockerImageBasePathStat.IsDir() {
		return "", fmt.Errorf("docker image definition does not exist. It seems like an ordinary file")
	}

	dockerfilePath := filepath.Join(dockerImageBasePath, "Dockerfile")
	if _, err = os.Stat(dockerfilePath); os.IsNotExist(err) {
		return "", fmt.Errorf("dockerfile for the given image does not exist")
	}

	return dockerfilePath, nil
}

// validateContextPath
// Validates Context path if the structure in the Context directory works
// Return nil if structure is valid, error if the structure is invalid.
func (context *ContextManager) validateContextPath() error {
	var err error
	ContextStat, err := os.Stat(context.ContextPath)
	if os.IsNotExist(err) {
		return fmt.Errorf("context path does not exist - %s\n", context.ContextPath)
	}
	if !ContextStat.IsDir() {
		return fmt.Errorf("context path is not a directory - %s\n", context.ContextPath)
	}

	dockerDirPath := path.Join(context.ContextPath, bringauto_const.DockerDirName)
	DockerStat, err := os.Stat(dockerDirPath)
	if os.IsNotExist(err) {
		return fmt.Errorf("docker dir path does not exist - %s\n", dockerDirPath)
	}
	if !DockerStat.IsDir() {
		return fmt.Errorf("docker path is not a directory - %s\n", dockerDirPath)
	}

	packageDirPath := path.Join(context.ContextPath, bringauto_const.PackageDirName)
	packageStat, err := os.Stat(packageDirPath)
	if os.IsNotExist(err) {
		return fmt.Errorf("package path does not exist - %s\n", packageDirPath)
	}
	if !packageStat.IsDir() {
		return fmt.Errorf("package path is not a directory - %s\n", packageDirPath)
	}

	return nil
}

// getAllFilesInDirByRegexp
// Get all file in subdirs of rootDir which matches given regexp.
func getAllFilesInSubdirByRegexp(rootDir string, reg *regexp.Regexp) (map[string][]string, error) {
	acceptedFileList := map[string][]string{}
	walkError := filepath.WalkDir(rootDir, func(item string, d fs.DirEntry, err error) error {
		if d.Name() == path.Base(rootDir) {
			return nil
		}
		packageName := d.Name()
		packageBaseDir := path.Join(rootDir, d.Name())
		packageFileDefs, err := getAllFilesInDirByRegexp(packageBaseDir, reg)
		if err != nil {
			return nil
		}
		acceptedFileList[packageName] = packageFileDefs
		return nil
	},
	)
	return acceptedFileList, walkError
}

// getAllFilesInDirByRegexp
// Get all files from given rootDir which matches given regexp.
func getAllFilesInDirByRegexp(rootDir string, reg *regexp.Regexp) ([]string, error) {
	var acceptedFileList []string
	dirEntryList, err := os.ReadDir(rootDir)
	if err != nil {
		return []string{}, fmt.Errorf("cannot list dir %s", rootDir)
	}

	for _, dirEntry := range dirEntryList {
		packageNameOk := reg.MatchString(dirEntry.Name())
		if !packageNameOk {
			continue
		}
		acceptedFileList = append(acceptedFileList, path.Join(rootDir, dirEntry.Name()))
	}
	return acceptedFileList, nil
}
