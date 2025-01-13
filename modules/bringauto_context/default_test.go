package bringauto_context

import (
	"bringauto/modules/bringauto_const"
	"bringauto/modules/bringauto_package"
	"path/filepath"
	"slices"
	"testing"
)

const (
	TestDataDirName = "test_data"
	Valid1DirName = "valid1"
	Valid2DirName = "valid2"
	Invalid1DirName = "invalid1"
	Valid1DirPath = TestDataDirName + "/" + Valid1DirName
	Valid2DirPath = TestDataDirName + "/" + Valid2DirName
	Invalid1DirPath = TestDataDirName + "/" + Invalid1DirName

	Pack1Name = "pack1"
	Pack2Name = "pack2"
	Pack3Name = "pack3"
	Pack4Name = "pack4"
	Pack5Name = "pack5"
	Image1Name = "image1"
	Image2Name = "image2"
	DockerfileName = "Dockerfile"
)

var defaultPlatformString bringauto_package.PlatformString

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
	m.Run()
}

func TestGetAllPackagesJsonDefPaths(t *testing.T) {
	context := ContextManager {
		ContextPath: Valid1DirPath,
	}

	jsonPaths, err := context.GetAllPackagesJsonDefPaths()
	if err != nil {
		t.Fatalf("GetAllPackagesJsonDefPaths failed - %s", err)
	}

	pack1Paths, ok1 := jsonPaths[Pack1Name]
	pack2Paths, ok2 := jsonPaths[Pack2Name]
	pack3Paths, ok3 := jsonPaths[Pack3Name]

	if !ok1 || !ok2 || !ok3 {
		t.Fatalf("some package was not returned")
	}

	commonPath := filepath.Join(Valid1DirPath, bringauto_const.PackageDirName)
	pack1Json1Path := filepath.Join(commonPath, Pack1Name, Pack1Name + "_release.json")
	pack1Json2Path := filepath.Join(commonPath, Pack1Name, Pack1Name + "_debug.json")
	pack2Json1Path := filepath.Join(commonPath, Pack2Name, Pack2Name + ".json")
	pack3Json1Path := filepath.Join(commonPath, Pack3Name, Pack3Name + "_debug.json")

	if (!slices.Contains(pack1Paths, pack1Json1Path) ||
		!slices.Contains(pack1Paths, pack1Json2Path) ||
		!slices.Contains(pack2Paths, pack2Json1Path) ||
		!slices.Contains(pack3Paths, pack3Json1Path)) {
		t.Fatalf("wrong returned paths")
	}
}

func TestGetPackageJsonDefPaths(t *testing.T) {
	context := ContextManager {
		ContextPath: Valid1DirPath,
	}

	pack1Paths, err := context.GetPackageJsonDefPaths(Pack1Name)
	if err != nil {
		t.Fatalf("GetPackageJsonDefPaths failed - %s", err)
	}

	commonPath := filepath.Join(Valid1DirPath, bringauto_const.PackageDirName)
	pack1Json1Path := filepath.Join(commonPath, Pack1Name, Pack1Name + "_release.json")
	pack1Json2Path := filepath.Join(commonPath, Pack1Name, Pack1Name + "_debug.json")

	if (!slices.Contains(pack1Paths, pack1Json1Path) ||
		!slices.Contains(pack1Paths, pack1Json2Path)) {
		t.Fatalf("wrong returned paths")
	}
}

func TestGetAllPackagesConfigs(t *testing.T) {
	context := ContextManager {
		ContextPath: Valid1DirPath,
	}

	configs, err := context.GetAllPackagesConfigs(&defaultPlatformString)
	if err != nil {
		t.Fatalf("GetAllPackagesConfigs failed - %s", err)
	}

	if len(configs) != 4 {
		t.Fatal("wrong number of returned configs")
	}

	// Checking some properties
	for _, config := range configs {
		if config.Package.Name == Pack1Name && config.Package.IsDebug {
			if config.Package.VersionTag != "v1.0.0" {
				t.Error("wrong config content")
			}
		} else if config.Package.Name == Pack1Name {
			if (config.DockerMatrix.ImageNames[0] != Image1Name ||
				len(config.DependsOn) != 0) {
				t.Error("wrong config content")
			}
		} else if config.Package.Name == Pack2Name {
			if config.Build.CMake.Defines["BRINGAUTO_INSTALL"] != "ON" {
				t.Error("wrong config content")
			}
		} else if config.Package.Name == Pack3Name {
			if (config.DependsOn[0] != "pack1" ||
				config.DependsOn[1] != "pack2") {
				t.Error("wrong config content")
			}
		} else {
			t.Error("returned config for unknown package")
		}
	}
}

