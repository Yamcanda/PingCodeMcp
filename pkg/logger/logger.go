package logger

import (
	"log"
	"os"
)

// Logger wraps the standard logger with additional functionality
type Logger struct {
	*log.Logger
}

// New creates a new logger instance
func New() *Logger {
	return &Logger{
		Logger: log.New(os.Stdout, "[PINGCODE-MCP] ", log.LstdFlags|log.Lshortfile),
	}
}

// Info logs an info message
func (l *Logger) Info(v ...interface{}) {
	l.SetPrefix("[PINGCODE-MCP] [INFO] ")
	l.Println(v...)
}

// Error logs an error message
func (l *Logger) Error(v ...interface{}) {
	l.SetPrefix("[PINGCODE-MCP] [ERROR] ")
	l.Println(v...)
}

// Debug logs a debug message
func (l *Logger) Debug(v ...interface{}) {
	l.SetPrefix("[PINGCODE-MCP] [DEBUG] ")
	l.Println(v...)
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(v ...interface{}) {
	l.SetPrefix("[PINGCODE-MCP] [FATAL] ")
	l.Fatal(v...)
}
