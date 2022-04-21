package ferr

import (
	"github.com/friendsofgo/errors"
	"github.com/lib/pq"
	"net/http"
)

// ExtractPQError will attempt to turn an error into a pq.Error
// Returns nil if the error is not a pq.Error
func ExtractPQError(err error) *pq.Error {
	cause := errors.Cause(err)

	if pqErr, ok := cause.(*pq.Error); ok {
		return pqErr
	} else {
		return nil
	}
}

// RetryFromPQError extracts retry information from a postgres error
// this function will decide if the error is retry-able, and for how long it should wait before retrying
func RetryFromPQError(err *pq.Error) *ErrorRetryInfo {
	// TODO add more logic

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
	// TODO add more logic

	code := http.StatusInternalServerError

	switch err.Code.Name() {
	default:
		code = http.StatusInternalServerError
	}

	return &code
}
