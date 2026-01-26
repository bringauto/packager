package build

import (
	"github.com/bacpack-system/packager/internal/constants"
	"github.com/bacpack-system/packager/internal/docker"
	"github.com/bacpack-system/packager/internal/git"
	"github.com/bacpack-system/packager/internal/log"
	"github.com/bacpack-system/packager/internal/bacpack_package"
	"github.com/bacpack-system/packager/internal/prerequisites"
	"github.com/bacpack-system/packager/internal/ssh"
	"github.com/bacpack-system/packager/internal/sysroot"
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
)

type Build struct {
	Env            *EnvironmentVariables
	Docker         *docker.Docker
	Git            *git.Git
	BuildSystem    *BuildSystem
	SSHCredentials *ssh.SSHCredentials
	Package        *bacpack_package.Package
	BuiltPackage   *sysroot.BuiltPackage
	UseLocalRepo   bool
	sysroot        *sysroot.Sysroot
}

type buildInitArgs struct {
	DockerImageName string
	DockerPort uint16
}

// FillDefault
// It fills up defaults for all members in the Build structure.
func (build *Build) FillDefault(args *prerequisites.Args) error {
	var argsStruct buildInitArgs
	var err error
	prerequisites.GetArgs(args, &argsStruct)
	if build.Git == nil {
		build.Git, err = prerequisites.CreateAndInitialize[git.Git]()
		if err != nil {
			return err
		}
	}
	if build.Docker == nil {
		build.Docker, err = prerequisites.CreateAndInitialize[docker.Docker](argsStruct.DockerImageName, argsStruct.DockerPort)
		if err != nil {
			return err
		}
	}
	if build.SSHCredentials == nil {
		build.SSHCredentials, err = prerequisites.CreateAndInitialize[ssh.SSHCredentials]()
		if err != nil {
			return err
		}
		build.SSHCredentials.Port = build.Docker.Port
	}
	if build.BuildSystem == nil {
		build.BuildSystem, err = prerequisites.CreateAndInitialize[BuildSystem]()
		if err != nil {
			return err
		}
	}
	if build.Env == nil {
		build.Env, err = prerequisites.CreateAndInitialize[EnvironmentVariables]()
		if err != nil {
			return err
		}
	}

	if build.Package == nil {
		build.Package, err = prerequisites.CreateAndInitialize[bacpack_package.Package]()
		if err != nil {
			return err
		}
	}

	build.UseLocalRepo = false

	return nil
}

func (build *Build) FillDynamic(*prerequisites.Args) error {
	return nil
}

func (build *Build) CheckPrerequisites(*prerequisites.Args) error {
	copyDir := build.GetLocalInstallDirPath()
	if _, err := os.Stat(copyDir); !os.IsNotExist(err) {
		return fmt.Errorf("package directory exist. Please delete it: %s", copyDir)
	}

	return nil
}

// performPreBuildTasks
// Downloads Package files for build in docker container. Clones the repository and updates all
// submodules.
func (build *Build) performPreBuildTasks(shellEvaluator *ssh.ShellEvaluator) error {
	gitClone := git.GitClone{Git: *build.Git}
	gitCheckout := git.GitCheckout{Git: *build.Git}
	gitSubmoduleUpdate := git.GitSubmoduleUpdate{Git: *build.Git}
	startupScript, err := prerequisites.CreateAndInitialize[StartupScript]()
	if err != nil {
		return err
	}

	if build.UseLocalRepo {
		build.Env.Env["BA_PACKAGE_LOCAL_PATH"] = constants.ContainerPackageRepoPath
	}

	startupChain := BuildChain{
		Chain: []CMDLineInterface{
			startupScript,
		},
	}	
	preparePackageChain := BuildChain{
		Chain: []CMDLineInterface{
			build.Env,
			&gitClone,
			&gitCheckout,
			&gitSubmoduleUpdate,
		},
	}

	shellEvaluator.PreparingCommands = startupChain.GenerateCommands()
	shellEvaluator.Commands = preparePackageChain.GenerateCommands()

	err = shellEvaluator.RunOverSSH(*build.SSHCredentials)
	if err != nil {
		logger := log.GetLogger()
		logger.Error("Failed to clone or checkout git repository, check the log file, is the git URI and revision correct?")
		return err
	}

	return nil
}
// prepareForBuild
// Prepares some fields of Build struct for build and makes pre build checks.
func (build *Build) prepareForBuild() error {
	err := build.CheckPrerequisites(nil)
	if err != nil {
		return err
	}

	if build.BuiltPackage == nil {
		return fmt.Errorf("BuiltPackage is nil")
	}

	build.Git.ClonePath = dockerGitCloneDirConst
	build.BuildSystem.SourceDir = dockerGitCloneDirConst
	build.BuildSystem.InstallPrefix = constants.DockerInstallDirConst

	if build.sysroot != nil {
		err := build.sysroot.CreateSysrootDir()
		if err != nil {
			return err
		}
		sysPath, err := build.sysroot.GetSysrootPath()
		if err != nil {
			return err
		}
		err = build.Docker.SetVolume(sysPath, constants.ContainerSysrootPath)
		if err != nil {
			return err
		}
		build.BuildSystem.PrefixPath = constants.ContainerSysrootPath
	}

	return nil
}

