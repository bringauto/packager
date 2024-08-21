package bringauto_log

import (
	"bringauto/modules/bringauto_prerequisites"
	"os"
	"time"
)

var globalLoggerSingleton *GlobalLogger

func GetLogger() *GlobalLogger {
	if globalLoggerSingleton == nil {
		globalLoggerSingleton = bringauto_prerequisites.CreateAndInitialize[GlobalLogger]()
		globalLoggerSingleton.Warn("Global logger was not initialized. Printing to console.")
	}
	return globalLoggerSingleton
}

type GlobalLogger struct {
	Logger
}

type globalLoggerInitArgs struct {
	Timestamp  time.Time
	LogDirPath string
}

func (logger *GlobalLogger) FillDefault(*bringauto_prerequisites.Args) error {
	logger.slogger = getDefaultLogger(os.Stdout)
	logger.timestamp = time.Time{}
	logger.logDirPath = ""
	return nil
}

func (logger *GlobalLogger) FillDynamic(args *bringauto_prerequisites.Args) error {
	if !bringauto_prerequisites.IsEmpty(args) {
		var argsStruct globalLoggerInitArgs
		bringauto_prerequisites.GetArgs(args, &argsStruct)
		logger.timestamp = argsStruct.Timestamp
		logger.logDirPath = argsStruct.LogDirPath + "/" + logger.getTimestampString()
	}
	return nil
}

func (logger *GlobalLogger) getTimestampString() string {
	return logger.timestamp.Format("2006-01-02_15:04:05")
}

func (logger *GlobalLogger) CheckPrerequisites(*bringauto_prerequisites.Args) error {
	globalLoggerSingleton = logger
	return nil
}

func (logger *GlobalLogger) CreateContextLogger(imageName string, packageName string, logContext string) *ContextLogger {
	packageContextLogger := bringauto_prerequisites.CreateAndInitialize[ContextLogger](
		logger.timestamp, logger.logDirPath, imageName, packageName, logContext,
	)
	return packageContextLogger
}
