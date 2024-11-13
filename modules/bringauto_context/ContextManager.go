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
// manage all operations on the given context
type ContextManager struct {
	ContextPath string
}

// GetAllPackagesJsonDefPaths
// return all package JSON definitions in the context directory
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
// Returns configs of all packages JSON defintions. Given platformString will be added to all packages.
func (context *ContextManager)  GetAllPackagesConfigs(platformString *bringauto_package.PlatformString) ([]bringauto_package.Package, error) {
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

// GetAllImagesDockerfilePaths
// returns all dockerfile located in the context directory
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
// returns all json definitions for given package
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
// returns all json defintions paths recursively for given package specified by its json definition path
func (context *ContextManager) getAllDepsJsonPaths(packageJsonPath string, visited map[string]struct{}) ([]string, error) {
	var config bringauto_config.Config
	err := config.LoadJSONConfig(packageJsonPath)
	if err != nil {
		return []string{}, fmt.Errorf("couldn't load JSON config from %s path - %s", packageJsonPath, err)
	}
	visited[packageJsonPath] = struct{}{}
	var jsonPathListWithDeps []string
	for _, packageDep := range config.DependsOn {
		packageDepsJsonPaths, err := context.GetPackageJsonDefPaths(packageDep)
		if err != nil {
			return []string{}, fmt.Errorf("couldn't get Json Path of %s package", packageDep)
		}
		var depConfig bringauto_config.Config
		for _, packageDepJsonPath := range packageDepsJsonPaths {
			_, packageVisited := visited[packageDepJsonPath]
			if packageVisited {
				continue
			}
			err := depConfig.LoadJSONConfig(packageDepJsonPath)
			if err != nil {
				return []string{}, fmt.Errorf("couldn't load JSON config from %s path - %s", packageDepJsonPath, err)
			}
			if depConfig.Package.IsDebug != config.Package.IsDebug {
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

	return jsonPathListWithDeps, nil
}

// GetPackageWithDepsJsonDefPaths
// returns all json definitions paths for given package and all its dependencies json definitions paths recursively
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

// GetImageDockerfilePath
// returns Dockerfile path for the given Image locate in the given context
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
// validates context path if the structure in the context directory works
// Return nil if structure is valid, error if the structure is invalid
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
// Get all file in subdirs of rootDir which matches given regexp
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
// get all files from given rootDir which matches given regexp
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
