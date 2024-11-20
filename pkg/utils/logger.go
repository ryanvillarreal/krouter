package utils

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Logger wraps zap logger
type Logger struct {
	*zap.Logger
	*zap.SugaredLogger
}

// Config holds the logger configuration
type Config struct {
	// LogLevel defines the minimum log level for output
	LogLevel string `yaml:"level" json:"level"`
	// Outputs defines where the logs will be written (console, file, both)
	Outputs []string `yaml:"outputs" json:"outputs"`
	// FileLocation specifies where to write log files
	FileLocation string `yaml:"file_location" json:"file_location"`
	// MaxSize is the maximum size in megabytes of the log file before it gets rotated
	MaxSize int `yaml:"max_size" json:"max_size"`
	// MaxBackups is the maximum number of old log files to retain
	MaxBackups int `yaml:"max_backups" json:"max_backups"`
	// MaxAge is the maximum number of days to retain old log files
	MaxAge int `yaml:"max_age" json:"max_age"`
	// Compress determines if the rotated log files should be compressed
	Compress bool `yaml:"compress" json:"compress"`
	// Development puts the logger in development mode
	Development bool `yaml:"development" json:"development"`
}

// DefaultConfig returns a default configuration
func DefaultConfig() Config {
	return Config{
		LogLevel:     "info",
		Outputs:      []string{"console"},
		FileLocation: "logs/app.log",
		MaxSize:      100,
		MaxBackups:   3,
		MaxAge:       28,
		Compress:     true,
		Development:  false,
	}
}

// NewLogger creates a new logger instance with the given configuration
func NewLogger(config Config) (*Logger, error) {
	// Parse log level
	level, err := zapcore.ParseLevel(config.LogLevel)
	if err != nil {
		return nil, err
	}

	// Create encoder config
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalColorLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// Create cores for different outputs
	var cores []zapcore.Core

	// Console output
	if contains(config.Outputs, "console") {
		consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)
		consoleCore := zapcore.NewCore(
			consoleEncoder,
			zapcore.AddSync(os.Stdout),
			level,
		)
		cores = append(cores, consoleCore)
	}

	// File output
	if contains(config.Outputs, "file") {
		// Configure log rotation
		rotator := &lumberjack.Logger{
			Filename:   config.FileLocation,
			MaxSize:    config.MaxSize,
			MaxBackups: config.MaxBackups,
			MaxAge:     config.MaxAge,
			Compress:   config.Compress,
		}

		fileEncoder := zapcore.NewJSONEncoder(encoderConfig)
		fileCore := zapcore.NewCore(
			fileEncoder,
			zapcore.AddSync(rotator),
			level,
		)
		cores = append(cores, fileCore)
	}

	// Combine cores
	core := zapcore.NewTee(cores...)

	// Create logger
	logger := zap.New(
		core,
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
		zap.Development(config.Development),
	)

	return &Logger{
		Logger:        logger,
		SugaredLogger: logger.Sugar(),
	}, nil
}

// Helper function to check if a slice contains a string
func contains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

// Helper methods for structured logging
func (l *Logger) WithFields(fields map[string]interface{}) *zap.SugaredLogger {
	return l.SugaredLogger.With(fieldsToArgs(fields)...)
}

// Convert map to alternating key-value pairs for zap
func fieldsToArgs(fields map[string]interface{}) []interface{} {
	args := make([]interface{}, 0, len(fields)*2)
	for k, v := range fields {
		args = append(args, k, v)
	}
	return args
}
