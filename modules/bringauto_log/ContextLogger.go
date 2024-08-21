package bringauto_log

import (
	"bringauto/modules/bringauto_prerequisites"
	"os"
)

const (
	BuildChainContext = "build_chain"
	TarContext = "tar"
	ImageBuildContext = "image_build"
)

type ContextLogger struct {
	Logger
	logFileName string
}

type contextLoggerInitArgs struct {
	LogDirPath string
	ImageName string
	PackageName string // If is empty, then it is considered as non-package log
	LogContext string
}

func (logger *ContextLogger) FillDefault(*bringauto_prerequisites.Args) error {
	logger.slogger = getDefaultLogger(os.Stdout)
	logger.logDirPath = ""
	logger.logFileName = ""
	return nil
}

func (logger *ContextLogger) FillDynamic(args *bringauto_prerequisites.Args) error {
	if !bringauto_prerequisites.IsEmpty(args) {
		var argsStruct contextLoggerInitArgs
		bringauto_prerequisites.GetArgs(args, &argsStruct)
		logger.logDirPath = argsStruct.LogDirPath + "/" + argsStruct.ImageName + "/" + argsStruct.PackageName
		logger.logFileName = argsStruct.LogContext + ".txt"
	}
	return nil
}

func (logger *ContextLogger) initLogDir() error {
	_, err := os.Stat(logger.logDirPath)
	if os.IsNotExist(err) {
		err = os.MkdirAll(logger.logDirPath, 0700)
		if err != nil {
			logger.Error("Failed to create log directory - %s", err)
			return err
		}
	}
	return nil
}

func (logger *ContextLogger) CheckPrerequisites(*bringauto_prerequisites.Args) error {
	if logger.logDirPath != "" {
		logger.initLogDir()
	}
	return nil
}

func (logger *ContextLogger) GetFile() (*os.File, error) {
	return os.OpenFile(logger.logDirPath + "/" + logger.logFileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
}
