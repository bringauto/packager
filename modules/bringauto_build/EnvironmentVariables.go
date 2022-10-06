package bringauto_build

import (
	"bringauto/modules/bringauto_prerequisites"
	"fmt"
	"regexp"
)

// EnvironmentVariables export env variables in the same shell instance as the build is run
type EnvironmentVariables struct {
	Env map[string]string
}

func (vars *EnvironmentVariables) FillDefault(*bringauto_prerequisites.Args) error {
	if vars.Env == nil {
		vars.Env = map[string]string{}
	}
	return nil
}

func (vars *EnvironmentVariables) FillDynamic(*bringauto_prerequisites.Args) error {
	return nil
}

func (vars *EnvironmentVariables) CheckPrerequisites(*bringauto_prerequisites.Args) error {
	return nil
}

func (vars *EnvironmentVariables) ConstructCMDLine() []string {
	var commands []string
	for key, value := range vars.Env {
		validateKey(key)
		commands = append(commands, "export "+key+"="+escapeValue(value))
	}
	return commands
}

func validateKey(key string) {
	regexp, regexpErr := regexp.CompilePOSIX("^([0-9a-zA-Z]+)")
	if regexpErr != nil {
		panic(regexpErr)
	}
	if !regexp.MatchString(key) {
		panic(fmt.Errorf("key %s is not valid", key))
	}
}

func escapeValue(value string) string {
	return "\"" + value + "\""
}
