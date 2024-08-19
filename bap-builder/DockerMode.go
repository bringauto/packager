package main

import (
	"bringauto/modules/bringauto_log"
	"bringauto/modules/bringauto_docker"
	"fmt"
	"path"
)

// BuildDockerImage
// process Docker mode of cmd line
//
func BuildDockerImage(cmdLine *BuildImageCmdLineArgs, contextPath string) error {
	buildAll := cmdLine.All
	if *buildAll {
		return buildAllDockerImages(contextPath)
	}
	return buildSingleDockerImage(contextPath, *cmdLine.Name)
}

// buildAllDockerImages
// builds all docker images in the given contextPath.
// It returns nil if everything is ok, or not nil in case of error
//
func buildAllDockerImages(contextPath string) error {
	contextManager := ContextManager{
		ContextPath: contextPath,
	}
	dockerfileList, err := contextManager.GetAllImagesDockerfilePaths()
	if err != nil {
		return err
	}

	logger := bringauto_log.GetLogger()

	for imageName, dockerfilePathList := range dockerfileList {
		if len(dockerfilePathList) != 1 {
			logger.Warn("Bug: multiple Dockerfile present for same image name %s", imageName)
			continue
		}
		dockerfileDir := path.Dir(dockerfilePathList[0])
		dockerBuild := bringauto_docker.DockerBuild{
			DockerfileDir: dockerfileDir,
			Tag:           imageName,
		}
		err = dockerBuild.Build()
		if err != nil {
			return fmt.Errorf("Build failed for %s image", imageName)
		}
	}
	return nil
}

// buildSingleDockerImage
// builds a single docker image specified by a name.
//
func buildSingleDockerImage(contextPath string, imageName string) error {
	contextManager := ContextManager{
		ContextPath: contextPath,
	}
	dockerfilePath, err := contextManager.GetImageDockerfilePath(imageName)
	if err != nil {
		return err
	}

	dockerfileDir := path.Dir(dockerfilePath)
	dockerBuild := bringauto_docker.DockerBuild{
		DockerfileDir: dockerfileDir,
		Tag:           imageName,
	}
	buildOk := dockerBuild.Build()
	if buildOk != nil {
		return fmt.Errorf("Build failed for %s image", imageName)
	}
	return nil
}
