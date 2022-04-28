package ferr

type APIResponseError struct {
	Message string        `json:"message"`
	Code    Code          `json:"code"`
	Fields  []*FieldError `json:"fields,omitempty"`
}

type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}
