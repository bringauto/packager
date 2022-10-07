package bringauto_build

import (
	"bringauto/modules/bringauto_prerequisites"
	"fmt"
)

// StartupScript represents possibility tu run script before build as part of the build shell
// eq "startup script is run in the same shell instance as a build itself"
type StartupScript struct {
	// ScriptPath path of the script to run. Default value is set to "/environment.sh"
	ScriptPath string
}

func (startupScript *StartupScript) FillDefault(*bringauto_prerequisites.Args) error {
	startupScript.ScriptPath = "/environment.sh"
	return nil
}

func (startupScript *StartupScript) FillDynamic(*bringauto_prerequisites.Args) error {
	return nil
}

func (startupScript *StartupScript) CheckPrerequisites(*bringauto_prerequisites.Args) error {
	return nil
}

func (startupScript *StartupScript) ConstructCMDLine() []string {
	command := fmt.Sprintf("test -f \"%s\" && . \"%s\"", startupScript.ScriptPath, startupScript.ScriptPath)
	return []string{command}
}
