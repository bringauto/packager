package main

import (
	"bringauto/modules/bringauto_log"
	"bringauto/modules/bringauto_prerequisites"
	"bringauto/modules/bringauto_process"
	"os"
	"time"
	"syscall"
)

func main() {
	var err error
	var args CmdLineArgs
	logger := bringauto_prerequisites.CreateAndInitialize[bringauto_log.Logger](time.Now(), "./log")

	args.InitFlags()
	err = args.ParseArgs(os.Args)
	if err != nil {
		logger.Error("Can't parse cmd line arguments - %s", err)
		return
	}
	bringauto_process.SignalHandlerRegisterSignal(syscall.SIGINT)

	if args.BuildImage {
		err = BuildDockerImage(&args.BuildImagesArgs, *args.Context)
		if err != nil {
			logger.Error("Failed to build Docker image: %s", err)
			return
		}
		return
	}

	if args.BuildPackage {
		err = BuildPackage(&args.BuildPackageArgs, *args.Context)
		if err != nil {
			logger.Error("Failed to build package: %s", err)
			return
		}
		return
	}

	if args.CreateSysroot {
		err = CreateSysroot(&args.CreateSysrootArgs, *args.Context)
		if err != nil {
			logger.Error("Failed to create sys: %s", err)
			return
		}
		return
	}

	return
}
