package bringauto_ssh

import (
	"bringauto/modules/bringauto_prerequisites"
	"strings"
	"os"
)

type Tar struct {
	ArchiveName string
	SourceDir string
}

func (tar *Tar) FillDefault(*bringauto_prerequisites.Args) error {
	tar.ArchiveName = ""
	tar.SourceDir = ""
	return nil
}

func (tar *Tar) FillDynamic(*bringauto_prerequisites.Args) error {
	return nil
}

func (tar *Tar) CheckPrerequisites(*bringauto_prerequisites.Args) error {
	return nil
}

func (tar *Tar) ConstructCMDLine() []string {
	var cmdLine []string
	cmdLine = append(cmdLine, "tar")
	cmdLine = append(cmdLine, "cvf")
	cmdLine = append(cmdLine, tar.SourceDir + string(os.PathSeparator) + tar.ArchiveName)
	cmdLine = append(cmdLine, "-C")
	cmdLine = append(cmdLine, tar.SourceDir)
	cmdLine = append(cmdLine, ".")

	return []string{strings.Join(cmdLine, " ")}
}