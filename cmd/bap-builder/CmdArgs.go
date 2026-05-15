package main

import (
	"github.com/bacpack-system/packager/internal/constants"
	"fmt"
	"github.com/akamensky/argparse"
)

// BuildImageCmdLineArgs
// Options/setting for Docker mode
type BuildImageCmdLineArgs struct {
	// All build all Docker images located in docker/ directory
	All *bool
	// Name of the image to build
	Name *string
}

// BuildPackageCmdLineArgs
// Options/setting for Package mode
type BuildPackageCmdLineArgs struct {
	// All build all Packages in package/ directory
	All *bool
	// Name of the Package to build (name of the directory in packages/ dir)
	Name *string
	// BuildDeps Build all dependencies of Package when building single Package
	BuildDeps *bool
	// BuildDepsOn Build Package with all Packages which depends on it
	BuildDepsOn *bool
	// BuildDepsOn Build Package with all Packages which depends on it recursively
	BuildDepsOnRecursive *bool
	// DockerImageName is a name of docker image to which Packages will be build.
	// If empty all docker images from DockerMatrix in config file are used for a given Package.
	// If not empty, only Packages which contains DockerImageName in DockerMatrix will be built.
	// If not empty, Packages are built only by toolchain represented by DockerImageName
	DockerImageName *string
	// OutputDir relative (to program working dir) ot absolute path where the Package will be stored
	OutputDir *string
	// Port for Docker container
	Port *int
}

// BuildAppCmdLineArgs
// Options/setting for App mode
type BuildAppCmdLineArgs struct {
	// All build all Apps in app/ directory
	All *bool
	// Name of the App to build (name of the directory in app/ dir)
	Name *string
	// DockerImageName is a name of docker image to which Apps will be build.
	// If empty all docker images from DockerMatrix in config file are used for a given App.
	// If not empty, only Apps which contains DockerImageName in DockerMatrix will be built.
	// If not empty, Apps are built only by toolchain represented by DockerImageName
	DockerImageName *string
	// OutputDir relative (to program working dir) ot absolute path where the App will be stored
	OutputDir *string
	// Port for Docker container
	Port *int
	// Use local Package Repository inside docker container
	UseLocalRepo *bool
}

// CreateSysrootCmdLineArgs
// Options/setting for Sysroot mode
type CreateSysrootCmdLineArgs struct {
	// Path to the Git Lfs repository with Packages
	Repo *string
	// Name of the new sysroot directory to be created
	Sysroot *string
	// Name of the docker image which are the Packages build for
	ImageName *string
	// Port for Docker container
	Port *int
}

// CmdLineArgs
// Represents Cmd line arguments passed to  cmd line of the target program.
// Program operates in three modes
// - build Docker images (Docker mode),
// - build package (package mode)
// - create sysroot (Sysroot mode)
// Exactly one of these modes can be active in a time.
type CmdLineArgs struct {
	// Absolute/relative path to config directory
	Context *string
	// If true the program is in the "Docker" mode
	BuildImage bool
	// Standard Cmd line arguments for Docker mode
	BuildImagesArgs BuildImageCmdLineArgs
	// If true the program is in the "Package" mode
	BuildPackage        bool
	// If true the program is in the "App" mode
	BuildApp        bool
	// If true the program is in the "Sysroot" mode
	CreateSysroot       bool
	BuildPackageArgs    BuildPackageCmdLineArgs
	BuildAppArgs        BuildAppCmdLineArgs
	CreateSysrootArgs   CreateSysrootCmdLineArgs
	buildImageParser    *argparse.Command
	buildPackageParser  *argparse.Command
	buildAppParser      *argparse.Command
	createSysrootParser *argparse.Command
	parser              *argparse.Parser
}

