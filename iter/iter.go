package iter

import "errors"

var ErrNoMoreElements = errors.New("no more elements")

type Iterator[T any] interface {
	Next() (T, error)
	HasNext() bool
}
