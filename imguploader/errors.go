package imguploader

import "fmt"

type InvalidFormatError struct {
	format string
}

func NewInvalidFormatError(format string) *InvalidFormatError {
	return &InvalidFormatError{
		format: format,
	}
}

func (i *InvalidFormatError) Error() string {
	return fmt.Sprintf("invalid format %s not supported", i.format)
}

type UnsupportedFormatError struct {
	format string
}

func NewUnsupportedFormatError(format string) *UnsupportedFormatError {
	return &UnsupportedFormatError{
		format: format,
	}
}

func (i *UnsupportedFormatError) Error() string {
	return fmt.Sprintf("Unsupported format %s not supported", i.format)
}
