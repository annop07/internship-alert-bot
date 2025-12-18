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

func Init() error {
	logDir := "logs"
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	logFileName := filepath.Join(logDir, "bot.log")
	var err error
	logFile, err = os.OpenFile(logFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	infoLogger = log.New(logFile, "INFO: ", log.Ldate|log.Ltime)
	errorLogger = log.New(logFile, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)

	Info("Logger initialized successfully")
	return nil
}

func Close() {
	if logFile != nil {
		Info("Closing logger...")
		logFile.Close()
	}
}

func Info(format string, v ...interface{}) {
	message := fmt.Sprintf(format, v...)
	log.Println(message)
	if infoLogger != nil {
		infoLogger.Println(message)
	}
}

func Error(format string, v ...interface{}) {
	message := fmt.Sprintf(format, v...)
	log.Printf("❌ ERROR: %s", message)
	if errorLogger != nil {
		errorLogger.Println(message)
	}
}

func Warn(format string, v ...interface{}) {
	message := fmt.Sprintf(format, v...)
	log.Printf("⚠️  WARNING: %s", message)
	if infoLogger != nil {
		infoLogger.Printf("WARN: %s", message)
	}
}

func Fatal(format string, v ...interface{}) {
	message := fmt.Sprintf(format, v...)
	log.Printf("💥 FATAL: %s", message)
	if errorLogger != nil {
		errorLogger.Printf("FATAL: %s", message)
	}
	Close()
	os.Exit(1)
}

func Success(format string, v ...interface{}) {
	message := fmt.Sprintf(format, v...)
	log.Printf("✅ %s", message)
	if infoLogger != nil {
		infoLogger.Printf("SUCCESS: %s", message)
	}
}

func LogSeparator() {
	separator := "=================================================="
	log.Println(separator)
	if infoLogger != nil {
		infoLogger.Println(separator)
	}
}

func LogSection(title string) {
	LogSeparator()
	Info(title)
	LogSeparator()
}

func LogScheduledRun() {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	LogSeparator()
	Info("⏰ Scheduled run at %s", timestamp)
	LogSeparator()
}
