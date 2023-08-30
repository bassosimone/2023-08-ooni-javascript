package dsl

import "context"

// Worker produces a given result.
type Worker[T any] interface {
	Produce(ctx context.Context) T
}
