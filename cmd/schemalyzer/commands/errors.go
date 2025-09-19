package commands

import "fmt"

// ExitError is an error that indicates the command should exit with a specific code
type ExitError struct {
	Code    int
	Message string
}

func (e *ExitError) Error() string {
	return e.Message
}

// NewExitError creates a new ExitError with the specified code and message
func NewExitError(code int, message string) *ExitError {
	return &ExitError{
		Code:    code,
		Message: message,
	}
}

// NewExitErrorf creates a new ExitError with formatted message
func NewExitErrorf(code int, format string, args ...interface{}) *ExitError {
	return &ExitError{
		Code:    code,
		Message: fmt.Sprintf(format, args...),
	}
}

const (
	// ExitCodeMismatch is returned when schemas don't match
	ExitCodeMismatch = 2
)