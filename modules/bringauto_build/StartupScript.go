package bringauto_build

import (
	"bringauto/modules/bringauto_prerequisites"
	"fmt"
)

type StartupScript struct {
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
