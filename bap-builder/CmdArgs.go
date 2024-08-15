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

type OutputDirMode int8

const (
	OutputDirModeGitLFS OutputDirMode = iota
)

// BuildPackageCmdLineArgs
// Options/setting for Package mode
type BuildPackageCmdLineArgs struct {
	// All build all packages in package/ directory
	All *bool
	// Name of the package to build (name of the directory in packages/ dir)
	Name *string
	// Build all dependencies of package when building single package
	BuildDeps *bool
	// DockerImageName is a name of docker image to which packages will be build.
	// If empty all docker images from DockerMatrix in config file are used for a given package.
	// If not empty, only packages which contains DockerImageName in DockerMatrix will be built.
	// If not empty, packages are built only by toolchain represented by DockerImageName
	DockerImageName *string
	// OutputDir relative (to program working dir) ot absolute path where the package will be stored
	OutputDir *string
	// Output dir mode
	OutputDirMode *OutputDirMode
}

// CmdLineArgs
// Represents Cmd line arguments passed to  cmd line of the target program.
// Program operates in two modes
// - build Docker images (Docker mode),
// - build package (package mode)
// Exactly one of these modes can be active in a time.
type CmdLineArgs struct {
	// Absolute/relative path to config directory
	Context *string
	// If true the program is in the "Docker" mode
	BuildImage bool
	// Standard Cmd line arguments for Docker mode
	BuildImagesArgs BuildImageCmdLineArgs
	// If true the program is in the "Package" mode
	BuildPackage       bool
	BuildPackageArgs   BuildPackageCmdLineArgs
	buildImageParser   *argparse.Command
	buildPackageParser *argparse.Command
	parser             *argparse.Parser
}

// InitFlags
// Initialize flags and fill up CmdLineArgs struct
// Function must be called before any use of CmdLineArgs
func (cmd *CmdLineArgs) InitFlags() {
	cmd.parser = argparse.NewParser("BringAuto Packager", "Build and track C++ dependencies")
	cmd.Context = cmd.parser.String("", "context",
		&argparse.Options{
			Required: false,
			Default:  ".",
			Help:     "Command context",
		},
	)

	cmd.buildPackageParser = cmd.parser.NewCommand("build-package", "Build package")
	cmd.BuildPackageArgs.All = cmd.buildPackageParser.Flag("", "all",
		&argparse.Options{
			Required: false,
			Help:     "Build all packages in the given context",
			Default:  false,
		},
	)
	cmd.BuildPackageArgs.Name = cmd.buildPackageParser.String("", "name",
		&argparse.Options{
			Required: false,
			Default:  "",
			Help:     "Name of the package to build",
		},
	)
	cmd.BuildPackageArgs.BuildDeps = cmd.buildPackageParser.Flag("", "build-deps",
		&argparse.Options{
			Required: false,
			Default:  false,
			Help:     "Build all dependencies of package when building single package",
		},
	)
	cmd.BuildPackageArgs.OutputDir = cmd.buildPackageParser.String("", "output-dir",
		&argparse.Options{
			Required: true,
			Help:     "Directory where to store built package",
		},
	)
	cmd.BuildPackageArgs.DockerImageName = cmd.buildPackageParser.String("", "image-name",
		&argparse.Options{
			Required: false,
			Default:  "",
			Help: "Docker image name for which packages will be build.\n" +
				"Only packages that contains image-name in the DockerMatrix will be built.\n" +
				"Given packages will be build by toolchain represented by image-name",
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

	outputMode := OutputDirModeGitLFS
	cmd.BuildPackageArgs.OutputDirMode = &outputMode

	cmd.BuildImage = cmd.buildImageParser.Happened()
	cmd.BuildPackage = cmd.buildPackageParser.Happened()

	return nil
}
