package bringauto_build

import (
	"bringauto/modules/bringauto_prerequisites"
	"strconv"
	"strings"
)

// GNUMake cmd line interface for standard GNU Make utility
type GNUMake struct {
	// Map of the cmd line arguments where map_key represents cmd option
	// and map_value value of the option
	CMDLineVars map[string]string

	// number of jobs passed to '-j'. Default: 10
	jobsCount int
}

func (make *GNUMake) FillDefault(*bringauto_prerequisites.Args) error {
	*make = GNUMake{
		jobsCount: 10,
	}
	return nil
}

func (make *GNUMake) FillDynamic(*bringauto_prerequisites.Args) error {
	return nil
}

func (make *GNUMake) CheckPrerequisites(*bringauto_prerequisites.Args) error {
	return nil
}

func (make *GNUMake) ConstructCMDLine() []string {
	cmdBuild := []string{"make", "-j", strconv.Itoa(make.jobsCount)}
	cmdInstall := []string{"make", "install"}
	return []string{
		strings.Join(cmdBuild, " "),
		strings.Join(cmdInstall, " "),
	}
}
