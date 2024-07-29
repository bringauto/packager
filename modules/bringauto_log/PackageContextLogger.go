package bringauto_log

import (
	"bringauto/modules/bringauto_prerequisites"
	"os"
	"time"
)

const (
	BuildChainContext = "build_chain"
	TarContext = "tar"
)

type PackageContextLogger struct {
	Logger
	logFileName string
}

type packageContextLoggerInitArgs struct {
	Timestamp time.Time
	LogDirPath string
	PackageName string
	LogContext string
}

func (logger *PackageContextLogger) FillDefault(*bringauto_prerequisites.Args) error {
	logger.slogger = getDefaultLogger(os.Stdout)
	logger.timestamp = time.Time{}
	logger.logDirPath = ""
	logger.logFileName = ""
	return nil
}

func (logger *PackageContextLogger) FillDynamic(args *bringauto_prerequisites.Args) error {
	if !bringauto_prerequisites.IsEmpty(args) {
		var argsStruct packageContextLoggerInitArgs
		bringauto_prerequisites.GetArgs(args, &argsStruct)
		logger.timestamp = argsStruct.Timestamp
		logger.logDirPath = argsStruct.LogDirPath + "/" + argsStruct.PackageName
		logger.logFileName = argsStruct.LogContext + ".txt"
	}
	return nil
}

func (logger *PackageContextLogger) initLogDir() error {
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

func (logger *PackageContextLogger) CheckPrerequisites(*bringauto_prerequisites.Args) error {
	if logger.logDirPath != "" {
		logger.initLogDir()
	}
	return nil
}

func (logger *PackageContextLogger) GetFile() (*os.File, error) {
	return os.OpenFile(logger.logDirPath + "/" + logger.logFileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
}
