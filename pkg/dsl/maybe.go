package dsl

// Maybe contains either value or an error.
type Maybe[T any] struct {
	Error error
	Value T
}

// AsSpecificMaybe converts a generic Maybe[any] to a specific Maybe[T].
func AsSpecificMaybe[T any](v Maybe[any]) (Maybe[T], *ErrException) {
	switch x := v.Value.(type) {
	case T:
		return Maybe[T]{Error: v.Error, Value: x}, nil
	default:
		return Maybe[T]{}, NewTypeErrException[T](v.Value)
	}
}

// AsGeneric converts a specific Maybe[T] to a generic Maybe[any].
func (m Maybe[T]) AsGeneric() Maybe[any] {
	return Maybe[any]{
		Error: m.Error,
		Value: m.Value,
	}
}

// NewError constructs a [Maybe] with the given error.
func NewError[T any](err error) Maybe[T] {
	return Maybe[T]{Error: err, Value: *new(T)}
}

// NewValue constructs a [Maybe] with the given value.
func NewValue[T any](value T) Maybe[T] {
	return Maybe[T]{Error: nil, Value: value}
}
