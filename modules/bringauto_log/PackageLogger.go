package bringauto_log

import (
	"bringauto/modules/bringauto_prerequisites"
	"os"
	"time"
)

type PackageLogger struct {
	Logger
}

type packageLoggerInitArgs struct {
	Timestamp time.Time
	LogDirPath string
	PackageName string
}

func (logger *PackageLogger) FillDefault(*bringauto_prerequisites.Args) error {
	logger.slogger = getDefaultLogger(os.Stdout)
	logger.timestamp = time.Time{}
	logger.logDirPath = ""
	return nil
}

func (logger *PackageLogger) FillDynamic(args *bringauto_prerequisites.Args) error {
	if !bringauto_prerequisites.IsEmpty(args) {
		var argsStruct packageLoggerInitArgs
		bringauto_prerequisites.GetArgs(args, &argsStruct)
		logger.timestamp = argsStruct.Timestamp
		logger.logDirPath = argsStruct.LogDirPath + "/" + argsStruct.PackageName
	}
	return nil
}

func (logger *PackageLogger) CheckPrerequisites(*bringauto_prerequisites.Args) error {
	return nil
}

func (logger *PackageLogger) CreatePackageContextLogger(logContext string) *PackageContextLogger {
	packageContextLogger := bringauto_prerequisites.CreateAndInitialize[PackageContextLogger](
		logger.timestamp, logger.logDirPath, logContext,
	)
	return packageContextLogger
}
