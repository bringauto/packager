package main

import (
	"fmt"
	"os"
)

func main() {
	var err error
	var args CmdLineArgs

	args.InitFlags()
	err = args.ParseArgs(os.Args)
	if err != nil {
		return
	}

	if args.BuildImage {
		err = BuildDockerImage(&args.BuildImagesArgs, *args.Context)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return
		}
		return
	}

	if args.BuildPackage {
		err = BuildPackage(&args.BuildPackageArgs, *args.Context)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return
		}
		return
	}

	return
}
