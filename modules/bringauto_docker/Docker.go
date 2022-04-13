package bringauto_docker

import (
	"bringauto/modules/bringauto_prerequisites"
	"bringauto/modules/bringauto_process"
	"fmt"
	"os"
)

const (
	defaultImageNameConst = "debian11"
)

// Docker
type Docker struct {
	// ImageName - tag or image hash
	ImageName string
	// Ports mapping between host and container
	// in manner map[int]int { <host>:<container> }
	Ports map[int]int `json:"-"`
	// Volumes map a host directory (represented by absolute path)
	// to the directory inside the docker container
	// in manner map[string]string { <host_volume_abs_path>:<> }
	Volumes map[string]string `json:"-"`
	// If true docker command will run in non-blocking mode - as a daemon.
	RunAsDaemon bool `json:"-"`
	containerId string
}

func (docker *Docker) FillDefault(*bringauto_prerequisites.Args) error {
	*docker = Docker{
		Volumes:     map[string]string{},
		RunAsDaemon: true,
		ImageName:   defaultImageNameConst,
		Ports: map[int]int{
			1122: 22,
		},
	}
	return nil
}

func (docker *Docker) FillDynamic(*bringauto_prerequisites.Args) error {
	return nil
}

// CheckPrerequisites
// It checks if the docker is installed and can be run by given user.
// Function returns nil if Docker installation is ok, not nil of the problem is recognized
func (docker *Docker) CheckPrerequisites(*bringauto_prerequisites.Args) error {
	process := bringauto_process.Process{
		CommandAbsolutePath: DockerExecutablePathConst,
		Args: bringauto_process.ProcessArgs{
			ExtraArgs: &[]string{
				"images",
			},
		},
	}
	err := process.Run()
	if err != nil {
		return err
	}

	for hostVolume, _ := range docker.Volumes {
		if _, err = os.Stat(hostVolume); os.IsNotExist(err) {
			return fmt.Errorf("connot mount non existent directory as volume: '%s'", hostVolume)
		}
	}

	return nil
}

// SetVolume set volume mapping for a Docker container.
// It's not possible to overwrite volume mapping that already exists (panic occure)
func (docker *Docker) SetVolume(hostDirectory string, containerDirectory string) {
	_, hostFound := docker.Volumes[hostDirectory]
	if hostFound {
		panic(fmt.Errorf("volume mapping is already set: '%s' --> '%s'", hostDirectory, containerDirectory))
	}
	docker.Volumes[hostDirectory] = containerDirectory
}
