package fcthelp

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"reflect"
	"time"
	"unsafe"
)

func StrPtr(s string) *string {
	return &s
}

func ToPtr[T any](v T) *T {
	return &v
}

func StringSliceFromAnySlice[St any](slice []St) (out []string) {
	for _, st := range slice {
		out = append(out, fmt.Sprintf("%s", st))
	}

	return
}

func SliceContains[St comparable](s []St, v St) bool {
	for _, st := range s {
		if st == v {
			return true
		}
	}

	return false
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
	}

	return input
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

func IsNil(i any) bool {
	if i == nil {
		return true
	}
	switch reflect.TypeOf(i).Kind() {
	case reflect.Ptr, reflect.Map, reflect.Array, reflect.Chan, reflect.Slice:
		return reflect.ValueOf(i).IsNil()
	}
	return false
}

// JSONConvert will convert anything into a type via json serialization
func JSONConvert[R any](input any) (R, bool, error) {
	var output R

	jsb, err := json.Marshal(input)
	if err != nil {
		return output, false, err
	}

	err = json.Unmarshal(jsb, &output)
	if err != nil {
		return output, false, err
	}

	return output, true, nil
}

type Array[T any] []T

func (a *Array[T]) Push(item T) {
	*a = append(*a, item)
}

var src = rand.NewSource(time.Now().UnixNano())

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

func RandString(n int) string {
	b := make([]byte, n)
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return *(*string)(unsafe.Pointer(&b))
}
