// Package maybe contains utility functions for an Optional data type.
package maybe

// Maybe is a generic type that holds a pointer to a value of type T
// and an error. Either *T will be nil, or error will be nil.
type Maybe[T any] struct {
	val *T
	err error
}

// Valid() returns true if Maybe contains a value of type *T.
func (m *Maybe[T]) Valid() bool {
	return m.err == nil && m.val != nil
}

// Error() returns the error contained in Maybe.
func (m *Maybe[T]) Error() error {
	return m.err
}

// Value() returns the value contained in Maybe.
func (m *Maybe[T]) Value() *T {
	return m.val
}

// Of is a function that takes a value of type T and returns
// a valid Maybe holding a pointer to that value.
func Of[T any](v T) *Maybe[T] {
	return &Maybe[T]{
		val: &v,
		err: nil,
	}
}

// Not is a function that returns an invalid Maybe that
// only contains the given error.
func Not[T any](err error) *Maybe[T] {
	return &Maybe[T]{
		val: nil,
		err: err,
	}
}
