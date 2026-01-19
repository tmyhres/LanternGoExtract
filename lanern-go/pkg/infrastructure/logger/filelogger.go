package logger

import (
	"bufio"
	"fmt"
	"os"
	"sync"
)

// FileLogger outputs log messages to a text file.
type FileLogger struct {
	verbosity Verbosity
	file      *os.File
	writer    *bufio.Writer
	mu        sync.Mutex
}

// NewFileLogger creates a new FileLogger that writes to the specified file path.
// Returns an error if the file cannot be created or opened.
func NewFileLogger(logFilePath string, verbosity Verbosity) (*FileLogger, error) {
	file, err := os.Create(logFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create log file: %w", err)
	}

	return &FileLogger{
		verbosity: verbosity,
		file:      file,
		writer:    bufio.NewWriter(file),
	}, nil
}

// GetVerbosity returns the current verbosity level.
func (l *FileLogger) GetVerbosity() Verbosity {
	return l.verbosity
}

// SetVerbosity sets the minimum severity level that will be logged.
func (l *FileLogger) SetVerbosity(verbosity Verbosity) {
	l.verbosity = verbosity
}

// LogInfo logs an informational message to the file.
func (l *FileLogger) LogInfo(message string) {
	if l.verbosity < VerbosityInfo {
		return
	}
	l.writeLine("<INFO> " + message)
}

// LogWarning logs a warning message to the file.
func (l *FileLogger) LogWarning(message string) {
	if l.verbosity < VerbosityWarning {
		return
	}
	l.writeLine("<WARN> " + message)
}

// LogError logs an error message to the file.
func (l *FileLogger) LogError(message string) {
	l.writeLine("<ERROR> " + message)
}

// writeLine writes a line to the log file with thread safety.
func (l *FileLogger) writeLine(line string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.writer.WriteString(line + "\n")
	l.writer.Flush()
}

// Close closes the underlying file. Should be called when logging is complete.
func (l *FileLogger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if err := l.writer.Flush(); err != nil {
		return fmt.Errorf("failed to flush log buffer: %w", err)
	}
	if err := l.file.Close(); err != nil {
		return fmt.Errorf("failed to close log file: %w", err)
	}
	return nil
}

// Ensure FileLogger implements Logger interface.
var _ Logger = (*FileLogger)(nil)
