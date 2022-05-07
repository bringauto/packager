package bringauto_package

import (
	"bringauto/modules/bringauto_docker"
	"bringauto/modules/bringauto_prerequisites"
	"bringauto/modules/bringauto_ssh"
	"fmt"
	"regexp"
	"strings"
)

// PlatformStringMode is a fill-up mode of the platform-string.
// Basically there are two main modes:
// - Auto: determine platform string by a heuristic algorithm
// - Explicit: the platform string is explicitly provided by user itself.
type PlatformStringMode string

const (
	// ModeExplicit denotes that the plafrom string is filled-up by user.
	// Function determinePlatformString should not be used with ModeExplicit.
	ModeExplicit PlatformStringMode = "explicit"
	// ModeAnyMachine is a same as ModeAuto except Machine that is se to "any"
	ModeAnyMachine = "any_machine"
	// ModeAuto compute platform string automatically by lsb_release and uname
	ModeAuto = "auto"
)

// PlatformString represents standard platform string
type PlatformString struct {
	// Mode of the platform string.
	Mode PlatformStringMode
	// Representation of platform-string. Constructed by one of the mode from PlatformStringMode
	String PlatformStringExplicit
}

// PlatformStringExplicit represent explicit platform string
// constructed by ModeAuto, ModeAnyMachine or ModeExplicit
type PlatformStringExplicit struct {
	DistroName    string
	DistroRelease string
	Machine       string
}

type platformStringInitArgs struct {
	Credentials *bringauto_ssh.SSHCredentials
	Docker      *bringauto_docker.Docker
}

func (pstr *PlatformString) FillDefault(args *bringauto_prerequisites.Args) error {
	if bringauto_prerequisites.IsEmpty(pstr) {
		*pstr = PlatformString{
			Mode: ModeExplicit,
			String: PlatformStringExplicit{
				DistroName:    "unknown",
				DistroRelease: "unknown",
				Machine:       "unknown",
			},
		}
	}
	return nil
}

func (pstr *PlatformString) FillDynamic(args *bringauto_prerequisites.Args) error {
	if !bringauto_prerequisites.IsEmpty(args) {
		if pstr.Mode == ModeExplicit {
			panic(fmt.Errorf("cannot init PlatformString for args. Explicit mode is set"))
		}
		var argsStruct platformStringInitArgs
		bringauto_prerequisites.GetArgs(args, &argsStruct)
		err := pstr.determinePlatformString(*argsStruct.Credentials, argsStruct.Docker)
		if err != nil {
			return err
		}
	}
	return nil
}

func (pstr *PlatformString) CheckPrerequisites(args *bringauto_prerequisites.Args) error {
	if pstr.Mode == "" {
		return fmt.Errorf("please fill up PlatformStringMode")
	}
	switch pstr.Mode {
	case ModeAuto:
	case ModeAnyMachine:
	case ModeExplicit:
		return nil
	default:
		return fmt.Errorf("unsupported PlatformStringMode '%s'", pstr.Mode)
	}
	if !bringauto_prerequisites.IsEmpty(args) {
		errorMsg := ""
		if pstr.String.DistroName == "" {
			errorMsg += fmt.Sprintf("please fill up DistroName for a PlatformString '%s'\n", pstr.Serialize())
		}
		if pstr.String.DistroRelease == "" {
			errorMsg += fmt.Sprintf("please fill up DistroRelease for a PlatformString '%s'\n", pstr.Serialize())
		}
		if pstr.String.Machine == "" {
			errorMsg += fmt.Sprintf("please fill up Machine for a PlatformString '%s'\n", pstr.Serialize())
		}
		if errorMsg != "" {
			return fmt.Errorf(errorMsg)
		}
	}
	return nil
}

// determinePlatformString tries to compute
// platform string for ModeAuto and ModeAnyMachine.
// If the PlatformString is in ModeExplicit the panic raise.
func (pstr *PlatformString) determinePlatformString(credentials bringauto_ssh.SSHCredentials, docker *bringauto_docker.Docker) error {
	if pstr.Mode == ModeExplicit {
		panic(fmt.Errorf("cannot determine PlatformString for explicit mode"))
	}

	dockerRun := (*bringauto_docker.DockerRun)(docker)
	err := dockerRun.Run()
	if err != nil {
		return err
	}
	defer func() {
		dockerStop := (*bringauto_docker.DockerStop)(docker)
		dockerRm := (*bringauto_docker.DockerRm)(docker)
		dockerStop.Stop()
		dockerRm.RemoveContainer()
	}()

	pstr.String.DistroName = getDistributionName(credentials)
	pstr.String.DistroRelease = getReleaseVersion(credentials)
	switch pstr.Mode {
	case ModeAuto:
		pstr.String.Machine = getSystemArchitecture(credentials)
	case ModeAnyMachine:
		pstr.String.Machine = "any"
	default:
		panic(fmt.Errorf("unsupported PlatformStringMode"))
	}

	return nil
}

// Serialize serializes PlatformString into human-readable string
// that can be used for package naming.
func (pstr *PlatformString) Serialize() string {
	if pstr.String.DistroName == "" && pstr.String.Machine == "" && pstr.String.DistroRelease == "" {
		panic("Sorry, invalid platform string")
	}
	return pstr.String.Machine + "-" + pstr.String.DistroName + "-" + pstr.String.DistroRelease
}

func runShellCommandOverSSH(credentials bringauto_ssh.SSHCredentials, command string) string {
	var err error
	commandSsh := bringauto_ssh.Command{
		Command: command,
	}

	var commandStdOut string
	commandStdOut, err = commandSsh.RunCommandOverSSH(credentials)
	if err != nil {
		panic(fmt.Errorf("cannot run command '%s', error: %s", command, err))
	}
	return commandStdOut
}

func stripNewline(str string) string {
	regexp, regexpErr := regexp.CompilePOSIX("^([^\n\r]+)")
	if regexpErr != nil {
		panic(fmt.Errorf("invalid regexp for strip newline"))
	}
	return regexp.FindString(str)
}

func getDistributionName(credentials bringauto_ssh.SSHCredentials) string {
	distroNameLSBRelease := runShellCommandOverSSH(credentials, "lsb_release -is")
	distroName := strings.ToLower(stripNewline(distroNameLSBRelease))
	return distroName
}
func getReleaseVersion(credentials bringauto_ssh.SSHCredentials) string {
	releaseVersionLSBRelease := runShellCommandOverSSH(credentials, "lsb_release -rs")
	releaseVersion := strings.ToLower(stripNewline(releaseVersionLSBRelease))
	return releaseVersion
}

func getSystemArchitecture(credentials bringauto_ssh.SSHCredentials) string {
	machineUname := runShellCommandOverSSH(credentials, "uname -m")
	machine := strings.ToLower(stripNewline(machineUname))
	machine = strings.Replace(machine, "_", "-", -1)
	return machine
}
