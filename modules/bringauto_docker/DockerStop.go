package bringauto_docker

import (
	"bringauto/modules/bringauto_process"
	"bytes"
	"fmt"
)

type DockerStop Docker

// Stop the docker container.
// Docker container must be run be the DockerRun to stop container by DockerStop
func (dockerStop *DockerStop) Stop() error {
	var outBuff, errBuff bytes.Buffer
	process := bringauto_process.Process{
		CommandAbsolutePath: DockerExecutablePathConst,
		Args: bringauto_process.ProcessArgs{
			CmdLineHandler: dockerStop,
		},
		StdOut: &outBuff,
		StdErr: &errBuff,
	}
	err := process.Run()

	if err != nil {
		return fmt.Errorf("DockerRun stop error - %s", err)
	}

	return nil
}

func (dockerStop *DockerStop) GenerateCmdLine() ([]string, error) {
	cmdArgs := make([]string, 0)
	cmdArgs = append(cmdArgs, "stop", dockerStop.containerId)
	return cmdArgs, nil
}
