// Package logger provides logging interfaces and implementations for the Lantern extractor.
package logger

import "fmt"

// Verbosity represents the logging verbosity level.
type Verbosity int

const (
	// VerbosityInfo logs all messages including info.
	VerbosityInfo Verbosity = iota
	// VerbosityWarning logs warning and error messages.
	VerbosityWarning
	// VerbosityError only logs error messages.
	VerbosityError
)

// Logger defines the interface for logging operations.
// Implementations can output to console, files, or discard messages entirely.
type Logger interface {
	// GetVerbosity returns the current verbosity level.
	GetVerbosity() Verbosity

	// SetVerbosity sets the minimum severity level that will be logged.
	SetVerbosity(verbosity Verbosity)

	// LogInfo logs an informational message.
	// The message will only be output if verbosity is set to VerbosityInfo.
	LogInfo(message string)

	// LogWarning logs a warning message.
	// The message will only be output if verbosity is set to VerbosityInfo or VerbosityWarning.
	LogWarning(message string)

	// LogError logs an error message.
	// Error messages are always output regardless of verbosity level.
	LogError(message string)
}

// NullLogger is a logger implementation that discards all messages.
type NullLogger struct {
	verbosity Verbosity
}

// NewNullLogger creates a new NullLogger.
func NewNullLogger() *NullLogger {
	return &NullLogger{verbosity: VerbosityError}
}

func (n *NullLogger) GetVerbosity() Verbosity          { return n.verbosity }
func (n *NullLogger) SetVerbosity(verbosity Verbosity) { n.verbosity = verbosity }
func (n *NullLogger) LogInfo(message string)           {}
func (n *NullLogger) LogError(message string)          {}
func (n *NullLogger) LogWarning(message string)        {}

// Ensure NullLogger implements Logger interface.
var _ Logger = (*NullLogger)(nil)

// ConsoleLogger logs messages to stdout/stderr.
type ConsoleLogger struct {
	verbosity Verbosity
}

// NewConsoleLogger creates a new ConsoleLogger with the specified verbosity.
func NewConsoleLogger(verbosity Verbosity) *ConsoleLogger {
	return &ConsoleLogger{verbosity: verbosity}
}

func (c *ConsoleLogger) GetVerbosity() Verbosity {
	return c.verbosity
}

func (c *ConsoleLogger) SetVerbosity(verbosity Verbosity) {
	c.verbosity = verbosity
}

func (c *ConsoleLogger) LogInfo(message string) {
	if c.verbosity <= VerbosityInfo {
		fmt.Println("[INFO]", message)
	}
}

func (c *ConsoleLogger) LogError(message string) {
	fmt.Println("[ERROR]", message)
}

func (c *ConsoleLogger) LogWarning(message string) {
	if c.verbosity <= VerbosityWarning {
		fmt.Println("[WARN]", message)
	}
}

// Ensure ConsoleLogger implements Logger interface.
var _ Logger = (*ConsoleLogger)(nil)
