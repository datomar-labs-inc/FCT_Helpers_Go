package fct_helpers

import "fmt"

func StrPtr(s string) *string {
	return &s
}

func StringSliceFromAnySlice[St any](slice []St) (out []string) {
	for _, st := range slice {
		out = append(out, fmt.Sprintf("%s", st))
	}

	return
}

func MapSlice[I any, T any](input []I, transform func(item I, index int) (T, error)) ([]T, error) {
	var result []T

	for i, inputItem := range input {
		transformed, err := transform(inputItem, i)
		if err != nil {
			return nil, err
		}

		result = append(result, transformed)
	}

	return result, nil
}

func MustMapSlice[I any, T any](input []I, transform func(item I, index int) (T, error)) []T {
	result, err := MapSlice(input, transform)
	if err != nil {
		panic(err)
	}

	return result
}

// NilToEmptySlice replaces a nil value with an empty slice of type S, otherwise returns the value unchanged
func NilToEmptySlice[S any](input []S) []S {
	if input == nil {
		return []S{}
	} else {
		return input
	}
}

func SliceToInterfaceSlice[I any](input []I) []any {
	var result []any

	for _, i := range input {
		result = append(result, i)
	}

	return result
}

func Must[R any](returnValue R, err error) R {
	if err != nil {
		panic(err)
	}

	return returnValue
}
