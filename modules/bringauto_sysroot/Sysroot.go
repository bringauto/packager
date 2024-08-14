package bringauto_sysroot

import (
	"bringauto/modules/bringauto_package"
	"bringauto/modules/bringauto_prerequisites"
	"fmt"
	"github.com/otiai10/copy"
	"os"
	"io/fs"
	"path/filepath"
	"strings"
)

const (
	SysrootDirectoryName = "install_sysroot"
)

// Sysroot represents a standard Linux sysroot with all needed libraries installed.
// Sysroot for each build type (Release, Debug) the separate sysroot is created
type Sysroot struct {
	// IsDebug - if true, it marks given sysroot as a sysroot with Debud builds
	IsDebug bool
	// PlatformString
	PlatformString *bringauto_package.PlatformString
}

func (sysroot *Sysroot) FillDefault(*bringauto_prerequisites.Args) error {
	return nil
}

func (sysroot *Sysroot) FillDynamic(*bringauto_prerequisites.Args) error {
	return nil
}

func (sysroot *Sysroot) CheckPrerequisites(args *bringauto_prerequisites.Args) error {
	if sysroot.PlatformString == nil {
		return fmt.Errorf("sysroot PlatformString cannot be nil")
	}
	return nil
}

// CopyToSysroot copy source to a sysroot
func (sysroot *Sysroot) CopyToSysroot(source string) error {
	if sysroot.anyFileAlreadyExistsInSysroot(source) {
		return fmt.Errorf("trying to ovewrite files in sysroot - sysroot consistency interrupted")
	}
	var err error
	copyOptions := copy.Options{
		OnSymlink:     onSymlink,
		PreserveOwner: true,
		PreserveTimes: true,
	}
	err = copy.Copy(source, sysroot.GetSysrootPath(), copyOptions)
	if err != nil {
		return err
	}
	return nil
}

func (sysroot *Sysroot) anyFileAlreadyExistsInSysroot(source string) bool {
	sysrootPath := sysroot.GetSysrootPath()

	err := filepath.WalkDir(source, func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() {
			filePath := strings.TrimPrefix(path, source)
			_, err := os.Stat(sysrootPath + filePath)
			if err == nil {
				return fmt.Errorf("file already exists in sysroot")
			}
		}
		
		return nil
	})

	return err != nil
}

// GetSysrootPath returns absolute path ot the sysroot
func (sysroot *Sysroot) GetSysrootPath() string {
	workingDir, err := os.Getwd()
	if err != nil {
		panic(fmt.Errorf("cannot call Getwd - %s", err))
	}

	platformString := sysroot.PlatformString.Serialize()
	sysrootDirName := platformString
	if sysroot.IsDebug {
		sysrootDirName += "_debug"
	}

	sysrootDir := filepath.Join(workingDir, SysrootDirectoryName, sysrootDirName)
	return sysrootDir
}

// CreateSysrootDir creates a Sysroot dir.
// If not succeed the panic occurred
func (sysroot *Sysroot) CreateSysrootDir() {
	var err error
	sysPath := sysroot.GetSysrootPath()
	if _, err = os.Stat(sysPath); os.IsNotExist(err) {
		err = os.MkdirAll(sysPath, 0777)
		if err != nil {
			panic(fmt.Errorf("cannot create sysroot dir: '%s'", sysPath))
		}
	}
}

func onSymlink(src string) copy.SymlinkAction {
	return copy.Shallow
}
