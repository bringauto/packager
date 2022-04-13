package bringauto_docker

import (
	"bringauto/modules/bringauto_process"
	"bytes"
	"fmt"
)

type DockerRm Docker

func (args *DockerRm) RemoveContainer() error {
	if args.containerId == "" {
		return fmt.Errorf("cannot delete unknown container. No container Id provided! Container not started")
	}

	var cerrBuff bytes.Buffer
	process := bringauto_process.Process{
		CommandAbsolutePath: DockerExecutablePathConst,
		Args: bringauto_process.ProcessArgs{
			CmdLineHandler: args,
		},
		StdErr: &cerrBuff,
	}
	err := process.Run()
	if err != nil {
		return fmt.Errorf("DockerRun rm error %s", err)
	}
	return nil
}

func (rmArgs *DockerRm) GenerateCmdLine() ([]string, error) {
	cmdArgs := make([]string, 0)
	cmdArgs = append(cmdArgs, "rm", rmArgs.containerId)
	return cmdArgs, nil
}