// InitFlags
// Initialize flags and fill up CmdLineArgs struct
// Function must be called before any use of CmdLineArgs
func (cmd *CmdLineArgs) InitFlags() {
	cmd.parser = argparse.NewParser("BringAuto Packager", "Build and track C++ dependencies")
	cmd.Context = cmd.parser.String("", "context",
		&argparse.Options{
			Required: true,
			Help:     "Context directory where are the json definition of Packages",
		},
	)

	cmd.buildPackageParser = cmd.parser.NewCommand("build-package", "Build package")
	cmd.BuildPackageArgs.All = cmd.buildPackageParser.Flag("", "all",
		&argparse.Options{
			Required: false,
			Help:     "Build all Packages in the given context",
			Default:  false,
		},
	)
	cmd.BuildPackageArgs.Port = cmd.buildPackageParser.Int("p", "port",
		&argparse.Options{
			Required: false,
			Help:     "Host port for docker container ssh bind",
			Default:  constants.DefaultSSHPort,
		},
	)
	cmd.BuildPackageArgs.Name = cmd.buildPackageParser.String("", "name",
		&argparse.Options{
			Required: false,
			Default:  "",
			Help:     "Name of the Package to build",
		},
	)
	cmd.BuildPackageArgs.BuildDeps = cmd.buildPackageParser.Flag("", "build-deps",
		&argparse.Options{
			Required: false,
			Default:  false,
			Help:     "Build all dependencies of Package when building single Package",
		},
	)
	cmd.BuildPackageArgs.BuildDepsOn = cmd.buildPackageParser.Flag("", "build-deps-on",
		&argparse.Options{
			Required: false,
			Default:  false,
			Help:     "Build Packages which depends on given Package without itself, " +
			"the Packages are built with its dependencies",
		},
	)
	cmd.BuildPackageArgs.BuildDepsOnRecursive = cmd.buildPackageParser.Flag("", "build-deps-on-recursive",
		&argparse.Options{
			Required: false,
			Default:  false,
			Help:     "Build Packages which depends on given Package without itself recursively, " +
			"the Packages are built with its dependencies",
		},
	)
	cmd.BuildPackageArgs.OutputDir = cmd.buildPackageParser.String("", "output-dir",
		&argparse.Options{
			Required: true,
			Help:     "Directory where to store built Package",
		},
	)
	cmd.BuildPackageArgs.DockerImageName = cmd.buildPackageParser.String("", "image-name",
		&argparse.Options{
			Required: true,
			Validate: checkForEmpty,
			Help: "Docker image name for which Packages will be build. " +
			"Only Packages that contains image-name in the DockerMatrix will be built. " +
			"Given Packages will be build by toolchain represented by image-name",
		},
	)

	cmd.buildAppParser = cmd.parser.NewCommand("build-app", "Build App")
	cmd.BuildAppArgs.All = cmd.buildAppParser.Flag("", "all",
		&argparse.Options{
			Required: false,
			Help:     "Build all Apps in the given context",
			Default:  false,
		},
	)
	cmd.BuildAppArgs.Name = cmd.buildAppParser.String("", "name",
		&argparse.Options{
			Required: false,
			Default:  "",
			Help:     "Name of the App to build",
		},
	)
	cmd.BuildAppArgs.OutputDir = cmd.buildAppParser.String("", "output-dir",
		&argparse.Options{
			Required: true,
			Help:     "Directory where to store built Apps",
		},
	)
	cmd.BuildAppArgs.DockerImageName = cmd.buildAppParser.String("", "image-name",
		&argparse.Options{
			Required: true,
			Validate: checkForEmpty,
			Help: "Docker image name for which the Apps will be build. " +
			"Only Apps that contains image-name in the DockerMatrix will be built. " +
			"Given Apps will be build by toolchain represented by image-name",
		},
	)
	cmd.BuildAppArgs.UseLocalRepo = cmd.buildAppParser.Flag("", "use-local-repo",
		&argparse.Options{
			Required: false,
			Help: "The build of Apps will use local Package Repository instead of " +
			"upstream. The same Package Repository from output-dir option will be " +
			"used. All needed dependencies required by the App's CMakeLists must " +
			"be present in this Package Repository or the build will fail.",
		},
	)
	cmd.BuildAppArgs.Port = cmd.buildAppParser.Int("p", "port",
		&argparse.Options{
			Required: false,
			Help:     "Host port for docker container ssh bind",
			Default:  constants.DefaultSSHPort,
		},
	)

	cmd.buildImageParser = cmd.parser.NewCommand("build-image", "Build Docker image")
	cmd.BuildImagesArgs.All = cmd.buildImageParser.Flag("", "all",
		&argparse.Options{
			Required: false,
			Help:     "Build all Images in the given context.",
			Default:  false,
		},
	)
	cmd.BuildImagesArgs.Name = cmd.buildImageParser.String("", "image-name",
		&argparse.Options{
			Required: false,
			Default:  "",
			Help:     "Name of the docker image to build",
		},
	)

	cmd.createSysrootParser = cmd.parser.NewCommand("create-sysroot", "Create Sysroot")
	cmd.CreateSysrootArgs.Sysroot = cmd.createSysrootParser.String("", "sysroot-dir",
		&argparse.Options{
			Required: true,
			Help:     "Name of the sysroot directory which will be created",
		},
	)
	cmd.CreateSysrootArgs.Repo = cmd.createSysrootParser.String("", "git-lfs",
		&argparse.Options{
			Required: true,
			Help:     "Git Lfs directory where Packages are stored",
		},
	)
	cmd.CreateSysrootArgs.ImageName = cmd.createSysrootParser.String("", "image-name",
		&argparse.Options{
			Required: true,
			Validate: checkForEmpty,
			Help:     "Name of docker image which are the Packages built for",
		},
	)
	cmd.CreateSysrootArgs.Port = cmd.createSysrootParser.Int("p", "port",
		&argparse.Options{
			Required: false,
			Help:     "Host port for docker container ssh bind",
			Default:  constants.DefaultSSHPort,
		},
	)
}

