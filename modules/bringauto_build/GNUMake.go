package bringauto_build

import (
	"bringauto/modules/bringauto_prerequisites"
	"strconv"
	"strings"
)

type GNUMake struct {
	CMDLineVars map[string]string
	jobsCount   int
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
