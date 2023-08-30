package dsl

import "context"

// operation is an internal definition used to characterize the internal implementation
// of network operations such as [dnsLookupGetaddrinfoOperation].
type operation[A, B any] interface {
	ASTNode() *SerializableASTNode
	Run(ctx context.Context, rtx Runtime, input A) (B, error)
}

// wrapOperation adapts an [operation] to behave like a [Stage].
func wrapOperation[A, B any](op operation[A, B]) Stage[A, B] {
	return &wrapOperationStage[A, B]{op}
}

type wrapOperationStage[A, B any] struct {
	op operation[A, B]
}

// ASTNode implements Stage.
func (sx *wrapOperationStage[A, B]) ASTNode() *SerializableASTNode {
	return sx.op.ASTNode()
}

// Run implements Stage.
func (sx *wrapOperationStage[A, B]) Run(ctx context.Context, rtx Runtime, input Maybe[A]) Maybe[B] {
	if input.Error != nil {
		return NewError[B](input.Error)
	}
	output, err := sx.op.Run(ctx, rtx, input.Value)
	if err != nil {
		return NewError[B](err)
	}
	return NewValue(output)
}
