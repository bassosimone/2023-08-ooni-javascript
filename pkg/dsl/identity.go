package dsl

import (
	"context"
)

// Identity returns a filter that copies its input to its output. We define as filter a
// [Stage] where the input and output type are the same type.
type Identity[T any] struct{}

const identityStageName = "identity"

// ASTNode implements Stage.
func (*Identity[T]) ASTNode() *SerializableASTNode {
	return &SerializableASTNode{
		StageName: identityStageName,
		Arguments: nil,
		Children:  []*SerializableASTNode{},
	}
}

type identityLoader struct{}

// Load implements ASTLoaderRule.
func (*identityLoader) Load(loader *ASTLoader, node *LoadableASTNode) (RunnableASTNode, error) {
	if err := loader.LoadEmptyArguments(node); err != nil {
		return nil, err
	}
	if err := loader.RequireExactlyNumChildren(node, 0); err != nil {
		return nil, err
	}
	stage := &Identity[any]{}
	return &StageRunnableASTNode[any, any]{stage}, nil
}

// StageName implements ASTLoaderRule.
func (*identityLoader) StageName() string {
	return identityStageName
}

// Run implements Stage.
func (*Identity[T]) Run(ctx context.Context, rtx Runtime, input Maybe[T]) Maybe[T] {
	return input
}
