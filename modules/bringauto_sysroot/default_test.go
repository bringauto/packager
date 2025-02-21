package bringauto_sysroot

import (
	"bringauto/modules/bringauto_testing"
	"bringauto/modules/bringauto_package"
	"bringauto/modules/bringauto_prerequisites"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

const (
	sysrootDir = "test_sysroot"
)

var defaultPlatformString bringauto_package.PlatformString
var defaultSysroot Sysroot

func TestMain(m *testing.M) {
	stringExplicit := bringauto_package.PlatformStringExplicit {
		DistroName: "distro",
		DistroRelease: "1.0",
		Machine: "machine",
	}

	defaultPlatformString = bringauto_package.PlatformString{
		Mode: bringauto_package.ModeExplicit,
		String: stringExplicit,
	}

	defaultSysroot = Sysroot {
		IsDebug: false,
		PlatformString: &defaultPlatformString,
	}
	err := bringauto_prerequisites.Initialize(&defaultSysroot)
	if err != nil {
		panic(err)
	}

	err = bringauto_testing.SetupPackageFiles()
	if err != nil {
		panic(fmt.Sprintf("can't setup package files - %s", err))
	}
	code := m.Run()
	err = bringauto_testing.DeletePackageFiles()
	if err != nil {
		panic(fmt.Sprintf("can't delete package files - %s", err))
	}
	os.Exit(code)
}

func TestInitializePlatformStringNil(t *testing.T) {
	var sysroot Sysroot
	err := bringauto_prerequisites.Initialize(&sysroot)
	if err == nil {
		t.Fail()
	}
}

func TestInitialize(t *testing.T) {
	sysroot := Sysroot {
		IsDebug: false,
		PlatformString: &defaultPlatformString,
	}
	err := bringauto_prerequisites.Initialize(&sysroot)
	if err != nil {
		t.Fail()
	}
}

func TestGetSysrootPath(t *testing.T) {
	sysrootPath := defaultSysroot.GetSysrootPath()

	workingDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("can't get workinǵ dir - %s", err)
	}
	testPath := filepath.Join(workingDir, sysrootDirectoryName, defaultSysroot.PlatformString.Serialize())

	if sysrootPath != testPath {
		t.Fail()
	}
}

func TestCreateSysrootDir(t *testing.T) {
	sysrootPath := defaultSysroot.GetSysrootPath()

	err := os.RemoveAll(sysrootPath)
	if err != nil {
		t.Fatalf("can't remove sysroot dir - %s", err)
	}

	defaultSysroot.CreateSysrootDir()

	_, err = os.Stat(sysrootPath)
	if os.IsNotExist(err) {
		t.Fail()
	}
}

func TestCopyToSysrootOnePackage(t *testing.T) {
	err := defaultSysroot.CopyToSysroot(bringauto_testing.Pack1Name, bringauto_testing.Pack1Name)
	if err != nil {
		t.Errorf("CopyToSysroot failed - %s", err)
	}

	pack1Path := filepath.Join(defaultSysroot.GetSysrootPath(), bringauto_testing.Pack1FileName)
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
	err := defaultSysroot.CopyToSysroot(bringauto_testing.Pack1Name, bringauto_testing.Pack1Name)
	if err != nil {
		t.Errorf("CopyToSysroot failed - %s", err)
	}

	err = defaultSysroot.CopyToSysroot(bringauto_testing.Pack2Name, bringauto_testing.Pack2Name)
	if err != nil {
		t.Errorf("CopyToSysroot failed - %s", err)
	}

	err = defaultSysroot.CopyToSysroot(bringauto_testing.Pack3Name, bringauto_testing.Pack3Name)
	if err != nil {
		t.Errorf("CopyToSysroot failed - %s", err)
	}

	pack1Path := filepath.Join(defaultSysroot.GetSysrootPath(), bringauto_testing.Pack1FileName)
	_, err = os.ReadFile(pack1Path)
	if os.IsNotExist(err) {
		t.Fail()
	}

	pack2Path := filepath.Join(defaultSysroot.GetSysrootPath(), bringauto_testing.Pack2FileName)
	_, err = os.ReadFile(pack2Path)
	if os.IsNotExist(err) {
		t.Fail()
	}

	pack3Path := filepath.Join(defaultSysroot.GetSysrootPath(), bringauto_testing.Pack3FileName)
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
	err := defaultSysroot.CopyToSysroot(bringauto_testing.Pack1Name, bringauto_testing.Pack1Name)
	if err != nil {
		t.Errorf("CopyToSysroot failed - %s", err)
	}

	err = defaultSysroot.CopyToSysroot(bringauto_testing.Pack1Name, bringauto_testing.Pack1Name)
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
	err := bringauto_prerequisites.Initialize(&defaultSysroot)
	if err != nil {
		t.Fatalf("sysroot initialization failed - %s", err)
	}

	err = sysroot.CopyToSysroot(bringauto_testing.Pack1Name, bringauto_testing.Pack1Name)
	if err != nil {
		t.Errorf("CopyToSysroot failed - %s", err)
	}

	if !sysroot.IsPackageInSysroot(bringauto_testing.Pack1Name) {
		t.Error("IsPackageInSysroot returned false after copying package to sysroot")
	}

	if sysroot.IsPackageInSysroot(bringauto_testing.Pack2Name) {
		t.Error("IsPackageInSysroot returned true for not copied package")
	}

	if sysroot.IsPackageInSysroot(bringauto_testing.Pack3Name) {
		t.Error("IsPackageInSysroot returned true for not copied package")
	}

	err = clearSysroot()
	if err != nil {
		t.Errorf("can't delete sysroot dir - %s", err)
	}
}

func clearSysroot() error {
	sysrootPath := defaultSysroot.GetSysrootPath()
	return os.RemoveAll(filepath.Dir(sysrootPath))
}
