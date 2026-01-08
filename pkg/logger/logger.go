package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var globalLogger *zap.Logger

// GetLogger returns the global logger
func GetLogger(verbose bool) (*zap.Logger, error) {
	if globalLogger != nil {
		return globalLogger, nil
	}

	// Configure logger based on verbosity
	var config zap.Config
	if verbose {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	} else {
		config = zap.NewProductionConfig()
		config.EncoderConfig.TimeKey = "timestamp"
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	}

	// Build logger
	logger, err := config.Build()
	if err != nil {
		return nil, err
	}

	globalLogger = logger
	return logger, nil
}

// Sync syncs the global logger
func Sync() error {
	if globalLogger != nil {
		return globalLogger.Sync()
	}
	return nil
}

// Init initializes the global logger with default settings
func Init(verbose bool) error {
	logger, err := GetLogger(verbose)
	if err != nil {
		return err
	}

	// Replace zap global logger
	zap.ReplaceGlobals(logger)
	return nil
}

// Default initializes logger with verbose=false and panics on error
func Default() {
	logger, err := GetLogger(false)
	if err != nil {
		panic(err)
	}
	zap.ReplaceGlobals(logger)
}

// Verbose initializes logger with verbose=true and panics on error
func Verbose() {
	logger, err := GetLogger(true)
	if err != nil {
		panic(err)
	}
	zap.ReplaceGlobals(logger)
}

// Production initializes logger for production use
func Production() error {
	logger, err := zap.NewProduction()
	if err != nil {
		return err
	}
	globalLogger = logger
	zap.ReplaceGlobals(logger)
	return nil
}

// Development initializes logger for development use
func Development() error {
	logger, err := zap.NewDevelopment()
	if err != nil {
		return err
	}
	globalLogger = logger
	zap.ReplaceGlobals(logger)
	return nil
}

// WithEnv initializes logger based on environment variables
func WithEnv() error {
	// Check if VERBOSE env var is set
	verbose := os.Getenv("VERBOSE") == "true" || os.Getenv("VERBOSE") == "1"

	logger, err := GetLogger(verbose)
	if err != nil {
		return err
	}
	zap.ReplaceGlobals(logger)
	return nil
}
