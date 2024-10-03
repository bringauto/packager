package bringauto_log

import (
	"log/slog"
	"context"
	"io"
	"fmt"
	"strconv"
	"strings"
)

// Color codes constants
const (
	red = 31
	blue = 34
	orange = 93
	white = 97
)

// Handler
// Struct for setting style of logging specified by slog module.
type Handler struct {
	// writer Writer to use for logs
	writer io.Writer
}

// NewHandler
// Returns new Handler struct with writer.
func NewHandler(writer io.Writer) *Handler {
	handler := &Handler{writer: writer}

	return handler
}

// Enabled
// Mandatory function as specified by slog module.
func (handler *Handler) Enabled(_ context.Context, _ slog.Level) bool {
	return true
}

// colorizeLevel
// Returns colorized level string.
func colorizeLevel(level slog.Level) string {
	var colorCode int
	switch level.String() {
	case "INFO" :
		colorCode = white
	case "WARN" :
		colorCode = orange
	case "ERROR" :
		colorCode = red
	}

	return fmt.Sprintf("\033[%sm%s\033[0m", strconv.Itoa(colorCode), level.String())
}

// Handle
// Mandatory function which sets style as specified by slog module.
func (handler *Handler) Handle(_ context.Context, r slog.Record) error {
	var stringBuilder strings.Builder

	// Append timestamp
	if !r.Time.IsZero() {
		stringBuilder.WriteString(r.Time.Format("2006-01-02 15:04:05"))
		stringBuilder.WriteByte(' ')
	}

	// Append colored level
	stringBuilder.WriteString(colorizeLevel(r.Level))

	// Append message with attributes
	stringBuilder.WriteString(fmt.Sprintf(" %s\n", r.Message))

	_, err := io.WriteString(handler.writer, stringBuilder.String())
	return err
}

// WithGroup
// Mandatory function as specified by slog module. Does not adjust logging.
func (handler *Handler) WithGroup(name string) slog.Handler {
	return handler
}

// WithAttrs
// Mandatory function as specified by slog module. Does not adjust logging.
func (handler *Handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return handler
}
