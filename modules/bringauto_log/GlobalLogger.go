package bringauto_log

import (
	"bringauto/modules/bringauto_prerequisites"
	"os"
	"time"
)

// globalLoggerSingleton Singleton module global variable for GlobalLogger.
var globalLoggerSingleton *GlobalLogger

// GetLogger
// Returns GlobalLogger singleton to use for logging.
func GetLogger() *GlobalLogger {
	if globalLoggerSingleton == nil {
		globalLoggerSingleton = bringauto_prerequisites.CreateAndInitialize[GlobalLogger]()
		globalLoggerSingleton.Warn("Global logger was not initialized. Printing to console.")
	}
	return globalLoggerSingleton
}

// GlobalLogger
// Struct used for logging on program level.
type GlobalLogger struct {
	Logger
}

type globalLoggerInitArgs struct {
	// Timestamp Current timestamp used for creating ContextLoggers.
	Timestamp  time.Time
	// LogDirPath Directory path, where created ContextLoggers will save logs.
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

// getTimestampString
// Return timestamp formatted string for use in path.
func (logger *GlobalLogger) getTimestampString() string {
	return logger.timestamp.Format("2006-01-02_15:04:05")
}

func (logger *GlobalLogger) CheckPrerequisites(*bringauto_prerequisites.Args) error {
	globalLoggerSingleton = logger
	return nil
}

// CreateContextLogger
// Creates ContextLogger for specified imageName, packageName and logContext.
func (logger *GlobalLogger) CreateContextLogger(imageName string, packageName string, logContext string) *ContextLogger {
	packageContextLogger := bringauto_prerequisites.CreateAndInitialize[ContextLogger](
		logger.logDirPath, imageName, packageName, logContext,
	)
	return packageContextLogger
}
