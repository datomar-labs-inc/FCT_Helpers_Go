package maybe

import (
	"encoding/json"
)

type Maybe[T any] struct {
	hasValue bool
	value    T
}

// Ish is any item that can have a value or not
// this interface is used to implement some nice helpers for multiple types of mabies
type Ish interface {
	HasValue() bool
}

//goland:noinspection GoMixedReceiverTypes
func (m Maybe[T]) HasValue() bool {
	return m.hasValue
}

func (m Maybe[T]) Value() (T, bool) {
	if !m.hasValue {
		return m.value, false
	}

	return m.value, true
}

// UnmarshalJSON
// Has a pointer receiver so that JSON can be unmarshalled into the proper target object
// Without this, it will not work
//
//goland:noinspection GoMixedReceiverTypes
func (m *Maybe[T]) UnmarshalJSON(bytes []byte) error {
	if m == nil {
		return nil
	}

	if string(bytes) == "null" {
		return nil
	} else {
		var value T

		err := json.Unmarshal(bytes, &value)
		if err != nil {
			return err
		}

		m.value = value
		m.hasValue = true
	}

	return nil
}

//goland:noinspection GoMixedReceiverTypes
func (m Maybe[T]) MarshalJSON() ([]byte, error) {
	if m.hasValue {
		return json.Marshal(m.value)
	} else {
		return []byte("null"), nil
	}
}

//goland:noinspection GoMixedReceiverTypes
func (m Maybe[T]) If(ifFunc func(val T)) {
	if m.hasValue {
		ifFunc(m.value)
	}
}

//goland:noinspection GoMixedReceiverTypes
func (m Maybe[T]) IfSetCopyTo(target *T) {
	if m.hasValue {
		if target != nil {
			*target = m.value
		}
	}
}

//goland:noinspection GoMixedReceiverTypes
func (m Maybe[T]) Or(defaultValue T) T {
	if m.hasValue {
		return m.value
	} else {
		return defaultValue
	}
}

func Map[St any, Mt any](from Maybe[St], mapFunc func(value St) Mt) Maybe[Mt] {
	if v, ok := from.Value(); ok {
		return WithValue(mapFunc(v))
	}

	return Maybe[Mt]{}
}

func IsMoreThanOneSet(maybies ...Ish) bool {
	isOneSet := false

	for _, maybe := range maybies {
		if maybe.HasValue() {
			if isOneSet {
				return true
			}

			isOneSet = true
		}
	}

	return false
}

func IsAnySet(maybies ...Ish) bool {
	for _, maybe := range maybies {
		if maybe.HasValue() {
			return true
		}
	}

	return false
}

func WithValue[T any](val T) Maybe[T] {
	return Maybe[T]{
		value:    val,
		hasValue: true,
	}
}

func Empty[T any]() Maybe[T] {
	return Maybe[T]{
		hasValue: false,
	}
}
