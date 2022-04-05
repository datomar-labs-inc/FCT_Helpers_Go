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
