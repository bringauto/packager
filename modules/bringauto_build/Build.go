package bringauto_build

import (
	"bringauto/modules/bringauto_docker"
	"bringauto/modules/bringauto_git"
	"bringauto/modules/bringauto_log"
	"bringauto/modules/bringauto_const"
	"bringauto/modules/bringauto_package"
	"bringauto/modules/bringauto_prerequisites"
	"bringauto/modules/bringauto_ssh"
	"bringauto/modules/bringauto_sysroot"
	"bringauto/modules/bringauto_process"
	"fmt"
	"os"
	"path/filepath"
)

type Build struct {
	Env            *EnvironmentVariables
	Docker         *bringauto_docker.Docker
	Git            *bringauto_git.Git
	CMake          *CMake
	GNUMake        *GNUMake
	SSHCredentials *bringauto_ssh.SSHCredentials
	Package        *bringauto_package.Package
	sysroot        *bringauto_sysroot.Sysroot
}

// FillDefault
// It fills up defaults for all members in the Build structure.
func (build *Build) FillDefault(*bringauto_prerequisites.Args) error {
	if build.Git == nil {
		build.Git = bringauto_prerequisites.CreateAndInitialize[bringauto_git.Git]()
	}
	if build.Docker == nil {
		build.Docker = bringauto_prerequisites.CreateAndInitialize[bringauto_docker.Docker]()
	}
	if build.SSHCredentials == nil {
		build.SSHCredentials = bringauto_prerequisites.CreateAndInitialize[bringauto_ssh.SSHCredentials]()
	}
	if build.CMake == nil {
		build.CMake = bringauto_prerequisites.CreateAndInitialize[CMake]()
	}
	if build.GNUMake == nil {
		build.GNUMake = bringauto_prerequisites.CreateAndInitialize[GNUMake]()
	}
	if build.Env == nil {
		build.Env = bringauto_prerequisites.CreateAndInitialize[EnvironmentVariables]()
	}

	if build.Package == nil {
		build.Package = bringauto_prerequisites.CreateAndInitialize[bringauto_package.Package]()
	}

	return nil
}

func (build *Build) FillDynamic(*bringauto_prerequisites.Args) error {
	return nil
}

func (build *Build) CheckPrerequisites(*bringauto_prerequisites.Args) error {
	copyDir := build.GetLocalInstallDirPath()
	if _, err := os.Stat(copyDir); !os.IsNotExist(err) {
		return fmt.Errorf("package directory exist. Please delete it: %s", copyDir)
	}

	return nil
}

// RunBuild
// s
func (build *Build) RunBuild() error {
	var err error

	err = build.CheckPrerequisites(nil)
	if err != nil {
		return err
	}

	build.Git.ClonePath = dockerGitCloneDirConst
	build.CMake.SourceDir = dockerGitCloneDirConst

	_, found := build.CMake.Defines["CMAKE_INSTALL_PREFIX"]
	if found {
		return fmt.Errorf("do not specify CMAKE_INSTALL_PREFIX")
	}
	build.CMake.Defines["CMAKE_INSTALL_PREFIX"] = bringauto_const.DockerInstallDirConst

	if build.sysroot != nil {
		build.sysroot.CreateSysrootDir()
		sysPath := build.sysroot.GetSysrootPath()
		build.Docker.SetVolume(sysPath, "/sysroot")
		build.CMake.SetDefine("CMAKE_PREFIX_PATH", "/sysroot")
	}

	gitClone := bringauto_git.GitClone{Git: *build.Git}
	gitCheckout := bringauto_git.GitCheckout{Git: *build.Git}
	gitSubmoduleUpdate := bringauto_git.GitSubmoduleUpdate{Git: *build.Git}
	startupScript := bringauto_prerequisites.CreateAndInitialize[StartupScript]()

	buildChain := BuildChain{
		Chain: []CMDLineInterface{
			startupScript,
			build.Env,
			&gitClone,
			&gitCheckout,
			&gitSubmoduleUpdate,
			build.CMake,
			build.GNUMake,
		},
	}

	logger := bringauto_log.GetLogger()
	packBuildChainLogger := logger.CreateContextLogger(build.Docker.ImageName, build.Package.GetShortPackageName(), bringauto_log.BuildChainContext)
	file, err := packBuildChainLogger.GetFile()

	if err != nil {
		logger.Error("Failed to open file - %s", err)
		return err
	}

	defer file.Close()

	shellEvaluator := bringauto_ssh.ShellEvaluator{
		Commands: buildChain.GenerateCommands(),
		StdOut:   file,
	}

	err = bringauto_prerequisites.Initialize(build.Docker)
	if err != nil {
		return err
	}

	dockerRun := (*bringauto_docker.DockerRun)(build.Docker)
	removeHandler := bringauto_process.SignalHandlerAddHandler(build.stopAndRemoveContainer)
	defer removeHandler()

	err = dockerRun.Run()
	if err != nil {
		return err
	}
	defer build.stopAndRemoveContainer()

	err = shellEvaluator.RunOverSSH(*build.SSHCredentials)
	if err != nil {
		return err
	}

	logger.InfoIndent("Copying install files from container to local directory")

	err = build.downloadInstalledFiles()
	return err
}

func (build *Build) SetSysroot(sysroot *bringauto_sysroot.Sysroot) {
	build.sysroot = sysroot
}

func (build *Build) GetLocalInstallDirPath() string {
	workingDir, err := os.Getwd()
	if err != nil {
		logger := bringauto_log.GetLogger()
		logger.Fatal("cannot call Getwd - %s", err)
	}
	copyBaseDir := filepath.Join(workingDir, localInstallDirNameConst)
	return copyBaseDir
}

func (build *Build) stopAndRemoveContainer() error {
	var err error

	dockerStop := (*bringauto_docker.DockerStop)(build.Docker)
	dockerRm := (*bringauto_docker.DockerRm)(build.Docker)
	logger := bringauto_log.GetLogger()
	err = dockerStop.Stop()
	if err != nil {
		logger.Error("Can't stop container - %s", err)
	}
	err = dockerRm.RemoveContainer()
	if err != nil {
		logger.Error("Can't remove container - %s", err)
	}
	return nil
}

func (build *Build) CleanUp() error {
	var err error
	copyDir := build.GetLocalInstallDirPath()
	if _, err = os.Stat(copyDir); os.IsNotExist(err) {
		return nil
	}
	err = os.RemoveAll(copyDir)
	if err != nil {
		return err
	}
	return nil
}

func (build *Build) downloadInstalledFiles() error {
	var err error

	copyDir := build.GetLocalInstallDirPath()
	if _, err = os.Stat(copyDir); os.IsNotExist(err) {
		err = os.MkdirAll(copyDir, 0766)
		if err != nil {
			return fmt.Errorf("cannot create directory %s", copyDir)
		}
	}

	packTarLogger := bringauto_log.GetLogger().CreateContextLogger(build.Docker.ImageName, build.Package.GetShortPackageName(), bringauto_log.TarContext)
	logFile, err := packTarLogger.GetFile()

	if err != nil {
		return fmt.Errorf("failed to open file - %s", err)
	}

	defer logFile.Close()

	sftpClient := bringauto_ssh.SFTP{
		RemoteDir:      bringauto_const.DockerInstallDirConst,
		EmptyLocalDir:  copyDir,
		SSHCredentials: build.SSHCredentials,
		LogWriter:      logFile,
	}
	err = sftpClient.DownloadDirectory()
	return err
}
