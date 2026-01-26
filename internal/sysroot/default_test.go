package sysroot

import (
	"github.com/bacpack-system/packager/internal/testtools"
	"github.com/bacpack-system/packager/internal/bacpack_package"
	"github.com/bacpack-system/packager/internal/prerequisites"
	"github.com/bacpack-system/packager/internal/constants"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

const (
	gitUri = "git_uri"
	sysrootDirName = "machine-distro-1.0"
)

var defaultPlatformString bacpack_package.PlatformString
var defaultSysroot Sysroot

var builtPackage1 BuiltPackage
var builtPackage2 BuiltPackage
var builtPackage3 BuiltPackage

func TestMain(m *testing.M) {
	stringExplicit := bacpack_package.PlatformStringExplicit {
		DistroName: "distro",
		DistroRelease: "1.0",
		Machine: "machine",
	}

	defaultPlatformString = bacpack_package.PlatformString{
		Mode: bacpack_package.ModeExplicit,
		String: stringExplicit,
	}

	defaultSysroot = Sysroot {
		IsDebug: false,
		PlatformString: &defaultPlatformString,
	}
	err := prerequisites.Initialize(&defaultSysroot)
	if err != nil {
		panic(err)
	}

	err = prerequisites.Initialize(&builtPackage1, testtools.Pack1Name, sysrootDirName, gitUri, constants.EmptyGitCommitHash)
	if err != nil {
		panic(err)
	}

	err = prerequisites.Initialize(&builtPackage2, testtools.Pack2Name, sysrootDirName, gitUri, constants.EmptyGitCommitHash)
	if err != nil {
		panic(err)
	}

	err = prerequisites.Initialize(&builtPackage3, testtools.Pack3Name, sysrootDirName, gitUri, constants.EmptyGitCommitHash)
	if err != nil {
		panic(err)
	}

	err = testtools.SetupPackageFiles()
	if err != nil {
		panic(fmt.Sprintf("can't setup package files - %s", err))
	}
	code := m.Run()
	err = testtools.DeletePackageFiles()
	if err != nil {
		panic(fmt.Sprintf("can't delete package files - %s", err))
	}
	os.Exit(code)
}

func TestInitializePlatformStringNil(t *testing.T) {
	var sysroot Sysroot
	err := prerequisites.Initialize(&sysroot)
	if err == nil {
		t.Fail()
	}
}

func TestInitialize(t *testing.T) {
	sysroot := Sysroot {
		IsDebug: false,
		PlatformString: &defaultPlatformString,
	}
	err := prerequisites.Initialize(&sysroot)
	if err != nil {
		t.Fail()
	}
}

func TestGetSysrootPath(t *testing.T) {
	sysrootPath, err := defaultSysroot.GetSysrootPath()
	if err != nil {
		t.Fatalf("GetSysrootPath failed - %s", err)
	}

	workingDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("can't get working dir - %s", err)
	}
	testPath := filepath.Join(workingDir, sysrootDirectoryName, defaultSysroot.PlatformString.Serialize())

	if sysrootPath != testPath {
		t.Fail()
	}
}

func TestCreateSysrootDir(t *testing.T) {
	sysrootPath, err := defaultSysroot.GetSysrootPath()
	if err != nil {
		t.Fatalf("GetSysrootPath failed - %s", err)
	}

	err = os.RemoveAll(sysrootPath)
	if err != nil {
		t.Fatalf("can't remove sysroot dir - %s", err)
	}

	defaultSysroot.CreateSysrootDir()

	_, err = os.Stat(sysrootPath)
	if os.IsNotExist(err) {
		t.Fail()
	}
}

func TestGetDirNameInSysroot(t *testing.T) {
	dirName := defaultSysroot.GetDirNameInSysroot()

	if dirName != sysrootDirName {
		t.Fail()
	}
}

func TestCopyToSysrootOnePackage(t *testing.T) {
	err := defaultSysroot.CopyToSysroot(testtools.Pack1Name, builtPackage1)
	if err != nil {
		t.Errorf("CopyToSysroot failed - %s", err)
	}
	sysrootPath, err := defaultSysroot.GetSysrootPath()
	if err != nil {
		t.Fatalf("GetSysrootPath failed - %s", err)
	}

	pack1Path := filepath.Join(sysrootPath, testtools.Pack1FileName)
	_, err = os.ReadFile(pack1Path)
	if os.IsNotExist(err) {
		t.Fail()
	}

	err = clearSysroot()
	if err != nil {
		t.Errorf("can't delete sysroot dir - %s", err)
	}
}

