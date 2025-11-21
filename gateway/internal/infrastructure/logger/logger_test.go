package logger

import (
	"os"
	"testing"

	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/logger"
)

func TestLoggerOutput(t *testing.T) {

	os.Remove("test.log")

	config := &LoggerConfig{
		LogLevel:      "debug",
		ConsoleOutput: "true",
		LogFile:       "test.log",
	}

	log, err := NewLogger(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer log.Close()

	t.Log("\n=== Testing Basic Logging ===")
	log.Debug("This is a debug message", logger.Field{Key: "component", Value: "auth"})
	log.Info("Application started successfully", logger.Field{Key: "version", Value: "1.0.0"})
	log.Warn("Connection pool running low", logger.Field{Key: "available", Value: 5}, logger.Field{Key: "total", Value: 20})
	log.Error("Failed to connect to database", logger.Field{Key: "error", Value: "connection timeout"}, logger.Field{Key: "retry_count", Value: 3})

	// Test WithFields
	t.Log("\n=== Testing WithFields ===")
	requestLogger := log.WithFields(map[string]interface{}{
		"request_id": "req-12345",
		"user_id":    42,
		"endpoint":   "/api/users",
	})
	requestLogger.Info("Processing user request")
	requestLogger.Warn("Rate limit approaching", logger.Field{Key: "requests_per_minute", Value: 95})

	// Test contextual logging
	t.Log("\n=== Testing Contextual Logging ===")
	dbLogger := log.WithFields(map[string]interface{}{
		"service": "database",
		"host":    "localhost",
		"port":    5432,
	})
	dbLogger.Info("Database connection established")
	dbLogger.Debug("Running migration", logger.Field{Key: "table", Value: "users"})

	// Multiple fields
	t.Log("\n=== Testing Multiple Fields ===")
	log.Info("User authentication",
		logger.Field{Key: "user_id", Value: 123},
		logger.Field{Key: "method", Value: "JWT"},
		logger.Field{Key: "ip_address", Value: "192.168.1.1"},
		logger.Field{Key: "status", Value: "success"},
	)

	if content, err := os.ReadFile("test.log"); err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	} else if len(content) == 0 {
		t.Fatal("Log file is empty")
	} else {
		t.Logf("Log file size: %d bytes\n", len(content))
		t.Log("\n=== Log File Content ===")
		t.Log(string(content))
	}

	// Clean up
	os.Remove("test.log")
}

func TestLoggerWithDifferentLevels(t *testing.T) {
	t.Log("\n=== Testing Different Log Levels ===")

	levels := []string{"debug", "info", "warn", "error"}

	for _, level := range levels {
		t.Logf("\n--- Log Level: %s ---", level)
		config := &LoggerConfig{
			LogLevel:      level,
			ConsoleOutput: "true",
			LogFile:       "",
		}

		log, err := NewLogger(config)
		if err != nil {
			t.Fatalf("Failed to create logger for level %s: %v", level, err)
		}

		log.Debug("Debug message")
		log.Info("Info message")
		log.Warn("Warning message")
		log.Error("Error message")

		log.Close()
	}
}

func TestLoggerFileOutput(t *testing.T) {
	t.Log("\n=== Testing File Logging ===")

	config := &LoggerConfig{
		LogLevel:      "info",
		ConsoleOutput: "false",
		LogFile:       "test_file.log",
	}

	log, err := NewLogger(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	for i := 0; i < 10; i++ {
		log.Info("Batch message",
			logger.Field{Key: "batch_num", Value: i},
			logger.Field{Key: "operation", Value: "test"},
		)
	}

	log.Close()

	if content, err := os.ReadFile("test_file.log"); err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	} else {
		t.Logf("File logging successful - file size: %d bytes", len(content))
	}

	os.Remove("test_file.log")
}
