package ferr

import "net/http"

var UserExists = New(ETValidation, CodeUserExists, "that user already exists").
	WithHTTPCode(http.StatusBadRequest)

var InvalidLoginDetails = New(ETAuth, CodeInvalidLoginDetails, "your login details were incorrect").
	WithHTTPCode(http.StatusBadRequest)
