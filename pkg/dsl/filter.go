package dsl

import (
	"context"
)

// IfFilterExists wraps a filter such that probes interpreting the AST can compile the filter
// to an identity function if a filter with the given name does not exist. We define filter as
// a [Stage] where the input type and the output type are the same type. This functionality
// allows supporting old probes that do not support specific filters. Such probes will compile
// and execute the AST and run identity functions in place of the unsupported filters.
func IfFilterExists[T any](fx Stage[T, T]) Stage[T, T] {
	return &ifFilterExistsStage[T]{fx}
}

type ifFilterExistsStage[T any] struct {
	fx Stage[T, T]
}

const ifFilterExistsStageName = "if_filter_exists"

// ASTNode implements Stage.
func (fx *ifFilterExistsStage[T]) ASTNode() *SerializableASTNode {
	return &SerializableASTNode{
		StageName: ifFilterExistsStageName,
		Arguments: nil,
		Children:  []*SerializableASTNode{fx.fx.ASTNode()},
	}
}

type ifFilterExistsLoader struct{}

// Load implements ASTLoaderRule.
func (*ifFilterExistsLoader) Load(loader *ASTLoader, node *LoadableASTNode) (RunnableASTNode, error) {
	if err := loader.LoadEmptyArguments(node); err != nil {
		return nil, err
	}
	if err := loader.RequireExactlyNumChildren(node, 1); err != nil {
		return nil, err
	}
	runnable, err := loader.Load(node.Children[0])
	if err != nil {
		// If we cannot load the given filter, replace it with the identity
		return &Identity[any]{}, nil
	}
	// Otherwise, we replace the ifFilterExists node with the existing underlying node
	return runnable, nil
}

// StageName implements ASTLoaderRule.
func (*ifFilterExistsLoader) StageName() string {
	return ifFilterExistsStageName
}

// Run implements Stage.
func (fx *ifFilterExistsStage[T]) Run(ctx context.Context, rtx Runtime, input Maybe[T]) Maybe[T] {
	// When we created the ifFilterExistsStage directly the Go code has compiled so
	// the underlying filter exists and we can run it directly.
	return fx.fx.Run(ctx, rtx, input)
}
