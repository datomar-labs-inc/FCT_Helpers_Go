package ferr

import "fmt"

type APIErrorResponse interface {
	Error() string
	GetBaseError() *APIError
	GetFullError() error
}

type APIError struct {
	Type    string        `json:"type"`
	Code    Code          `json:"code"`
	Detail  string        `json:"detail"`
	Summary *ErrorSummary `json:"summary,omitempty"`
}

func (a *APIError) GetBaseError() *APIError {
	return a
}

func (a *APIError) GetFullError() error {
	return a
}

func (a *APIError) Error() string {
	return fmt.Sprintf("%s %s %s", a.Type, a.Code, a.Detail)
}

type APIValidationError struct {
	*APIError
	Fields []*FieldError `json:"fields"`
}

func (ave *APIValidationError) GetFullError() error {
	return ave
}

type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}
