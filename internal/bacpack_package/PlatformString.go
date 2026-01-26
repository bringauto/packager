package bacpack_package

import (
	"github.com/acobaugh/osrelease"
	"github.com/bacpack-system/packager/internal/docker"
	"github.com/bacpack-system/packager/internal/log"
	"github.com/bacpack-system/packager/internal/prerequisites"
	"github.com/bacpack-system/packager/internal/ssh"
	"fmt"
	"os"
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
	// ModeAuto compute platform string automatically by lsb_release and uname
	ModeAuto = "auto"
	osReleaseFileName = "os-release"
	osReleaseFilePath = "/etc/" + osReleaseFileName
)

// PlatformString represents standard platform string
type PlatformString struct {
	// Mode of the platform string.
	Mode PlatformStringMode
	// Representation of platform-string. Constructed by one of the mode from PlatformStringMode
	String PlatformStringExplicit
}

// PlatformStringExplicit represent explicit platform string
// constructed by ModeAuto or ModeExplicit
type PlatformStringExplicit struct {
	DistroName    string
	DistroRelease string
	Machine       string
}

type platformStringInitArgs struct {
	Credentials *ssh.SSHCredentials
	Docker      *docker.Docker
}

func (pstr *PlatformString) FillDefault(args *prerequisites.Args) error {
	if prerequisites.IsEmpty(pstr) {
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

func (pstr *PlatformString) FillDynamic(args *prerequisites.Args) error {
	if !prerequisites.IsEmpty(args) {
		if pstr.Mode == ModeExplicit {
			panic(fmt.Errorf("cannot init PlatformString for args. Explicit mode is set"))
		}
		var argsStruct platformStringInitArgs
		prerequisites.GetArgs(args, &argsStruct)
		err := pstr.determinePlatformString(*argsStruct.Credentials, argsStruct.Docker)
		if err != nil {
			return err
		}
	}
	return nil
}

func (pstr *PlatformString) CheckPrerequisites(args *prerequisites.Args) error {
	if pstr.Mode == "" {
		return fmt.Errorf("please fill up PlatformStringMode")
	}
	switch pstr.Mode {
	case ModeAuto:
		return nil
	case ModeExplicit:
		break
	default:
		return fmt.Errorf("unsupported PlatformStringMode '%s'", pstr.Mode)
	}
	if !prerequisites.IsEmpty(args) {
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

// determinePlatformString
// Gets PlatformString for ModeAuto from container using credentials The container must be already running.
// If the PlatformString is in ModeExplicit the panic raise.
func (pstr *PlatformString) determinePlatformString(credentials ssh.SSHCredentials, dock *docker.Docker) error {
	if pstr.Mode == ModeExplicit {
		panic(fmt.Errorf("cannot determine PlatformString for explicit mode"))
	}

	distroName, distroRelease := getDistroIdAndReleaseFromDockerContainer(dock)
	if distroName == "" || distroRelease == "" {
		return fmt.Errorf("can't get distro name and id from os-release file")
	}

	pstr.String.DistroName = distroName
	pstr.String.DistroRelease = distroRelease
	switch pstr.Mode {
	case ModeAuto:
		pstr.String.Machine = getSystemArchitecture(credentials)
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

func runShellCommandOverSSH(credentials ssh.SSHCredentials, command string) string {
	var err error
	commandSsh := ssh.Command{
		Command: command,
	}

	commandStdOut, err := commandSsh.RunCommandOverSSH(credentials)
	if err != nil {
		panic(fmt.Errorf("cannot run command '%s' - %w", command, err))
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

func getSystemArchitecture(credentials ssh.SSHCredentials) string {
	machineUname := runShellCommandOverSSH(credentials, "uname -m")
	machine := strings.ToLower(stripNewline(machineUname))
	machine = strings.Replace(machine, "_", "-", -1)
	return machine
}

// copyOsReleaseAndGetFileLines
// Copies os-release file from docker container, opens it, parses its content and return Distro id
// and release version. 
func getDistroIdAndReleaseFromDockerContainer(dock *docker.Docker) (string, string) {
	logger := log.GetLogger()

	dockerCopy := (*docker.DockerCopy)(dock)

	err := dockerCopy.Copy(osReleaseFilePath, ".")
	if err != nil {
		logger.Error("Can't copy os-release file from docker container - %s", err)
		return "", ""
	}
	defer os.Remove(osReleaseFileName)

	osRelease, err := osrelease.ReadFile(osReleaseFileName)
	if err != nil {
		logger.Error("Can't parse os-release file - %s", err)
		return "", ""
	}

	return osRelease["ID"], osRelease["VERSION_ID"]
}
