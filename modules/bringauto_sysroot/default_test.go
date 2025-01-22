package bringauto_sysroot

import (
	"bringauto/modules/bringauto_package"
	"bringauto/modules/bringauto_prerequisites"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

const (
	sysrootDir = "test_sysroot"

	pack1Name = "pack1"
	pack2Name = "pack2"
	pack3Name = "pack3"
	pack1FileName = pack1Name + "_file"
	pack2FileName = pack2Name + "_file"
	pack3FileName = pack3Name + "_file"
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

	err = setupPackageFiles()
	if err != nil {
		panic(fmt.Sprintf("can't setup package files - %s", err))
	}
	code := m.Run()
	err = deletePackageFiles()
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
	err := defaultSysroot.CopyToSysroot(pack1Name, pack1Name)
	if err != nil {
		t.Errorf("CopyToSysroot failed - %s", err)
	}

	pack1Path := filepath.Join(defaultSysroot.GetSysrootPath(), pack1FileName)
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
	err := defaultSysroot.CopyToSysroot(pack1Name, pack1Name)
	if err != nil {
		t.Errorf("CopyToSysroot failed - %s", err)
	}

	err = defaultSysroot.CopyToSysroot(pack2Name, pack2Name)
	if err != nil {
		t.Errorf("CopyToSysroot failed - %s", err)
	}

	err = defaultSysroot.CopyToSysroot(pack3Name, pack3Name)
	if err != nil {
		t.Errorf("CopyToSysroot failed - %s", err)
	}

	pack1Path := filepath.Join(defaultSysroot.GetSysrootPath(), pack1FileName)
	_, err = os.ReadFile(pack1Path)
	if os.IsNotExist(err) {
		t.Fail()
	}

	pack2Path := filepath.Join(defaultSysroot.GetSysrootPath(), pack2FileName)
	_, err = os.ReadFile(pack2Path)
	if os.IsNotExist(err) {
		t.Fail()
	}

	pack3Path := filepath.Join(defaultSysroot.GetSysrootPath(), pack3FileName)
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
	err := defaultSysroot.CopyToSysroot(pack1Name, pack1Name)
	if err != nil {
		t.Errorf("CopyToSysroot failed - %s", err)
	}

	err = defaultSysroot.CopyToSysroot(pack1Name, pack1Name)
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

	err = sysroot.CopyToSysroot(pack1Name, pack1Name)
	if err != nil {
		t.Errorf("CopyToSysroot failed - %s", err)
	}

	if !sysroot.IsPackageInSysroot(pack1Name) {
		t.Error("IsPackageInSysroot returned false after copying package to sysroot")
	}

	if sysroot.IsPackageInSysroot(pack2Name) {
		t.Error("IsPackageInSysroot returned true for not copied package")
	}

	if sysroot.IsPackageInSysroot(pack3Name) {
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

func setupPackageFiles() error {
	err := os.Mkdir(pack1Name, 0755)
	if err != nil {
		return fmt.Errorf("failed to create a directory - %s", err)
	}
	err = os.Mkdir(pack2Name, 0755)
	if err != nil {
		return fmt.Errorf("failed to create a directory - %s", err)
	}
	err = os.Mkdir(pack3Name, 0755)
	if err != nil {
		return fmt.Errorf("failed to create a directory - %s", err)
	}

	file1, err := os.Create(filepath.Join(pack1Name, pack1FileName))
	if err != nil {
		return fmt.Errorf("failed to create a file - %s", err)
	}
	defer file1.Close()
	file2, err := os.Create(filepath.Join(pack2Name, pack2FileName))
	if err != nil {
		return fmt.Errorf("failed to create a file - %s", err)
	}
	defer file2.Close()
	file3, err := os.Create(filepath.Join(pack3Name, pack3FileName))
	if err != nil {
		return fmt.Errorf("failed to create a file - %s", err)
	}
	defer file3.Close()

	_, err = file1.WriteString("file1 content")
	if err != nil {
		return fmt.Errorf("failed to write to file - %s", err)
	}
	_, err = file2.WriteString("file2 content")
	if err != nil {
		return fmt.Errorf("failed to write to file - %s", err)
	}
	_, err = file3.WriteString("file3 content")
	if err != nil {
		return fmt.Errorf("failed to write to file - %s", err)
	}

	return nil
}

func deletePackageFiles() error {
	err := os.RemoveAll(pack1Name)
	if err != nil {
		return err
	}
	err = os.RemoveAll(pack2Name)
	if err != nil {
		return err
	}
	err = os.RemoveAll(pack3Name)
	if err != nil {
		return err
	}
	return nil
}
