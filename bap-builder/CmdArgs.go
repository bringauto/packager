package main

import (
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
	// If true the program is in the "Sysroot" mode
	CreateSysroot       bool
	BuildPackageArgs    BuildPackageCmdLineArgs
	CreateSysrootArgs   CreateSysrootCmdLineArgs
	buildImageParser    *argparse.Command
	buildPackageParser  *argparse.Command
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

	cmd.buildImageParser = cmd.parser.NewCommand("build-image", "Build Docker image")
	cmd.BuildImagesArgs.All = cmd.buildImageParser.Flag("", "all",
		&argparse.Options{
			Required: false,
			Default:  false,
		},
	)
	cmd.BuildImagesArgs.Name = cmd.buildImageParser.String("", "name",
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
	cmd.CreateSysroot = cmd.createSysrootParser.Happened()

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
