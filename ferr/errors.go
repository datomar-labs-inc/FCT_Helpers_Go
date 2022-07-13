package ferr

import (
	"fmt"
	"net/http"
)

var AccountExists = New(ETValidation, CodeAccountExists, "that account already exists").
	WithHTTPCode(http.StatusBadRequest)

var Unauthenticated = New(ETAuth, CodeNotAuthenticated, "no valid authentication was found").
	WithHTTPCode(http.StatusUnauthorized)

var AccountDisabled = New(ETPermissions, CodeAccountDisabled, "this account is disabled").
	WithHTTPCode(http.StatusForbidden)

var InvalidLoginDetails = New(ETAuth, CodeInvalidLoginDetails, "your login details were incorrect").
	WithHTTPCode(http.StatusBadRequest)

var MissingPermissions = New(ETPermissions, CodeMissingPermissions, "you do not have the required permissions for this action").
	WithHTTPCode(http.StatusForbidden)

var MissingArgument = func(argName string) *Error {
	return New(ETValidation, CodeMissingArgument, fmt.Sprintf("missing required argument: %s", argName)).
		WithHTTPCode(http.StatusBadRequest)
}

var ResourceTimedOut = func(timedOutResource string) *Error {
	return New(ETGeneric, CodeTimeout, fmt.Sprintf("resource timed out: %s", timedOutResource)).WithHTTPCode(http.StatusInternalServerError)
}

var InvalidArgument = func(resourceType string, reason ...string) error {
	err := New(ETValidation, CodeInvalidInput, fmt.Sprintf("invalid argument: %s, with reasons: %+v", resourceType, reason)).
		WithHTTPCode(http.StatusBadRequest)

	err.ResourceType = &resourceType
	err.Detail = reason

	return WrapWithOffset(err, 2)
}

var NotFound = func(resourceType string, detail ...string) error {
	err := New(ETGeneric, CodeNotFound, fmt.Sprintf("could not locate resource: %s, with ids: %+v", resourceType, detail)).
		WithHTTPCode(http.StatusNotFound)

	err.ResourceType = &resourceType
	err.Detail = detail

	return WrapWithOffset(err, 2)
}

var InternalServerError = func(resourceType string, detail ...string) error {
	err := New(ETGeneric, CodeUnknown, fmt.Sprintf("internal server error: %s, detail: %+v", resourceType, detail)).
		WithHTTPCode(http.StatusInternalServerError)

	err.ResourceType = &resourceType
	err.Detail = detail

	return WrapWithOffset(err, 2)
}
