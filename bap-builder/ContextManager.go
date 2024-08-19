package main

import (
	"bringauto/modules/bringauto_config"
	"bringauto/modules/bringauto_log"
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

	packageDir := path.Join(context.ContextPath, PackageDirectoryNameConst)

	reg, err := regexp.CompilePOSIX("^.*\\.json$")
	if err != nil {
		return nil, fmt.Errorf("cannot compile regext for .json extension")
	}

	packageJsonList, err := getAllFilesInSubdirByRegexp(packageDir, reg)
	return packageJsonList, err
}

// GetAllImagesDockerfilePaths
// returns all dockerfile located in the context directory
func (context *ContextManager) GetAllImagesDockerfilePaths() (map[string][]string, error) {
	var err error
	err = context.validateContextPath()
	if err != nil {
		return nil, err
	}

	imageDir := path.Join(context.ContextPath, DockerDirectoryNameConst)

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
	packageBasePath := path.Join(context.ContextPath, PackageDirectoryNameConst, packageName)

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

func (context *ContextManager) getAllDepsJsonPaths(packageJsonPath string) []string {
	var config bringauto_config.Config
	logger := bringauto_log.GetLogger()
	err := config.LoadJSONConfig(packageJsonPath)
	if err != nil {
		logger.Warn("Couldn't load JSON config from %s path - %s", packageJsonPath, err)
		return []string{}
	}
	var jsonPathListWithDeps []string
	for _, packageDep := range config.DependsOn {
		packageDepsJsonPaths, err := context.GetPackageJsonDefPaths(packageDep)
		if err != nil {
			logger.Warn("Couldn't get Json Path of %s package", packageDep)
			continue
		}
		var depConfig bringauto_config.Config
		for _, packageDepJsonPath := range packageDepsJsonPaths {
			err := depConfig.LoadJSONConfig(packageDepJsonPath)
			if err != nil {
				logger.Warn("Couldn't load JSON config from %s path - %s", packageJsonPath, err)
				continue
			}
			if depConfig.Package.IsDebug == config.Package.IsDebug {
				jsonPathListWithDeps = append(jsonPathListWithDeps, packageDepJsonPath)
				jsonPathListWithDeps = append(jsonPathListWithDeps, context.getAllDepsJsonPaths(packageDepJsonPath)...)
			}
		}
	}

	return jsonPathListWithDeps
}

// GetPackageWithDepsJsonDefPaths
// returns all json definitions for given package and all its dependencies json definitions recursively
func (context *ContextManager) GetPackageWithDepsJsonDefPaths(packageName string) ([]string, error) {
	packageDefs, err := context.GetPackageJsonDefPaths(packageName)
	if err != nil {
		return []string{}, fmt.Errorf("cannot get config paths for package '%s'", packageName)
	}
	var packageDeps []string
	for _, packageDef := range packageDefs {
		packageDeps = append(packageDeps, context.getAllDepsJsonPaths(packageDef)...)
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
	dockerImageBasePath := path.Join(context.ContextPath, DockerDirectoryNameConst, imageName)

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

	dockerDirPath := path.Join(context.ContextPath, DockerDirectoryNameConst)
	DockerStat, err := os.Stat(dockerDirPath)
	if os.IsNotExist(err) {
		return fmt.Errorf("docker dir path does not exist - %s\n", dockerDirPath)
	}
	if !DockerStat.IsDir() {
		return fmt.Errorf("docker path is not a directory - %s\n", dockerDirPath)
	}

	packageDirPath := path.Join(context.ContextPath, PackageDirectoryNameConst)
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
