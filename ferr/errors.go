package ferr

import "net/http"

var AccountExists = New(ETValidation, CodeAccountExists, "that account already exists").
	WithHTTPCode(http.StatusBadRequest)

var Unauthenticated = New(ETAuth, CodeNotAuthenticated, "no valid authentication was found").
	WithHTTPCode(http.StatusUnauthorized)

var InvalidLoginDetails = New(ETAuth, CodeInvalidLoginDetails, "your login details were incorrect").
	WithHTTPCode(http.StatusBadRequest)

var MissingPermissions = New(ETPermissions, CodeMissingPermissions, "you do not have the required permissions for this action").
	WithHTTPCode(http.StatusForbidden)
