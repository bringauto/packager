// Bringauto package for consistent logging in packager.
//
// The Logger is used for logging to console and for creating ContextLoggers. ContextLogger is used
// for logging output of tools or programs to log files.
//
// The Logger should be created and initialized with bringauto_prerequisites.CreateAndInitialize
// at the beginning of the program. Then anywhere in the codebase the bringauto_log.GetLogger()
// can be called to get created Logger singleton. Thanks to singleton design pattern, the Logger
// doesn't have to be forwarded throughout the codebase. Logger can create ContextLoggers, which
// can return writable log file with GetFile() method. These log file can be used for e.g. build
// output from docker container.
package bringauto_log

import (
	"bringauto/modules/bringauto_prerequisites"
	"io"
	"log/slog"
	"os"
	"time"
	"fmt"
)

const (
	indent = "    "
)

// loggerSingleton Singleton module global variable for Logger.
var loggerSingleton *Logger

// GetLogger
// Returns Logger singleton to use for logging.
func GetLogger() *Logger {
	if loggerSingleton == nil {
		loggerSingleton = bringauto_prerequisites.CreateAndInitialize[Logger]()
		loggerSingleton.Warn("Logger was not initialized. Printing to console.")
	}
	return loggerSingleton
}

// Struct which is used for logging on program level and for creating ContextLoggers.
type Logger struct {
	// slogger slog.Logger struct.
	slogger *slog.Logger
	timestamp time.Time
	logDirPath string
}

type loggerInitArgs struct {
	// Timestamp Current timestamp used for creating ContextLoggers.
	Timestamp  time.Time
	// LogDirPath Directory path, where created ContextLoggers will save logs.
	LogDirPath string
}

func (logger *Logger) FillDefault(*bringauto_prerequisites.Args) error {
	logger.slogger = getDefaultLogger(os.Stdout)
	logger.timestamp = time.Time{}
	logger.logDirPath = ""
	return nil
}

func (logger *Logger) FillDynamic(args *bringauto_prerequisites.Args) error {
	if !bringauto_prerequisites.IsEmpty(args) {
		var argsStruct loggerInitArgs
		bringauto_prerequisites.GetArgs(args, &argsStruct)
		logger.timestamp = argsStruct.Timestamp
		logger.logDirPath = argsStruct.LogDirPath + "/" + logger.getTimestampString()
	}
	return nil
}

func (logger *Logger) CheckPrerequisites(*bringauto_prerequisites.Args) error {
	loggerSingleton = logger
	return nil
}

// getDefaultLogger
// Returns default logger with style defined by Handler struct.
func getDefaultLogger(writer io.Writer) *slog.Logger {
	return slog.New(NewHandler(writer))
}

// getTimestampString
// Return timestamp formatted string for use in path.
func (logger *Logger) getTimestampString() string {
	return logger.timestamp.Format("2006-01-02_15:04:05")
}

// Info
// Global logging function with Info level. Formatted string with args can be added similarly as
// with fmt.printf function.
func (logger *Logger) Info(msg string, args ...any)  {
	if len(args) == 0 {
		logger.slogger.Info(msg)
	} else {
		logger.slogger.Info(fmt.Sprintf(msg, args...))
	}
}

// InfoIndent
// Global logging function with Info level with added pre-indent. Formatted string with args can be
// added similarly as with fmt.printf function.
func (logger *Logger) InfoIndent(msg string, args ...any)  {
	if len(args) == 0 {
		logger.slogger.Info(indent + msg)
	} else {
		logger.slogger.Info(indent + fmt.Sprintf(msg, args...))
	}
}

// Warn
// Global logging function with Warning level. Formatted string with args can be added similarly as
// with fmt.printf function.
func (logger *Logger) Warn(msg string, args ...any) {
	if len(args) == 0 {
		logger.slogger.Warn(msg)
	} else {
		logger.slogger.Warn(fmt.Sprintf(msg, args...))
	}
}

// WarnIndent
// Global logging function with Warning level with added pre-indent. Formatted string with args can
// be added similarly as with fmt.printf function.
func (logger *Logger) WarnIndent(msg string, args ...any) {
	if len(args) == 0 {
		logger.slogger.Warn(indent + msg)
	} else {
		logger.slogger.Warn(indent + fmt.Sprintf(msg, args...))
	}
}

// Error
// Global logging function with Error level. Formatted string with args can be added similarly as
// with fmt.printf function.
func (logger *Logger) Error(msg string, args ...any) {
	if len(args) == 0 {
		logger.slogger.Error(msg)
	} else {
		logger.slogger.Error(fmt.Sprintf(msg, args...))
	}
}

// ErrorIndent
// Global logging function with Error level with added pre-indent. Formatted string with args can
// be added similarly as with fmt.printf function.
func (logger *Logger) ErrorIndent(msg string, args ...any) {
	if len(args) == 0 {
		logger.slogger.Error(indent + msg)
	} else {
		logger.slogger.Error(indent + fmt.Sprintf(msg, args...))
	}
}

// Fatal
// Global logging function with Error level. Formatted string with args can be added similarly as
// with fmt.printf function. After writing a log, the whole program exits with code 1.
func (logger *Logger) Fatal(msg string, args ...any) {
	if len(args) == 0 {
		logger.slogger.Error(msg)
	} else {
		logger.slogger.Error(fmt.Sprintf(msg, args...))
	}
	os.Exit(1)
}

// CreateContextLogger
// Creates ContextLogger for specified imageName, packageName and logContext.
func (logger *Logger) CreateContextLogger(imageName string, packageName string, logContext string) *ContextLogger {
	packageContextLogger := bringauto_prerequisites.CreateAndInitialize[ContextLogger](
		logger.logDirPath, imageName, packageName, logContext,
	)
	return packageContextLogger
}
