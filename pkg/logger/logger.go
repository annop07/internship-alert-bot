package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

var (
	infoLogger  *log.Logger
	errorLogger *log.Logger
	logFile     *os.File
)

// Init initializes the logger with file output
func Init() error {
	// Create logs directory if it doesn't exist
	logDir := "logs"
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	// Create log file with timestamp
	logFileName := filepath.Join(logDir, "bot.log")
	var err error
	logFile, err = os.OpenFile(logFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	// Create multi-writer (console + file)
	infoLogger = log.New(logFile, "INFO: ", log.Ldate|log.Ltime)
	errorLogger = log.New(logFile, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)

	Info("Logger initialized successfully")
	return nil
}

// Close closes the log file
func Close() {
	if logFile != nil {
		Info("Closing logger...")
		logFile.Close()
	}
}

// Info logs an informational message
func Info(format string, v ...interface{}) {
	message := fmt.Sprintf(format, v...)
	log.Println(message) // Console
	if infoLogger != nil {
		infoLogger.Println(message) // File
	}
}

// Error logs an error message
func Error(format string, v ...interface{}) {
	message := fmt.Sprintf(format, v...)
	log.Printf("❌ ERROR: %s", message) // Console with emoji
	if errorLogger != nil {
		errorLogger.Println(message) // File
	}
}

// Warn logs a warning message
func Warn(format string, v ...interface{}) {
	message := fmt.Sprintf(format, v...)
	log.Printf("⚠️  WARNING: %s", message) // Console
	if infoLogger != nil {
		infoLogger.Printf("WARN: %s", message) // File
	}
}

// Fatal logs a fatal error and exits
func Fatal(format string, v ...interface{}) {
	message := fmt.Sprintf(format, v...)
	log.Printf("💥 FATAL: %s", message)
	if errorLogger != nil {
		errorLogger.Printf("FATAL: %s", message)
	}
	Close()
	os.Exit(1)
}

// Success logs a success message
func Success(format string, v ...interface{}) {
	message := fmt.Sprintf(format, v...)
	log.Printf("✅ %s", message)
	if infoLogger != nil {
		infoLogger.Printf("SUCCESS: %s", message)
	}
}

// LogSeparator logs a visual separator
func LogSeparator() {
	separator := "=================================================="
	log.Println(separator)
	if infoLogger != nil {
		infoLogger.Println(separator)
	}
}

// LogSection logs a section header
func LogSection(title string) {
	LogSeparator()
	Info(title)
	LogSeparator()
}

// LogScheduledRun logs a scheduled run timestamp
func LogScheduledRun() {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	LogSeparator()
	Info("⏰ Scheduled run at %s", timestamp)
	LogSeparator()
}