// RunBuild
// Creates a Docker container, performs a build in it. After build, the files are downloaded to
// local directory and Docker container is stopped and removed. Returns bool which indicates if
// the build was performed succesfully.
func (build *Build) RunBuild() (error, bool) { // Long function - it is hard to refactor to make readability better
	err := build.prepareForBuild()
	if err != nil {
		return err, false
	}

	shellEvaluator := ssh.ShellEvaluator{
		Commands: []string{},
	}

	logger := log.GetLogger()
	packBuildChainLogger := logger.CreateContextLogger(build.Docker.ImageName, build.Package.GetShortPackageName(), log.BuildChainContext)
	if packBuildChainLogger != nil {
		file, err := packBuildChainLogger.GetFile()
		if err != nil {
			logger.Error("Failed to open file - %s", err)
			return err, false
		}
		defer file.Close()

		shellEvaluator.StdOut = file
	}

	dockerRun := (*docker.DockerRun)(build.Docker)
	removeHandler := dockerRun.GetUndoHandler()
	defer removeHandler()

	logger.InfoIndent("Starting docker container")

	err = dockerRun.Run()
	if err != nil {
		return err, false
	}
	build.SSHCredentials.Port = build.Docker.Port

	logger.InfoIndent("Cloning Package git repository inside docker container")

	err = build.performPreBuildTasks(&shellEvaluator)
	if err != nil {
		return err, false
	}

	build.BuiltPackage.GitCommitHash, err = build.getGitCommitHash()
	if err != nil {
		return fmt.Errorf("can't get git commit hash from container - %w", err), false
	}
	build.BuiltPackage.DirName = build.sysroot.GetDirNameInSysroot()

	if build.sysroot.IsPackageInSysroot(*build.BuiltPackage) {
		logger.InfoIndent("Package already built in sysroot - skipping build")
		return nil, false
	}
	startupScript, err := prerequisites.CreateAndInitialize[StartupScript]()
	if err != nil {
		return err, false
	}

	startupChain := BuildChain{
		Chain: []CMDLineInterface{
			startupScript,
		},
	}
	buildChain := BuildChain{
		Chain: []CMDLineInterface{
			build.Env,
			build.BuildSystem,
		},
	}

	shellEvaluator.PreparingCommands = startupChain.GenerateCommands()
	shellEvaluator.Commands = buildChain.GenerateCommands()

	logger.InfoIndent("Running build inside container")

	err = shellEvaluator.RunOverSSH(*build.SSHCredentials)
	if err != nil {
		return fmt.Errorf("build failed inside docker container, check the log file"), false
	}

	logger.InfoIndent("Copying install files from container to local directory")

	err = build.downloadInstalledFiles()
	if err != nil {
		return fmt.Errorf("can't download files from container to local directory"), false
	}

	return nil, true
}

func (build *Build) SetSysroot(sysroot *sysroot.Sysroot) {
	build.sysroot = sysroot
}

func (build *Build) GetLocalInstallDirPath() string {
	workingDir, err := os.Getwd()
	if err != nil {
		logger := log.GetLogger()
		logger.Fatal("cannot call Getwd - %s", err)
	}
	suffix := ""
	if build.Docker.Port != constants.DefaultSSHPort {
		suffix = strconv.Itoa(int(build.Docker.Port) - constants.DefaultSSHPort)
	}
	copyBaseDir := filepath.Join(workingDir, localInstallDirNameConst + suffix)
	return copyBaseDir
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

	sftpClient := ssh.SFTP{
		RemoteDir:      constants.DockerInstallDirConst,
		EmptyLocalDir:  copyDir,
		SSHCredentials: build.SSHCredentials,
	}

	packTarLogger := log.GetLogger().CreateContextLogger(build.Docker.ImageName, build.Package.GetShortPackageName(), log.TarContext)
	if packTarLogger != nil {
		logFile, err := packTarLogger.GetFile()
		if err != nil {
			return fmt.Errorf("failed to open file - %w", err)
		}
		defer logFile.Close()

		sftpClient.LogWriter = logFile
	}

	err = sftpClient.DownloadDirectory()
	return err
}

func (build *Build) getGitCommitHash() (string, error) {
	pipeReader, pipeWriter := io.Pipe()
	defer pipeReader.Close()
	defer pipeWriter.Close()
	gitGetHash := git.GitGetHash{Git: *build.Git}
	shellEvaluator := ssh.ShellEvaluator{
		Commands: gitGetHash.ConstructCMDLine(),
		StdOut:   pipeWriter,
	}

	err := shellEvaluator.RunOverSSH(*build.SSHCredentials)
	if err != nil {
		return "", err
	}

	buf := bufio.NewReader(pipeReader)
	var line string

	for {
		line, err = buf.ReadString('\n')
		if err != nil && err != io.EOF {
			return "", err
		}
		if err == nil { // The newline character is present
			line = line[:len(line)-1]
		}

		hash := getGitCommitHashFromLine(line)
		if hash != "" {
			return hash, nil
		}

		if err == io.EOF {
			break
		}
	}

	return "", fmt.Errorf("no commit hash in output")
}

func getGitCommitHashFromLine(line string) string {
	// regexp for long git commit hash, it must be used, because the ssh output has several commands and it is long
	re := regexp.MustCompile("[a-f0-9]{40}")
	match := re.FindStringSubmatch(line)
	if len(match) > 0 {
		return match[0]
	}

	return ""
}