// checkForEmpty
// Checks the given argument. If it is empty, returns error, else nil.
func checkForEmpty(args []string) error {
	if len(args) == 1 {
		if len(args[0]) == 0 {
			return fmt.Errorf("cannot be empty")
		}
	}
	return nil
}

// ParseArgs
// Parse arguments from given 'args' list of strings.
// Return error if cmdline is not valid or nil in case of no problem.
func (cmd *CmdLineArgs) ParseArgs(args []string) error {
	err := cmd.parser.Parse(args)
	if err != nil {
		fmt.Print(cmd.parser.Usage(err))
		return err
	}

	cmd.BuildImage = cmd.buildImageParser.Happened()
	cmd.BuildPackage = cmd.buildPackageParser.Happened()
	cmd.BuildApp = cmd.buildAppParser.Happened()
	cmd.CreateSysroot = cmd.createSysrootParser.Happened()

	return cmd.checkInvalidOptionCombinations()
}

// checkInvalidOptionCombinations
// Return appropriate error when any invalid option combination is detected, else nil.
func (cmd *CmdLineArgs) checkInvalidOptionCombinations() error {
	if cmd.BuildPackage {
		if *cmd.BuildPackageArgs.Name == "" && !*cmd.BuildPackageArgs.All {
			return fmt.Errorf("build-package requires either --name or --all")
		}
	}

	if cmd.BuildApp {
		if *cmd.BuildAppArgs.Name == "" && !*cmd.BuildAppArgs.All {
			return fmt.Errorf("build-app requires either --name or --all")
		}
	}

	if *cmd.BuildPackageArgs.All {
		if *cmd.BuildPackageArgs.BuildDeps {
			return fmt.Errorf("all and build-deps flags at the same time")
		}
		if *cmd.BuildPackageArgs.BuildDepsOn {
			return fmt.Errorf("all and build-deps-on flags at the same time")
		}
		if *cmd.BuildPackageArgs.BuildDepsOnRecursive {
			return fmt.Errorf("all and build-deps-on-recursive flags at the same time")
		}
	} else if *cmd.BuildPackageArgs.BuildDepsOn {
		if *cmd.BuildPackageArgs.BuildDepsOnRecursive {
			return fmt.Errorf("build-deps-on and build-deps-on-recursive flags at the same time")
		}
	}

	return nil
}
