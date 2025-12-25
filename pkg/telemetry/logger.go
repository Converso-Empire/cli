package telemetry

import (
	"io"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Logger interface for structured logging
type Logger interface {
	Debug(msg string, fields ...interface{})
	Info(msg string, fields ...interface{})
	Warn(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
	Fatal(msg string, fields ...interface{})
	Panic(msg string, fields ...interface{})
}

// ZerologAdapter adapts zerolog to our Logger interface
type ZerologAdapter struct {
	logger zerolog.Logger
}

// NewLogger creates a new structured logger
func NewLogger(debug bool) Logger {
	// Configure output
	var output io.Writer
	if debug {
		// Human-readable output for debugging
		output = zerolog.ConsoleWriter{
			Out:        os.Stderr,
			TimeFormat: "15:04:05",
		}
	} else {
		// JSON output for production
		output = os.Stderr
	}

	// Create logger
	logger := zerolog.New(output).With().Timestamp().Logger()

	// Set log level
	if debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		logger = logger.Level(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
		logger = logger.Level(zerolog.InfoLevel)
	}

	return &ZerologAdapter{logger: logger}
}

// Debug logs a debug message
func (l *ZerologAdapter) Debug(msg string, fields ...interface{}) {
	l.logEvent("debug", msg, fields...)
}

// Info logs an info message
func (l *ZerologAdapter) Info(msg string, fields ...interface{}) {
	l.logEvent("info", msg, fields...)
}

// Warn logs a warning message
func (l *ZerologAdapter) Warn(msg string, fields ...interface{}) {
	l.logEvent("warn", msg, fields...)
}

// Error logs an error message
func (l *ZerologAdapter) Error(msg string, fields ...interface{}) {
	l.logEvent("error", msg, fields...)
}

// Fatal logs a fatal message and exits
func (l *ZerologAdapter) Fatal(msg string, fields ...interface{}) {
	l.logEvent("fatal", msg, fields...)
	os.Exit(1)
}

// Panic logs a panic message and panics
func (l *ZerologAdapter) Panic(msg string, fields ...interface{}) {
	l.logEvent("panic", msg, fields...)
	panic(msg)
}

// logEvent logs an event with structured fields
func (l *ZerologAdapter) logEvent(level, msg string, fields ...interface{}) {
	event := l.logger.With().Logger()

	// Process fields in pairs (key, value)
	for i := 0; i < len(fields); i += 2 {
		if i+1 < len(fields) {
			key, ok := fields[i].(string)
			if ok {
				event = event.With().Interface(key, fields[i+1]).Logger()
			}
		}
	}

	// Log the message
	switch level {
	case "debug":
		event.Debug().Msg(msg)
	case "info":
		event.Info().Msg(msg)
	case "warn":
		event.Warn().Msg(msg)
	case "error":
		event.Error().Msg(msg)
	case "fatal":
		event.Fatal().Msg(msg)
	case "panic":
		event.Panic().Msg(msg)
	default:
		event.Info().Msg(msg)
	}
}

// Global logger instance
var (
	globalLogger Logger
)

// SetGlobalLogger sets the global logger
func SetGlobalLogger(logger Logger) {
	globalLogger = logger
}

// GetGlobalLogger returns the global logger
func GetGlobalLogger() Logger {
	if globalLogger == nil {
		globalLogger = NewLogger(false)
	}
	return globalLogger
}

// Debug logs a debug message using the global logger
func Debug(msg string, fields ...interface{}) {
	GetGlobalLogger().Debug(msg, fields...)
}

// Info logs an info message using the global logger
func Info(msg string, fields ...interface{}) {
	GetGlobalLogger().Info(msg, fields...)
}

// Warn logs a warning message using the global logger
func Warn(msg string, fields ...interface{}) {
	GetGlobalLogger().Warn(msg, fields...)
}

// Error logs an error message using the global logger
func Error(msg string, fields ...interface{}) {
	GetGlobalLogger().Error(msg, fields...)
}

// Fatal logs a fatal message using the global logger and exits
func Fatal(msg string, fields ...interface{}) {
	GetGlobalLogger().Fatal(msg, fields...)
}

// Panic logs a panic message using the global logger and panics
func Panic(msg string, fields ...interface{}) {
	GetGlobalLogger().Panic(msg, fields...)
}
