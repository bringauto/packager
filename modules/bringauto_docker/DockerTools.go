package bringauto_docker

import (
	"bringauto/modules/bringauto_const"
	"bringauto/modules/bringauto_process"
	"bytes"
	"fmt"
	"strconv"
)

// IsDefaultPortAvailable
// Returns true if default port for docker is available, else returns false.
// When false is returned, the error contains message from the docker command.
func IsDefaultPortAvailable() (bool, error) {
	var outBuff, errBuff bytes.Buffer

	process := bringauto_process.Process{
		CommandAbsolutePath: DockerExecutablePathConst,
		Args: bringauto_process.ProcessArgs{
			ExtraArgs: &[]string{
				"container",
				"ls",
				"--filter",
				"publish=" + strconv.Itoa(bringauto_const.DefaultSSHPort),
				"--format",
				"{{.ID}}{{.Ports}}",
			},
		},
		StdOut: &outBuff,
		StdErr: &errBuff,
	}

	err := process.Run()
	if err != nil {
		return false, fmt.Errorf(errBuff.String())
	}

	return outBuff.Len() == 0, nil
}
