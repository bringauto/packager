package bringauto_docker

import (
	"bringauto/modules/bringauto_log"
	"bytes"
	"fmt"
	"os/exec"
)

// DockerBuild
// options needed for a build a docker image
type DockerBuild struct {
	// directory where the Dockerfile is located
	DockerfileDir string
	// tag which will be used to tag image after build
	Tag string
}

// Build given docker image
func (dockerBuild *DockerBuild) Build() error {
	if dockerBuild.DockerfileDir == "" {
		return fmt.Errorf("DockerBuild - DockerfileDir is empty")
	}
	logger := bringauto_log.GetLogger()
	logger.Info("Build Docker Image: %s", dockerBuild.Tag)

	var ok, _, err = dockerBuild.prepareAndRun(prepareBuildArgs)
	if !ok {
		return fmt.Errorf("DockerBuild build error - %s", err)
	}
	return nil
}

func (dockerBuild *DockerBuild) prepareAndRun(f func(build *DockerBuild) []string) (bool, *bytes.Buffer, *bytes.Buffer) {
	var cmd exec.Cmd
	var outBuffer, errBuffer bytes.Buffer
	cmdArgs := f(dockerBuild)
	cmdArgs = append([]string{DockerExecutablePathConst}, cmdArgs...)
	cmd.Args = cmdArgs
	cmd.Path = DockerExecutablePathConst
	cmd.Stderr = &errBuffer
	cmd.Stdout = &outBuffer
	err := cmd.Run()
	if err != nil {
		return false, &outBuffer, &errBuffer
	}
	if cmd.ProcessState.ExitCode() > 0 {
		return false, &outBuffer, &errBuffer
	}
	return true, &outBuffer, &errBuffer
}

func prepareBuildArgs(dockerBuild *DockerBuild) []string {
	cmdArgs := make([]string, 0)
	cmdArgs = append(cmdArgs, "build")
	cmdArgs = append(cmdArgs, dockerBuild.DockerfileDir)
	if dockerBuild.DockerfileDir != "" {
		cmdArgs = append(cmdArgs, "--tag", dockerBuild.Tag)
	}
	return cmdArgs
}
