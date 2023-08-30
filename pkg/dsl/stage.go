package dsl

import "context"

// Stage is a stage of a measurement pipeline.
type Stage[A, B any] interface {
	//ASTNode converts this stage to an ASTNode.
	ASTNode() *SerializableASTNode

	// Run runs the pipeline stage.
	Run(ctx context.Context, rtx Runtime, input Maybe[A]) Maybe[B]
}
