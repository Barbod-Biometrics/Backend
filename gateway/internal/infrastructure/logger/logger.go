package logger

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/logger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type LoggerConfig struct {
	LogLevel      string
	ConsoleOutput string
	LogFile       string
}

type Logger struct {
	zap *zap.Logger
}

var loggerInstance logger.Logger


func NewLogger(config *LoggerConfig) (*Logger, error) {
	if config == nil {
		return nil, fmt.Errorf("logger config cannot be nil")
	}

	if config.LogLevel == "" {
		config.LogLevel = "info"
	}

	var level zapcore.Level
	switch config.LogLevel {
	case "debug":
		level = zapcore.DebugLevel
	case "info":
		level = zapcore.InfoLevel
	case "warn":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	case "fatal":
		level = zapcore.FatalLevel
	default:
		level = zapcore.InfoLevel
	}

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	encoder := zapcore.NewJSONEncoder(encoderConfig)

	var cores []zapcore.Core

	// Console output
	consoleOutput, err := strconv.ParseBool(config.ConsoleOutput)
	if err != nil {
		consoleOutput = true
		log.Println("Error parsing console output setting. Defaulting to true")
	}
	if consoleOutput {
		consoleCore := zapcore.NewCore(
			encoder,
			zapcore.AddSync(os.Stdout),
			level,
		)
		cores = append(cores, consoleCore)
	}

	// File output
	if config.LogFile != "" {
		file, err := os.OpenFile(config.LogFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file: %w", err)
		}
		fileCore := zapcore.NewCore(
			encoder,
			zapcore.AddSync(file),
			level,
		)
		cores = append(cores, fileCore)
	}

	// Create composite core
	var core zapcore.Core
	if len(cores) == 0 {
		return nil, fmt.Errorf("at least one output must be configured")
	} else if len(cores) == 1 {
		core = cores[0]
	} else {
		core = zapcore.NewTee(cores...)
	}

	zapLogger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	logger := &Logger{
		zap: zapLogger,
	}

	loggerInstance = logger
	return logger, nil
}

func GetLogger() logger.Logger {
	if loggerInstance == nil {
		loggerInstance = &Logger{
			zap: zap.NewExample(),
		}
	}
	return loggerInstance
}

func toZapFields(fields []logger.Field) []zap.Field {
	zapFields := make([]zap.Field, len(fields))
	for i, field := range fields {
		zapFields[i] = zap.Any(field.Key, field.Value)
	}
	return zapFields
}

func (l *Logger) Debug(msg string, fields ...logger.Field) {
	l.zap.Debug(msg, toZapFields(fields)...)
}

func (l *Logger) Info(msg string, fields ...logger.Field) {
	l.zap.Info(msg, toZapFields(fields)...)
}

func (l *Logger) Warn(msg string, fields ...logger.Field) {
	l.zap.Warn(msg, toZapFields(fields)...)
}

func (l *Logger) Error(msg string, fields ...logger.Field) {
	l.zap.Error(msg, toZapFields(fields)...)
}

func (l *Logger) Fatal(msg string, fields ...logger.Field) {
	l.zap.Fatal(msg, toZapFields(fields)...)
}

func (l *Logger) WithFields(fields map[string]interface{}) logger.Logger {
	zapFields := make([]zap.Field, 0, len(fields))
	for k, v := range fields {
		zapFields = append(zapFields, zap.Any(k, v))
	}

	return &Logger{
		zap: l.zap.With(zapFields...),
	}
}

func (l *Logger) Close() {
	l.zap.Sync()
}
