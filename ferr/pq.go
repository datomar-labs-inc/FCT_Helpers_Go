package ferr

import (
	"fmt"
	"github.com/friendsofgo/errors"
	"github.com/lib/pq"
	"net/http"
	"strings"
)

// ExtractPQError will attempt to turn an error into a pq.Error
// Returns nil if the error is not a pq.Error
func ExtractPQError(err error) *pq.Error {
	cause := errors.Cause(err)

	if pqErr, ok := cause.(*pq.Error); ok {
		return pqErr
	}

	return nil
}

// RetryFromPQError extracts retry information from a postgres error
// this function will decide if the error is retry-able, and for how long it should wait before retrying
func RetryFromPQError(err *pq.Error) *ErrorRetryInfo {
	switch err.Code.Name() {
	default:
		return &ErrorRetryInfo{
			ShouldRetry: true,
			WaitTimeMS:  250,
		}
	}
}

// HTTPCodeFromPQError extracts an http code from a postgres error
func HTTPCodeFromPQError(err *pq.Error) *int {
	code := http.StatusInternalServerError

	switch err.Code.Name() {
	case "unique_violation", "not_null_violation", "integrity_constraint_violation":
		code = http.StatusBadRequest
	default:
		code = http.StatusInternalServerError
	}

	return &code
}

// HandlePostgresError will attempt to convert dbErr into a postgres error, and then reason about it
func HandlePostgresError(dbErr error) error {
	pqErr := ExtractPQError(dbErr)

	if pqErr == nil {
		return dbErr
	}

	// For the year triggers
	if pqErr.Message == "product_line_different_year" {
		return &Error{
			Message:         "Product Line has a different year than the Product.",
			Type:            ETValidation,
			Code:            CodeInvalidInput,
			ResourceType:    &pqErr.Table,
			Detail:          []string{pqErr.Message},
			HTTPCode:        HTTPCodeFromPQError(pqErr),
			Retry:           RetryFromPQError(pqErr),
			UnderlyingError: dbErr,
		}
	} else if strings.Contains(pqErr.Message, "invalid input syntax for type uuid") {
		br := http.StatusBadRequest

		return &Error{
			Message:      "Invalid ID Specified",
			Type:         ETValidation,
			Code:         CodeInvalidInput,
			ResourceType: &pqErr.Table,
			HTTPCode:     &br,
		}
	}

	if pqErr.Message == "create_validation_error" {
		br := http.StatusBadRequest

		return &Error{
			Message:      fmt.Sprintf("%s: %s", pqErr.Message, pqErr.Constraint),
			Type:         ETValidation,
			Code:         CodeInvalidInput,
			ResourceType: &pqErr.Table,
			HTTPCode:     &br,
		}
	}

	return &Error{
		Message:         fmt.Sprintf("(%s) %s", pqErr.Code, pqErr.Message),
		Type:            ETDatabase,
		Code:            CodeUnknown,
		ResourceType:    &pqErr.Table,
		HTTPCode:        HTTPCodeFromPQError(pqErr),
		Retry:           RetryFromPQError(pqErr),
		UnderlyingError: dbErr,
	}
}
