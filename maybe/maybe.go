package maybe

import (
	"encoding/json"
)

type Maybe[T any] struct {
	hasValue bool
	value    T
}

func (m Maybe[T]) HasValue() bool {
	return m.hasValue
}

func (m Maybe[T]) Value() (T, bool) {
	if !m.hasValue {
		return m.value, false
	}

	return m.value, true
}

func (m Maybe[T]) UnmarshalJSON(bytes []byte) error {
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

func (m Maybe[T]) MarshalJSON() ([]byte, error) {
	if m.hasValue {
		return json.Marshal(m.value)
	} else {
		return []byte("null"), nil
	}
}

func (m Maybe[T]) If(ifFunc func(val T)) {
	if m.hasValue {
		ifFunc(m.value)
	}
}

func (m Maybe[T]) IfSetCopyTo(target *T) {
	if m.hasValue {
		if target != nil {
			*target = m.value
		}
	}
}

func (m Maybe[T]) Or(defaultValue T) T {
	if m.hasValue {
		return m.value
	} else {
		return defaultValue
	}
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
