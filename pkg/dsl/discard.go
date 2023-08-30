package dsl

import (
	"context"
)

// Discard returns a stage that discards its input value with type T. You need this stage
// to make sure your endpoint pipeline returns a [*Void] value.
func Discard[T any]() Stage[T, *Void] {
	return &discardStage[T]{}
}

type discardStage[T any] struct{}

const discardStageName = "discard"

// ASTNode implements Stage.
func (sx *discardStage[T]) ASTNode() *SerializableASTNode {
	// There is type erasure when we AST-serialize
	return &SerializableASTNode{
		StageName: discardStageName,
		Arguments: nil,
		Children:  []*SerializableASTNode{},
	}
}

type discardLoader struct{}

// Load implements ASTLoaderRule.
func (*discardLoader) Load(loader *ASTLoader, node *LoadableASTNode) (RunnableASTNode, error) {
	if err := loader.LoadEmptyArguments(node); err != nil {
		return nil, err
	}
	if err := loader.RequireExactlyNumChildren(node, 0); err != nil {
		return nil, err
	}
	// Because there is type erasure, we create a Discard[any], which is fine because the
	// type parameter is for the input and not for the output
	stage := Discard[any]()
	return &StageRunnableASTNode[any, *Void]{stage}, nil
}

// StageName implements ASTLoaderRule.
func (*discardLoader) StageName() string {
	return discardStageName
}

// Run implements Stage.
func (sx *discardStage[T]) Run(ctx context.Context, rtx Runtime, input Maybe[T]) Maybe[*Void] {
	if input.Error != nil {
		return NewError[*Void](input.Error)
	}
	return NewValue(&Void{})
}
