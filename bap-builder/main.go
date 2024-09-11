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
	logger := bringauto_prerequisites.CreateAndInitialize[bringauto_log.GlobalLogger](time.Now(), "./log")

	args.InitFlags()
	err = args.ParseArgs(os.Args)
	if err != nil {
		return
	}
	bringauto_process.RegisterSignal(syscall.SIGINT)

	if args.BuildImage {
		err = BuildDockerImage(&args.BuildImagesArgs, *args.Context)
		if err != nil {
			logger.Error(err.Error())
			return
		}
		return
	}

	if args.BuildPackage {
		err = BuildPackage(&args.BuildPackageArgs, *args.Context)
		if err != nil {
			logger.Error(err.Error())
			return
		}
		return
	}

	return
}
