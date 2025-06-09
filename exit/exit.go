package exit

import "fmt"

const (
	CodeSuccess         = 0
	CodeGoFailingTest   = 1
	CodeGoBuildError    = 2
	CodeFlakeguardError = 3
)

// Error represents an error with a specific exit code
type Error struct {
	Code int
	Err  error
}

func (e *Error) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return fmt.Sprintf("exit code %d", e.Code)
}

func (e *Error) Unwrap() error {
	return e.Err
}

// New creates a new Error that can be used to exit the program with a specific code
func New(code int, err error) *Error {
	return &Error{Code: code, Err: err}
}

// GetCode returns the exit code of an error if it is an Error, otherwise it returns CodeFlakeguardError
func GetCode(err error) int {
	if exitErr, ok := err.(*Error); ok {
		return exitErr.Code
	}
	return CodeFlakeguardError
}
