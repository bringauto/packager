package bringauto_docker

import (
	"bringauto/modules/bringauto_log"
	"bringauto/modules/bringauto_process"
	"bytes"
	"fmt"
	"regexp"
	"strings"
)

type DockerImage Docker

func (dockerImage *DockerImage) ImageExists() bool {
	var err error
	output, err := dockerImage.runDockerImageCommand([]string{"ls"})

	logger := bringauto_log.GetLogger()
	if err != nil {
		logger.Error("cannot start docker process: %s", err)
		return false
	}
	reg, err := regexp.CompilePOSIX("^(?P<container_id>[0-9a-zA-Z]+)\\s+(?P<image_name>[^ ]+)")
	if err != nil {
		logger.Error("cannot compile regexp: %s", err)
	}

	outputLines := strings.Split(output, "\n")
	containersInfoLines := outputLines[1:]

	imageNameIndex := reg.SubexpIndex("image_name")
	for _, line := range containersInfoLines {
		dockerImageLine := reg.FindStringSubmatch(line)
		if dockerImageLine == nil {
			logger.Fatal("Bad imageLine from docker images connect - %s", line)
		}
		imageName := dockerImageLine[imageNameIndex]
		if imageName == dockerImage.ImageName {
			return true
		}
	}

	return false
}

func (dockerImage *DockerImage) runDockerImageCommand(extraArgs []string) (string, error) {
	var stdOut bytes.Buffer
	process := bringauto_process.Process{
		CommandAbsolutePath: DockerExecutablePathConst,
		Args: bringauto_process.ProcessArgs{
			ExtraArgs: &extraArgs,
		},
		StdOut: &stdOut,
	}
	err := process.Run()
	if err != nil {
		return "", fmt.Errorf("DockerRun rm error %s", err)
	}
	return stdOut.String(), nil
}
