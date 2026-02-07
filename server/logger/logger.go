package logger

import (
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog"
	"gopkg.in/natefinch/lumberjack.v2"
)

var Log zerolog.Logger

// InitLogger initializes the global logger
func InitLogger(level, logPath string, maxSize, maxBackups, maxAge int) error {
	// Parse log level
	logLevel, err := zerolog.ParseLevel(level)
	if err != nil {
		logLevel = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(logLevel)

	// Create log directory if it doesn't exist
	logDir := filepath.Dir(logPath)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return err
	}

	// Configure log rotation
	fileWriter := &lumberjack.Logger{
		Filename:   logPath,
		MaxSize:    maxSize,    // megabytes
		MaxBackups: maxBackups, // number of backups
		MaxAge:     maxAge,     // days
		Compress:   true,
	}

	// Console writer for development
	consoleWriter := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	}

	// Multi-writer for console and file
	multi := zerolog.MultiLevelWriter(consoleWriter, fileWriter)

	// Create logger
	Log = zerolog.New(multi).
		With().
		Timestamp().
		Caller().
		Logger()

	return nil
}

// GetLogger returns the global logger
func GetLogger() *zerolog.Logger {
	return &Log
}

// Info logs info message
func Info(msg string, args ...any) {
	if len(args) > 0 {
		Log.Info().Msgf(msg, args...)
	} else {
		Log.Info().Msg(msg)
	}
}

// Error logs error message
func Error(msg string, args ...any) {
	if len(args) > 0 {
		Log.Error().Msgf(msg, args...)
	} else {
		Log.Error().Msg(msg)
	}
}

// Debug logs debug message
func Debug(msg string, args ...any) {
	if len(args) > 0 {
		Log.Debug().Msgf(msg, args...)
	} else {
		Log.Debug().Msg(msg)
	}
}

// Warn logs warning message
func Warn(msg string, args ...any) {
	if len(args) > 0 {
		Log.Warn().Msgf(msg, args...)
	} else {
		Log.Warn().Msg(msg)
	}
}

// Fatal logs fatal message and exits
func Fatal(msg string, args ...any) {
	if len(args) > 0 {
		Log.Fatal().Msgf(msg, args...)
	} else {
		Log.Fatal().Msg(msg)
	}
}

// WithContext creates a child logger with context fields
func WithContext(fields map[string]any) *zerolog.Logger {
	logger := Log.With()
	for k, v := range fields {
		logger = logger.Interface(k, v)
	}
	l := logger.Logger()
	return &l
}
