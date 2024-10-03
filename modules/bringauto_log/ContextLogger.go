package bringauto_log

import (
	"bringauto/modules/bringauto_prerequisites"
	"os"
	"path/filepath"
)

const (
	BuildChainContext = "build_chain"
	TarContext = "tar"
	ImageBuildContext = "image_build"
)

// ContextLogger
// Is used for getting writable file for context logs of a package.
type ContextLogger struct {
	logDirPath string
	// logFileName Whole log file name with context and extrnsion
	logFileName string
}

type contextLoggerInitArgs struct {
	// LogDirPath Directory path, where logs will be save.
	LogDirPath string
	// ImageName Image name to use as part of path to log file.
	ImageName string
	// PackageName Package name to use as part of path to log file. If is empty, then it is considered
	// as non-package log.
	PackageName string
	// LogContext Context of a log. Is used as log file name.
	LogContext string
}

func (logger *ContextLogger) FillDefault(*bringauto_prerequisites.Args) error {
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

// initLogDir
// Creates directory for logs if it does not exists already.
func (logger *ContextLogger) initLogDir() error {
	_, err := os.Stat(logger.logDirPath)
	if os.IsNotExist(err) {
		err = os.MkdirAll(logger.logDirPath, 0700)
		if err != nil {
			GetLogger().Error("Failed to create log directory - %s", err)
			return err
		}
	}
	return nil
}

func (logger *ContextLogger) CheckPrerequisites(*bringauto_prerequisites.Args) error {
	if logger.logDirPath != "" {
		return logger.initLogDir()
	}
	return nil
}

// GetFile
// Returns writable file for writing logs for specified context of a package. The file must be
// closed by the caller.
func (logger *ContextLogger) GetFile() (*os.File, error) {
	filepath := filepath.Join(logger.logDirPath, logger.logFileName)
	return os.OpenFile(filepath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
}
