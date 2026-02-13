package docker

import (
	"github.com/bacpack-system/packager/internal/process"
	"github.com/bacpack-system/packager/internal/log"
	"bytes"
	"fmt"
	"regexp"
	"strconv"
)

type DockerRun Docker

// Run starts the container. If succeed
// the container Id is stored and used for stop and other commands that needs it.
func (args *DockerRun) Run() error {

	var outBuff, errBuff bytes.Buffer
	process := process.Process{
		CommandAbsolutePath: DockerExecutablePathConst,
		Args: process.ProcessArgs{
			CmdLineHandler: args,
		},
		StdOut: &outBuff,
		StdErr: &errBuff,
	}
	err := process.Run()

	if err != nil {
		return fmt.Errorf("dockerRun run error - %s, stderr: %s", err, errBuff.String())
	}
	regexp, regexpErr := regexp.CompilePOSIX("^([0-9a-zA-Z]+)")
	if regexpErr != nil {
		return fmt.Errorf("DockerRun run - invalid regexp")
	}
	id := outBuff.String()
	args.containerId = regexp.FindString(id)
	return nil
}

func (args *DockerRun) GetUndoHandler() func() {
	return process.SignalHandlerAddHandler(func() error {
		if args.containerId == "" {
			return nil
		}
		dockerStop := (*DockerStop)(args)
		dockerRm := (*DockerRm)(args)

		logger := log.GetLogger()
		err := dockerStop.Stop()
		if err != nil {
			logger.Error("Can't stop container - %s", err)
		}
		err = dockerRm.RemoveContainer()
		if err != nil {
			logger.Error("Can't remove container - %s", err)
		}
		return nil
	})
}

func (runArgs *DockerRun) GenerateCmdLine() ([]string, error) {
	cmdArgs := make([]string, 0)
	cmdArgs = append(cmdArgs, "run")
	if runArgs.RunAsDaemon {
		cmdArgs = append(cmdArgs, "-d")
	}
	portPair := strconv.Itoa(int(runArgs.Port)) + ":" + strconv.Itoa(sshPort)
	cmdArgs = append(cmdArgs, "-p")
	cmdArgs = append(cmdArgs, portPair)
	for key, value := range runArgs.Volumes {
		volumePair := key + ":" + value + ":Z" // :Z is for SELinux relabeling
		cmdArgs = append(cmdArgs, "-v", volumePair)
	}

	cmdArgs = append(cmdArgs, runArgs.ImageName)
	return cmdArgs, nil
}