func TestGetAllImagesDockerfilePaths(t *testing.T) {
	context := ContextManager {
		ContextPath: Valid1DirPath,
	}

	paths, err := context.GetAllImagesDockerfilePaths()
	if err != nil {
		t.Fatalf("GetAllImagesDockerfilePaths failed - %s", err)
	}

	image1Paths, ok1 := paths[Image1Name]
	image2Paths, ok2 := paths[Image2Name]

	if !ok1 || !ok2 {
		t.Fatalf("some image was not returned")
	}

	commonPath := filepath.Join(Valid1DirPath, bringauto_const.DockerDirName)
	image1Path := filepath.Join(commonPath, Image1Name, DockerfileName)
	image2Path := filepath.Join(commonPath, Image2Name, DockerfileName)

	if (!slices.Contains(image1Paths, image1Path) ||
		!slices.Contains(image2Paths, image2Path)) {
		t.Fatalf("wrong returned paths")
	}
}

func TestGetPackageWithDepsJsonDefPaths(t *testing.T) {
	context := ContextManager {
		ContextPath: Valid2DirPath,
	}

	paths, err := context.GetPackageWithDepsJsonDefPaths(Pack3Name)
	if err != nil {
		t.Fatalf("GetPackageWithDepsJsonDefPaths failed - %s", err)
	}

	commonPath := filepath.Join(Valid2DirPath, bringauto_const.PackageDirName)
	pack1Path := filepath.Join(commonPath, Pack1Name, Pack1Name + ".json")
	pack2Path := filepath.Join(commonPath, Pack2Name, Pack2Name + ".json")
	pack3Path := filepath.Join(commonPath, Pack3Name, Pack3Name + ".json")
	pack4Path := filepath.Join(commonPath, Pack4Name, Pack4Name + ".json")

	if (len(paths) != 4 ||
		!slices.Contains(paths, pack1Path) ||
		!slices.Contains(paths, pack2Path) ||
		!slices.Contains(paths, pack3Path) ||
	 	!slices.Contains(paths, pack4Path)) {
		t.Fatalf("wrong returned paths")
	}
}

func TestGetPackageWithDepsJsonDefPathsNoDepWithBuildType(t *testing.T) {
	context := ContextManager {
		ContextPath: Invalid1DirPath,
	}

	_, err := context.GetPackageWithDepsJsonDefPaths(Pack2Name)
	if err == nil {
		t.Error("GetPackageWithDepsJsonDefPaths didn't returned error")
	}
}

func TestGetDepsOnJsonDefPaths(t *testing.T) {
	context := ContextManager {
		ContextPath: Valid2DirPath,
	}

	paths, err := context.GetDepsOnJsonDefPaths(Pack1Name, false)
	if err != nil {
		t.Fatalf("GetDepsOnJsonDefPaths failed - %s", err)
	}

	commonPath := filepath.Join(Valid2DirPath, bringauto_const.PackageDirName)
	pack4Path := filepath.Join(commonPath, Pack4Name, Pack4Name + ".json")

	if (len(paths) != 1 ||
		!slices.Contains(paths, pack4Path)) {
		t.Fatalf("wrong returned paths")
	}
}

func TestGetDepsOnJsonDefPathsRecursively(t *testing.T) {
	context := ContextManager {
		ContextPath: Valid2DirPath,
	}

	paths, err := context.GetDepsOnJsonDefPaths(Pack1Name, true)
	if err != nil {
		t.Fatalf("GetDepsOnJsonDefPaths failed - %s", err)
	}

	commonPath := filepath.Join(Valid2DirPath, bringauto_const.PackageDirName)
	pack3Path := filepath.Join(commonPath, Pack3Name, Pack3Name + ".json")
	pack4Path := filepath.Join(commonPath, Pack4Name, Pack4Name + ".json")
	pack5Path := filepath.Join(commonPath, Pack5Name, Pack5Name + ".json")

	if (len(paths) != 3 ||
		!slices.Contains(paths, pack3Path) ||
		!slices.Contains(paths, pack4Path) ||
		!slices.Contains(paths, pack5Path)) {
		t.Fatalf("wrong returned paths")
	}
}
