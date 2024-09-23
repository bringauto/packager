// Bringauto package for consistent logging in packager.
//
// The Logger struct is a base for other loggers. The GlobalLogger is used for logging to console
// and for creating ContextLoggers. ContextLogger is used for logging output of tools or programs
// to log files.
//
// The Logger struct shouldn't be used directly. The GlobalLogger should be created and initialized
// with bringauto_prerequisites.CreateAndInitialize at the beginning of the program. Then anywhere
// in the codebase the bringauto_log.GetLogger() can be called to get created GlobalLogger
// singleton. Thanks to singleton design pattern, the GlobalLogger doesn't have to be forwarded
// throughout the codebase. GlobalLogger can create ContextLoggers, which can return writable log
// file with GetFile() method. These log file can be used for e.g. build output from docker
// container.
package bringauto_log

import (
	"io"
	"log/slog"
	"os"
	"time"
	"fmt"
)

const (
	indent = "    "
)

// Struct which is used as a base for GlobalLogger and ContextLogger. Contains methods for global
// logging.
type Logger struct {
	// slogger slog.Logger struct.
	slogger *slog.Logger
	timestamp time.Time
	logDirPath string
}

// getDefaultLogger
// Returns default logger with style defined by Handler struct.
func getDefaultLogger(writer io.Writer) *slog.Logger {
	return slog.New(NewHandler(writer))
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
