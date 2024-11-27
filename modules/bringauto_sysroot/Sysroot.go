package bringauto_sysroot

import (
	"bringauto/modules/bringauto_log"
	"bringauto/modules/bringauto_package"
	"bringauto/modules/bringauto_prerequisites"
	"fmt"
	"github.com/otiai10/copy"
	"os"
	"io"
	"io/fs"
	"path/filepath"
	"strings"
)

const (
	sysrootDirectoryName = "install_sysroot"
	// Constant for number of problematic files which will be printed when trying to overwrite files
	// in sysroot
	listFilesCount = 10
)

// Sysroot represents a standard Linux sysroot with all needed libraries installed.
// Sysroot for each build type (Release, Debug) the separate sysroot is created
type Sysroot struct {
	// IsDebug - if true, it marks given sysroot as a sysroot with Debud builds
	IsDebug bool
	// PlatformString
	PlatformString *bringauto_package.PlatformString
	builtPackages BuiltPackages
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
func (sysroot *Sysroot) CopyToSysroot(source string, packageName string) error {
	err := sysroot.checkForOverwritingFiles(source)
	if err != nil {
		return err
	}
	copyOptions := copy.Options{
		OnSymlink:     onSymlink,
		PreserveOwner: true,
		PreserveTimes: true,
	}
	err = copy.Copy(source, sysroot.GetSysrootPath(), copyOptions)
	if err != nil {
		return err
	}
	err = sysroot.builtPackages.AddToBuiltPackages(packageName)
	if err != nil {
		return err
	}
	return nil
}

// checkForOverwritingFiles
// Checks if in dirPath directory are not files which are also in sysroot directory. If there are
// some, then prints Error with listing problematic files and returns non nil error. Else returns
// nil error without printing anything.
func (sysroot *Sysroot) checkForOverwritingFiles(dirPath string) error {
	filesToCopy := getExistingFilesInDir(dirPath)
	filesInSysrootMap := make(map[string]struct{})
	for _, file := range getExistingFilesInDir(sysroot.GetSysrootPath()) {
		filesInSysrootMap[file] = struct{}{}
	}
	var intersection []string
	for _, fileToCopy := range filesToCopy {
		_, exists := filesInSysrootMap[fileToCopy]
		if exists {
			intersection = append(intersection, fileToCopy)
		}
	}
	if len(intersection) > 0 {
		sysroot.printOverwriteFilesError(intersection, listFilesCount)
		return fmt.Errorf("trying to overwrite files in sysroot")
	}
	return nil
}

// printOverwriteFilesError
// Prints error for overwriting files in sysroot. Lists first n files in problematic_files.
func (sysroot *Sysroot) printOverwriteFilesError(problematicFiles []string, n int) {
	logger := bringauto_log.GetLogger()
	logger.Error("Trying to overwrite files in sysroot - sysroot consistency interrupted.")
	logger.Error("Listing first %d problematic files:", n)
	for i, filePath := range problematicFiles {
		logger.ErrorIndent(sysrootDirectoryName + "/" + sysroot.PlatformString.Serialize() + filePath)
		if i == n - 1 {
			break
		}
	}
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

	sysrootDir := filepath.Join(workingDir, sysrootDirectoryName, sysrootDirName)
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

// IsSysrootDirectoryEmpty
// Returns true if specified dir do not exists or exists but is empty, otherwise returns false
func (sysroot *Sysroot) IsSysrootDirectoryEmpty() bool {
	f, err := os.Open(sysroot.GetSysrootPath())
	if err != nil { // The directory do not exists
		return true
	}
	defer f.Close()

	_, err = f.Readdirnames(1)

	if err == io.EOF { // The directory exists, but is empty
		return true
	} else if err != nil {
		bringauto_log.GetLogger().Warn("Sysroot directory is not readable: %s", err)
	}

	return false
}

func onSymlink(src string) copy.SymlinkAction {
	return copy.Shallow
}

// getExistingFilesInDir
// Returns array of string which contains all file paths existing in dirPath directory. The
// returned paths are without dirPath prefix.
func getExistingFilesInDir(dirPath string) []string {
	var existingFiles []string

	filepath.WalkDir(dirPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			filePath := strings.TrimPrefix(path, dirPath)
			_, err := os.Stat(path)
			if err == nil {
				existingFiles = append(existingFiles, filePath)
			}
		}

		return nil
	})

	return existingFiles
}
