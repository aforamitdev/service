package web

import "github.com/pkg/errors"

// FieldError is used to indicate an error with a specific request field.
type FieldError struct {
	Field string `json:"field"`
	Error string `json:"error"`
}

// ErrorResponse is the form used for API response form failure in the API
type ErrorResponse struct {
	Error  string       `json:"error"`
	Fields []FieldError `json:"fields,omitempty"`
}

// Error is used to pass error during the request through the application with specific context

type Error struct {
	Err    error
	Status int
	Fields []FieldError
}

func NewRequestError(err error, status int) error {
	return &Error{err, status, nil}
}

// error implements the error interface, Is used the default message of the
// wrapped error . this is what will be show in the services log
func (err *Error) Error() string {
	return err.Err.Error()
}

type shutdown struct {
	Message string
}

func NewShutdownError(message string) error {
	return &shutdown{Message: message}
}

func (s *shutdown) Error() string {
	return s.Message
}

func IsShutdown(err error) bool {
	if _, ok := errors.Cause(err).(*shutdown); ok {
		return true
	}
	return false
}
