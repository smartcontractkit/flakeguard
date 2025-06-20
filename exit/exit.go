// Package exit provides a way to exit the program with a specific exit code.
// This helps identify what kind of error occurred and what to do about it.
// If you notice that a Go build is failing, you'll likely want to directly fail flakeguard and bubble the error up.
// But if the error is a test failure, based on the context, you might not want to fail flakeguard.
package exit

import "fmt"

const (
	// CodeSuccess is the exit code for a successful go test run.
	CodeSuccess = 0
	// CodeGoFailingTest is the exit code go test returns when a test fails (along with a bunch of other issues at test runtime).
	CodeGoFailingTest = 1
	// CodeGoBuildError is the exit code for a build error when running go test.
	CodeGoBuildError = 2
	// CodeFlakeguardError is the exit code for a flakeguard specific error.
	CodeFlakeguardError = 3
)

// Error represents an error with a specific exit code
type Error struct {
	Code int
	Err  error
}

// Error returns a string representation of the error.
func (e *Error) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return fmt.Sprintf("exit code %d", e.Code)
}

// ExitCode returns the exit code of the error.
func (e *Error) ExitCode() int {
	return e.Code
}

// Unwrap returns the wrapped error.
func (e *Error) Unwrap() error {
	return e.Err
}

// New creates a new Error that can be used to exit the program with a specific code
func New(code int, err error) *Error {
	return &Error{Code: code, Err: err}
}

// GetCode returns the exit code of an error if it is an Error, otherwise it returns CodeFlakeguardError
func GetCode(err error) int {
	if err == nil {
		return CodeSuccess
	}
	if exitErr, ok := err.(*Error); ok {
		return exitErr.Code
	}
	return CodeFlakeguardError
}
