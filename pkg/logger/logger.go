package logger

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
)

var log = logrus.New()

// Init initializes the logger with the specified level
func Init(level string) error {
	// Parse log level
	logLevel, err := logrus.ParseLevel(level)
	if err != nil {
		return err
	}

	// Configure logger
	log.SetLevel(logLevel)
	log.SetOutput(os.Stdout)
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
		DisableColors:   false,
	})

	return nil
}

// Get returns the logger instance
func Get() *logrus.Logger {
	return log
}

// WithField adds a field to the log entry
func WithField(key string, value interface{}) *logrus.Entry {
	return log.WithField(key, value)
}

// WithFields adds multiple fields to the log entry
func WithFields(fields logrus.Fields) *logrus.Entry {
	return log.WithFields(fields)
}

// Debug logs a debug message
func Debug(format string, args ...interface{}) {
	if len(args) > 0 {
		log.Debug(fmt.Sprintf(format, args...))
	} else {
		log.Debug(format)
	}
}

// Info logs an info message
func Info(format string, args ...interface{}) {
	if len(args) > 0 {
		log.Info(fmt.Sprintf(format, args...))
	} else {
		log.Info(format)
	}
}

// Warn logs a warning message
func Warn(format string, args ...interface{}) {
	if len(args) > 0 {
		log.Warn(fmt.Sprintf(format, args...))
	} else {
		log.Warn(format)
	}
}

// Error logs an error message
func Error(format string, args ...interface{}) {
	if len(args) > 0 {
		log.Error(fmt.Sprintf(format, args...))
	} else {
		log.Error(format)
	}
}

// Fatal logs a fatal message and exits
func Fatal(format string, args ...interface{}) {
	if len(args) > 0 {
		log.Fatal(fmt.Sprintf(format, args...))
	} else {
		log.Fatal(format)
	}
} 