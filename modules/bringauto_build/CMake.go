package bringauto_build

import (
	"bringauto/modules/bringauto_prerequisites"
	"fmt"
	"os"
	"path"
	"regexp"
	"strings"
)

type CMake struct {
	Defines      map[string]string
	CMakeListDir string
	SourceDir    string `json:"-"`
}

func (cmake *CMake) FillDefault(*bringauto_prerequisites.Args) error {
	*cmake = CMake{
		Defines: map[string]string{
			"CMAKE_BUILD_TYPE": "Debug",
		},
		CMakeListDir: "." + string(os.PathSeparator),
	}
	return nil
}

func (cmake *CMake) FillDynamic(*bringauto_prerequisites.Args) error {
	return nil
}

func (cmake *CMake) CheckPrerequisites(*bringauto_prerequisites.Args) error {
	return nil
}

func (cmake *CMake) ConstructCMDLine() []string {
	var cmdLine []string
	cmdLine = append(cmdLine, "cmake")
	for key, value := range cmake.Defines {
		if !validateVariableName(key) {
			panic(fmt.Errorf("invalid CMake variable: %s", key))
		}
		valuePair := "-D" + key + "=" + escapeVariableValue(value)
		cmdLine = append(cmdLine, valuePair)
	}
	if cmake.SourceDir == "" {
		panic(fmt.Errorf("cmake source source directory does not exist"))
	}
	cmdLine = append(cmdLine, path.Join(cmake.SourceDir, cmake.CMakeListDir))
	return []string{strings.Join(cmdLine, " ")}
}

func (cmake *CMake) SetDefine(key string, value string) {
	_, prefixpathFound := cmake.Defines[key]
	if prefixpathFound {
		panic(fmt.Errorf("cmake define - do not specify %s", value))
	}
	cmake.Defines[key] = value
}

func validateVariableName(varName string) bool {
	regexp, regexpErr := regexp.CompilePOSIX("^[0-9a-zA-Z_]+$")
	if regexpErr != nil {
		panic(fmt.Errorf("invalid regexp for CMake variable validation"))
		return false
	}
	return regexp.MatchString(varName)
}

func escapeVariableValue(varValue string) string {
	return "\"" + varValue + "\""
}
