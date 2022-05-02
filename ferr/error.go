package ferr

import (
	"errors"
	"fmt"
	"github.com/datomar-labs-inc/FCT_Helpers_Go/ferr/valid"
	"github.com/gofiber/fiber/v2"
	"github.com/iancoleman/strcase"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/temporal"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
	"strings"
)

type Error struct {
	Message string `json:"message"`

	// Type is a broad error category
	Type ErrorType `json:"type"`

	// Code is a machine-readable code that provides specific information about the error, see codes.go
	Code Code `json:"code"`

	// Source is a developer readable string that indicates where the error originated
	Source string `json:"source"`

	// HTTPCode is an optional code that indicates what http status code this error represents
	HTTPCode *int `json:"http_code,omitempty"`

	// Retry is an optional
	Retry *ErrorRetryInfo `json:"retry,omitempty"`

	// UnderlyingError is any error that is being wrapped by this Error
	UnderlyingError error `json:"underlying_error,omitempty"`

	Fields []*FieldError `json:"fields,omitempty"`
}

// New creates a new Error with a message, code, and type
func New(eType ErrorType, code Code, msg string) *Error {
	return &Error{
		Message: msg,
		Type:    eType,
		Code:    code,
	}
}

// Infer will attempt to smartly extract error information into an Error
func Infer(err error) *Error {
	if err == nil {
		return nil
	}

	if fctErr, ok := err.(*Error); ok {
		return fctErr
	}

	// Check if it's a postgres error
	pqErr := ExtractPQError(err)

	if pqErr != nil {
		return &Error{
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

	// Check for unmarshal errors
	var unmarshalTypeError *fiber.UnmarshalTypeError
	if errors.As(err, &unmarshalTypeError) {
		return New(ETValidation, CodeInvalidInput, "invalid input").
			WithHTTPCode(http.StatusBadRequest).
			WithFieldError(&FieldError{
				Field:   strcase.ToSnakeWithIgnore(unmarshalTypeError.Field, "."),
				Message: unmarshalTypeError.Error(),
			})
	}

	var validationError validator.ValidationErrors
	if errors.As(err, &validationError) {
		fe := New(ETValidation, CodeInvalidInput, "invalid input").
			WithHTTPCode(http.StatusBadRequest)

		translated := validationError.Translate(valid.UniversalTranslator)

		for field, message := range translated {
			// field starts with the struct name, followed by a dot, so it should be removed
			field = strcase.ToSnakeWithIgnore(strings.Join(strings.Split(field, ".")[1:], "."), ".")

			fe = fe.WithFieldError(&FieldError{
				Field:   field,
				Message: message,
			})
		}

		return fe
	}

	return &Error{
		Message:         err.Error(),
		Type:            ETGeneric,
		Code:            CodeUnknown,
		UnderlyingError: err,
	}
}

// InferAPIError will attempt to smrtley extract error information into an Error
func InferAPIError(err error) error {
	if err == nil {
		return nil
	}

	if fctErr, ok := err.(*Error); ok {
		return fctErr
	}

	// Check if it's a postgres error
	pqErr := ExtractPQError(err)

	if pqErr != nil {

		return &Error{
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

	// Check for unmarshal errors
	var unmarshalTypeError *fiber.UnmarshalTypeError
	if errors.As(err, &unmarshalTypeError) {
		return New(ETValidation, CodeInvalidInput, "invalid input").
			WithHTTPCode(http.StatusBadRequest).
			WithFieldError(&FieldError{
				Field:   strcase.ToSnakeWithIgnore(unmarshalTypeError.Field, "."),
				Message: unmarshalTypeError.Error(),
			})
	}

	var validationError validator.ValidationErrors
	if errors.As(err, &validationError) {
		fe := New(ETValidation, CodeInvalidInput, "invalid input").
			WithHTTPCode(http.StatusBadRequest)

		translated := validationError.Translate(valid.UniversalTranslator)

		for field, message := range translated {
			// field starts with the struct name, followed by a dot, so it should be removed
			field = strcase.ToSnakeWithIgnore(strings.Join(strings.Split(field, ".")[1:], "."), ".")

			fe = fe.WithFieldError(&FieldError{
				Field:   field,
				Message: message,
			})
		}

		return fe
	}

	return &Error{
		Message:         err.Error(),
		Type:            ETGeneric,
		Code:            CodeUnknown,
		UnderlyingError: err,
	}
}

// Wrap creates a new Error out of any go error, this should be used sparingly, and most errors should
// be converted into a full Error so that the system can identify specifically what is wrong
func Wrap(err error) *Error {
	return &Error{
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

// WithUnderlying will attach any go error to the Error. This indicates that err is the cause of the Error
func (f *Error) WithUnderlying(err error) *Error {
	f.UnderlyingError = err

	return f
}

// WithRetry will attach retry information, indicating that the upstream caller can retry this call after waitTime
func (f *Error) WithRetry(waitTime int) *Error {
	f.Retry = &ErrorRetryInfo{
		ShouldRetry: true,
		WaitTimeMS:  waitTime,
	}

	return f
}

// WithHTTPCode will attach a http status code to this Error, this is used by the upstream caller to set the
// http response status code
func (f *Error) WithHTTPCode(code int) *Error {
	f.HTTPCode = &code

	return f
}

func (f *Error) WithFieldError(ferr *FieldError) *Error {
	f.Fields = append(f.Fields, ferr)
	return f
}

func (f *Error) Error() string {
	return fmt.Sprintf("(%d-%d) %s: %v", f.Code, f.Type, f.Message, f.UnderlyingError)
}

func (f *Error) Unwrap() error {
	return f.UnderlyingError
}

func (f *Error) ToAPIResponseError() error {
	switch f.Type {
	case ETGeneric:
		return getAPIError(ETGeneric, f.Code, f.Message)
	}

	switch f.Code {
	case CodeInvalidInput:
		return &APIValidationError{
			APIError: APIError{
				Type:   "",
				Code:   f.Code,
				Detail: f.Message,
			},
			Fields: f.Fields,
		}
	default:
		return &APIError{
			Type:   "",
			Code:   f.Code,
			Detail: f.Message,
		}
	}
}

func getAPIError(errorType ErrorType, code Code, message string) error {
	return APIValidationError{
		APIError: APIError{
			Type:   errorType.String(),
			Code:   code,
			Detail: message,
		},
	}
}
