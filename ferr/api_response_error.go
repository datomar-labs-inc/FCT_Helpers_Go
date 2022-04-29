package ferr

import "fmt"

type APIError struct {
	Type   string `json:"type"`
	Code   Code   `json:"code"`
	Detail string `json:"detail"`
}

func (a APIError) Error() string {
	return fmt.Sprintf("%s %s %s", a.Type, a.Code, a.Detail)
}

type APIValidationError struct {
	APIError
	Fields []*FieldError `json:"fields"`
}

type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}
