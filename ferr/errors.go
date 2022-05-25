package ferr

import (
	"fmt"
	"net/http"
)

var AccountExists = New(ETValidation, CodeAccountExists, "that account already exists").
	WithHTTPCode(http.StatusBadRequest)

var Unauthenticated = New(ETAuth, CodeNotAuthenticated, "no valid authentication was found").
	WithHTTPCode(http.StatusUnauthorized)

var InvalidLoginDetails = New(ETAuth, CodeInvalidLoginDetails, "your login details were incorrect").
	WithHTTPCode(http.StatusBadRequest)

var MissingPermissions = New(ETPermissions, CodeMissingPermissions, "you do not have the required permissions for this action").
	WithHTTPCode(http.StatusForbidden)

var MissingArgument = func(argName string) *Error {
	return New(ETValidation, CodeMissingArgument, fmt.Sprintf("missing required argument: %s", argName)).
		WithHTTPCode(http.StatusBadRequest)
}

var InvalidArgument = func(argName string) *Error {
	return New(ETValidation, CodeInvalidInput, fmt.Sprintf("invalid argument: %s", argName)).
		WithHTTPCode(http.StatusBadRequest)
}

var NotFound = func(resourceName string) *Error {
	return New(ETGeneric, CodeNotFound, fmt.Sprintf("could not locate resource: %s", resourceName)).
		WithHTTPCode(http.StatusNotFound)
}