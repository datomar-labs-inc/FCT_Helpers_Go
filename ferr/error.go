package ferr

import (
	"errors"
	"fmt"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/temporal"
	"net/http"
)

type ErrorType = int

const (
	// ETGeneric is for errors that do not fall into any other category
	ETGeneric = ErrorType(iota)

	// ETValidation indicates that a piece of data is invalid, either user provided, or provided by the programmer
	ETValidation

	// ETNetwork Indicates an error that occurred due to a networking issue. eg. failed to connect, failed to resolve host, etc...
	ETNetwork

	// ETSystem An error has occurred in the underlying system. eg. out of memory, disk full, etc...
	ETSystem

	// ETTemporal an error has occurred in temporal
	ETTemporal

	// ETAuth an error occurred during authentication, the requester failed to authenticate
	ETAuth

	// ETDatabase an error occurred during a database operation
	ETDatabase

	// ETThirdPartySystem Any error originating from a system not controlled by Datomar (database is not a third party system)
	ETThirdPartySystem

	// ETPermissions an error caused by a user attempting to perform an action that they do not have permissions for
	ETPermissions
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

// Infer will attempt to smartly extract error information into an FCTError
func Infer(err error) *FCTError {
	if err == nil {
		return nil
	}

	if fctErr, ok := err.(*FCTError); ok {
		return fctErr
	}

	// Check if it's a postgres error
	pqErr := ExtractPQError(err)

	if pqErr != nil {
		return &FCTError{
			Message:         pqErr.Message,
			Type:            ETDatabase,
			Code:            CodeUnknown,
			HTTPCode:        HTTPCodeFromPQError(pqErr),
			Retry:           RetryFromPQError(pqErr),
			UnderlyingError: err,
		}
	}

	var applicationErr *temporal.ApplicationError
	if errors.As(err, &applicationErr) {
		unwrapped := applicationErr.Unwrap()

		if unwrapped != nil {
			return Infer(applicationErr.Unwrap())
		} else {
			return New(ETTemporal, CodeUnknown, applicationErr.Error()).WithUnderlying(applicationErr)
		}
	}

	var canceledErr *temporal.CanceledError
	if errors.As(err, &canceledErr) {
		return New(ETTemporal, CodeUnknown, canceledErr.Error()).WithUnderlying(canceledErr)
	}

	var timeoutErr *temporal.TimeoutError
	if errors.As(err, &timeoutErr) {

		switch timeoutErr.TimeoutType() {
		case enums.TIMEOUT_TYPE_SCHEDULE_TO_START, enums.TIMEOUT_TYPE_SCHEDULE_TO_CLOSE:
			return New(ETTemporal, CodeTimeout, timeoutErr.Error()).
				WithHTTPCode(http.StatusInternalServerError).
				WithUnderlying(Infer(timeoutErr.Unwrap()))
		case enums.TIMEOUT_TYPE_UNSPECIFIED, enums.TIMEOUT_TYPE_HEARTBEAT:
			return New(ETTemporal, CodeTimeout, timeoutErr.Error()).
				WithHTTPCode(http.StatusInternalServerError).
				WithUnderlying(Infer(timeoutErr.Unwrap()))
		case enums.TIMEOUT_TYPE_START_TO_CLOSE:
			return New(ETTemporal, CodeTimeout, timeoutErr.Error()).
				WithHTTPCode(http.StatusInternalServerError).
				WithUnderlying(Infer(timeoutErr.Unwrap()))
		default:
		}
	}

	var panicErr *temporal.PanicError
	if errors.As(err, &panicErr) {
		return New(ETTemporal, CodePanic, panicErr.Error()).
			WithHTTPCode(http.StatusInternalServerError).
			WithUnderlying(panicErr)
	}

	return &FCTError{
		Message:         err.Error(),
		Type:            ETGeneric,
		Code:            CodeUnknown,
		UnderlyingError: err,
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
