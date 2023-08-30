package dsl

import "errors"

// ErrSkip is a sentinel error indicating to a [Stage] that it should not run.
var ErrSkip = errors.New("dsl: skip this stage")

// IsErrSkip returns true when an error is an [ErrSkip].
func IsErrSkip(err error) bool {
	return errors.Is(err, ErrSkip)
}
