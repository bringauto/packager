package bringauto_package

import (
	"bringauto/modules/bringauto_prerequisites"
	"compress/flate"
	"fmt"
	"github.com/mholt/archiver/v3"
	"os"
	"path"
	"regexp"
	"strings"
)

const (
	defaultPackageNameConst = "generic-package"
	defaultVersionTagConst  = "v0.0.0"
)

// Package enables us to easily create a package
type Package struct {
	// Base name of the package
	Name string
	// Standard VersionTag
	VersionTag string
	// Standard platform string
	PlatformString PlatformString
	// Mark package as library if true
	IsLibrary bool
	// Mark package as development library if true
	IsDevLib bool
	// Mark package as debug build if true
	IsDebug bool
}

func (packg *Package) FillDefault(*bringauto_prerequisites.Args) error {
	*packg = Package{
		Name:       defaultPackageNameConst,
		VersionTag: defaultVersionTagConst,
		IsDebug:    false,
		IsDevLib:   true,
		IsLibrary:  true,
		PlatformString: PlatformString{
			Mode: ModeAuto,
		},
	}
	return nil
}

func (packg *Package) FillDynamic(args *bringauto_prerequisites.Args) error {
	err := bringauto_prerequisites.Initialize(&packg.PlatformString, args)
	return err
}

func (packg *Package) CheckPrerequisites(*bringauto_prerequisites.Args) error {
	if !packg.IsLibrary && packg.IsDevLib {
		return fmt.Errorf("IsDevLib is true but IsLibrary is false")
	}

	versionTagRegex, _ := regexp.CompilePOSIX("^v[0-9]+\\.[0-9]+\\.[0-9]+$")
	if !versionTagRegex.MatchString(packg.VersionTag) {
		return fmt.Errorf("VersionTag %s is not valid version tag", packg.VersionTag)
	}
	if packg.Name == "" {
		return fmt.Errorf("package name cannot be empty")
	}
	return nil
}

// CreatePackage creates a package from sourceDir directory
//   - construct package name
//   - zip all files into archive with name <package_name>.zip
//   - copy the zip archive to the outputDir
func (packg *Package) CreatePackage(sourceDir string, outputDir string) error {
	var err error
	if _, err = os.Stat(sourceDir); os.IsNotExist(err) {
		return err
	}
	err = os.MkdirAll(outputDir, os.ModeDir)
	if err != nil {
		return err
	}

	packageName := packg.CreatePackageName() + ".zip"

	err = createZIPArchive(sourceDir, outputDir+"/"+packageName)
	if err != nil {
		return fmt.Errorf("cannot create zip archive")
	}

	return nil
}

// CreatePackageName
// construct only a package name.
// Function returns nonempty string.
func (packg *Package) CreatePackageName() string {
	var packageName []string
	if packg.IsLibrary {
		packageName = append([]string{"lib"}, packageName...)
	}
	packageName = append(packageName, packg.Name)
	if packg.IsDebug {
		packageName = append(packageName, "d")
	}
	if packg.IsDevLib {
		packageName = append(packageName, "-dev")
	}
	packageName = append(packageName, "_")
	packageName = append(packageName, packg.VersionTag)
	packageName = append(packageName, "_")
	packageName = append(packageName, packg.PlatformString.Serialize())
	return strings.Join(packageName, "")
}

// Returns short package name without versions
func (packg *Package) GetShortPackageName() string {
	var packageName []string

	packageName = append(packageName, packg.Name)
	if packg.IsDebug {
		packageName = append(packageName, "d")
	}
	if packg.IsDevLib {
		packageName = append(packageName, "-dev")
	}

	return strings.Join(packageName, "")
}

func createZIPArchive(sourceDir string, archivePath string) error {
	var files []string
	var err error

	dirEntryList, err := os.ReadDir(sourceDir)
	for _, dirEntry := range dirEntryList {
		files = append(files, path.Join(sourceDir, dirEntry.Name()))
	}

	zipArchive := archiver.Zip{
		MkdirAll:             true,
		FileMethod:           archiver.Deflate,
		OverwriteExisting:    true,
		SelectiveCompression: true,
		CompressionLevel:     flate.BestSpeed,
	}

	err = zipArchive.Archive(files, archivePath)
	if err != nil {
		return fmt.Errorf("cannot create archive: %s", err)
	}

	return nil
}
