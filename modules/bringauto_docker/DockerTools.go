package bringauto_docker

import (
	"bringauto/modules/bringauto_process"
	"bringauto/modules/bringauto_const"
	"bytes"
	"strconv"
)

func IsDefaultPortAvailable() bool {
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

	process.Run()

	return outBuff.Len() == 0
}
