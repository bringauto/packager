package bringauto_log

import (
	"io"
	"log/slog"
	"os"
	"time"
	"fmt"
)

type Logger struct {
	slogger *slog.Logger
	timestamp time.Time
	logDirPath string
}

func getDefaultLogger(writer io.Writer) *slog.Logger {
	return slog.New(NewHandler(writer))
}

func (logger *Logger) Info(msg string, args ...any)  {
	if len(args) == 0 {
		logger.slogger.Info(msg)
	} else {
		logger.slogger.Info(fmt.Sprintf(msg, args))
	}
}

func (logger *Logger) Warn(msg string, args ...any) {
	if len(args) == 0 {
		logger.slogger.Warn(msg)
	} else {
		logger.slogger.Warn(fmt.Sprintf(msg, args))
	}
}

func (logger *Logger) Error(msg string, args ...any) {
	if len(args) == 0 {
		logger.slogger.Error(msg)
	} else {
		logger.slogger.Error(fmt.Sprintf(msg, args))
	}
}

func (logger *Logger) Fatal(msg string, args ...any) {
	if len(args) == 0 {
		logger.slogger.Error(msg)
	} else {
		logger.slogger.Error(fmt.Sprintf(msg, args))
	}
	os.Exit(1)
}
