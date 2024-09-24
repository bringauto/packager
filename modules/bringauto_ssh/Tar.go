package bringauto_ssh

import (
	"bringauto/modules/bringauto_prerequisites"
	"strings"
	"os"
)

// Tar
// Struct for creating tar archive using a tar tool
type Tar struct {
	// ArchiveName name of the archive which will be created
	ArchiveName string
	// SourceDir source directory where are files which will be added to archive (without root folder)
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

// ConstructCMDLine
// Constructs command for tar tool.
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
