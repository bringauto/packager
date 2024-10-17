package bringauto_ssh

import (
	"bringauto/modules/bringauto_prerequisites"
	"fmt"
	"path/filepath"
	"strings"
)

// Tar
// Struct for creating tar archive using a tar tool
type Tar struct {
	// ArchiveName name of the archive which will be created
	ArchiveName string
	// SourceDir source directory where are files which will be added to archive (without root folder)
	SourceDir string
}

type tarInitArgs struct {
	ArchiveName string
	SourceDir string
}

func (tar *Tar) FillDefault(*bringauto_prerequisites.Args) error {
	tar.ArchiveName = ""
	tar.SourceDir = ""
	return nil
}

func (tar *Tar) FillDynamic(args *bringauto_prerequisites.Args) error {
	var argsStruct tarInitArgs
	bringauto_prerequisites.GetArgs(args, &argsStruct)
	tar.ArchiveName = argsStruct.ArchiveName
	tar.SourceDir = argsStruct.SourceDir
	return nil
}

func (tar *Tar) CheckPrerequisites(*bringauto_prerequisites.Args) error {
	if tar.ArchiveName == "" {
		return fmt.Errorf("empty archive name")
	}
	if tar.SourceDir == "" {
		return fmt.Errorf("empty source directory name")
	}
	return nil
}

// ConstructCMDLine
// Constructs command for tar tool.
func (tar *Tar) ConstructCMDLine() []string {
	var cmdBuilder strings.Builder
	cmdBuilder.WriteString("tar cvf ")
	cmdBuilder.WriteString(filepath.Join(tar.SourceDir, tar.ArchiveName))
	cmdBuilder.WriteString(" -C ")
	cmdBuilder.WriteString(tar.SourceDir)
	cmdBuilder.WriteString(" .")

	return []string{cmdBuilder.String()}
}
