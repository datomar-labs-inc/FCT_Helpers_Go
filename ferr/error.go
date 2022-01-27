package ferr

import "fmt"

type ErrorType = int

const (
	// ETGeneric is for errors that do not fall into any other category
	ETGeneric = ErrorType(iota)

	// ETValidation indicates that a piece of data is invalid, either user provided, or provided by the programmer
	ETValidation

	// ETNetwork Indicates an error that occurred due to a networking issue. eg. failed to connect, failed to resolve host, etc...
	ETNetwork

	// ETSystem An error has occured in the underlying system. eg. out of memory, disk full, etc...
	ETSystem

	// ETAuth an error occurred during authentication, the requester failed to authenticate
	ETAuth

	// ETThirdPartySystem Any error originating from a system not controlled by Datomar (database is not a third party system)
	ETThirdPartySystem
)

type FCTError struct {
	Message string `json:"message"`

	// Type is a broad error category
	Type ErrorType `json:"type"`

	// Code is a machine-readable code that provides specific information about the error, see codes.go
	Code Code `json:"code"`

	// HTTPCode is an optional code that indicates what http status code this error represents
	HTTPCode *int `json:"http_code,omitempty"`

	// Retry is an optional
	Retry *ErrorRetryInfo `json:"retry,omitempty"`

	// UnderlyingError is any error that is being wrapped by this FCTError
	UnderlyingError error `json:"underlying_error,omitempty"`
}

// New creates a new FCTError with a message, code, and type
func New(eType ErrorType, code Code, msg string) *FCTError {
	return &FCTError{
		Message: msg,
		Type:    eType,
		Code:    code,
	}
}

// Wrap creates a new FCTError out of any go error, this should be used sparingly, and most errors should
// be converted into a full FCTError so that the system can identify specifically what is wrong
func Wrap(err error) *FCTError {
	return &FCTError{
		Message:         err.Error(),
		Code:            CodeWrapped,
		UnderlyingError: err,
	}
}

// ErrorRetryInfo contains information that clients should use to determine if they should retry a request
// and how long they should wait before retrying a request
type ErrorRetryInfo struct {
	ShouldRetry bool `json:"should_retry"`
	WaitTimeMS  int  `json:"wait_time_ms"`
}

// WithUnderlying will attach any go error to the FCTError. This indicates that err is the cause of the FCTError
func (f *FCTError) WithUnderlying(err error) *FCTError {
	f.UnderlyingError = err

	return f
}

// WithRetry will attach retry information, indicating that the upstream caller can retry this call after waitTime
func (f *FCTError) WithRetry(waitTime int) *FCTError {
	f.Retry = &ErrorRetryInfo{
		ShouldRetry: true,
		WaitTimeMS:  waitTime,
	}

	return f
}

// WithHTTPCode will attach a http status code to this FCTError, this is used by the upstream caller to set the
// http response status code
func (f *FCTError) WithHTTPCode(code int) *FCTError {
	f.HTTPCode = &code

	return f
}

func (f *FCTError) Error() string {
	return fmt.Sprintf("(%d-%d) %s: %v", f.Code, f.Type, f.Message, f.UnderlyingError)
}

func (f *FCTError) Unwrap() error {
	return f.UnderlyingError
}