func TestCopyToSysrootMultiplePackages(t *testing.T) {
	err := defaultSysroot.CopyToSysroot(testtools.Pack1Name, builtPackage1)
	if err != nil {
		t.Errorf("CopyToSysroot failed - %s", err)
	}

	err = defaultSysroot.CopyToSysroot(testtools.Pack2Name, builtPackage2)
	if err != nil {
		t.Errorf("CopyToSysroot failed - %s", err)
	}

	err = defaultSysroot.CopyToSysroot(testtools.Pack3Name, builtPackage3)
	if err != nil {
		t.Errorf("CopyToSysroot failed - %s", err)
	}
	sysrootPath, err := defaultSysroot.GetSysrootPath()
	if err != nil {
		t.Fatalf("GetSysrootPath failed - %s", err)
	}

	pack1Path := filepath.Join(sysrootPath, testtools.Pack1FileName)
	_, err = os.ReadFile(pack1Path)
	if os.IsNotExist(err) {
		t.Fail()
	}

	pack2Path := filepath.Join(sysrootPath, testtools.Pack2FileName)
	_, err = os.ReadFile(pack2Path)
	if os.IsNotExist(err) {
		t.Fail()
	}

	pack3Path := filepath.Join(sysrootPath, testtools.Pack3FileName)
	_, err = os.ReadFile(pack3Path)
	if os.IsNotExist(err) {
		t.Fail()
	}

	err = clearSysroot()
	if err != nil {
		t.Errorf("can't delete sysroot dir - %s", err)
	}
}

func TestCopyToSysrootOvewriteFiles(t *testing.T) {
	err := defaultSysroot.CopyToSysroot(testtools.Pack1Name, builtPackage1)
	if err != nil {
		t.Errorf("CopyToSysroot failed - %s", err)
	}

	err = defaultSysroot.CopyToSysroot(testtools.Pack1Name, builtPackage1)
	if err == nil {
		t.Error("ovewriting files not detected")
	}

	err = clearSysroot()
	if err != nil {
		t.Errorf("can't delete sysroot dir - %s", err)
	}
}

func TestIsPackageInSysroot(t *testing.T) {
	sysroot := Sysroot {
		IsDebug: false,
		PlatformString: &defaultPlatformString,
	}
	err := prerequisites.Initialize(&defaultSysroot)
	if err != nil {
		t.Fatalf("sysroot initialization failed - %s", err)
	}

	err = sysroot.CopyToSysroot(testtools.Pack1Name, builtPackage1)
	if err != nil {
		t.Errorf("CopyToSysroot failed - %s", err)
	}

	if !sysroot.IsPackageInSysroot(builtPackage1) {
		t.Error("IsPackageInSysroot returned false after copying package to sysroot")
	}

	if sysroot.IsPackageInSysroot(builtPackage2) {
		t.Error("IsPackageInSysroot returned true for not copied package")
	}

	if sysroot.IsPackageInSysroot(builtPackage3) {
		t.Error("IsPackageInSysroot returned true for not copied package")
	}

	err = clearSysroot()
	if err != nil {
		t.Errorf("can't delete sysroot dir - %s", err)
	}
}

func TestIsPackageInSysrootDifferentHash(t *testing.T) {
	sysroot := Sysroot {
		IsDebug: false,
		PlatformString: &defaultPlatformString,
	}
	err := prerequisites.Initialize(&defaultSysroot)
	if err != nil {
		t.Fatalf("sysroot initialization failed - %s", err)
	}

	err = sysroot.CopyToSysroot(testtools.Pack1Name, builtPackage1)
	if err != nil {
		t.Errorf("CopyToSysroot failed - %s", err)
	}

	otherPackage := builtPackage1
	otherPackage.GitCommitHash = "different_hash"

	if sysroot.IsPackageInSysroot(otherPackage) {
		t.Error("IsPackageInSysroot returned true for different package")
	}


	err = clearSysroot()
	if err != nil {
		t.Errorf("can't delete sysroot dir - %s", err)
	}
}

func clearSysroot() error {
	sysrootPath, err := defaultSysroot.GetSysrootPath()
	if err != nil {
		return err
	}
	return os.RemoveAll(filepath.Dir(sysrootPath))
}
