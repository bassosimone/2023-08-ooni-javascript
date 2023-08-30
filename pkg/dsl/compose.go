package dsl

//
// Stage composition
//

import (
	"context"

	"github.com/ooni/probe-engine/pkg/runtimex"
)

// Compose composes two [Stage] together.
func Compose[A, B, C any](s1 Stage[A, B], s2 Stage[B, C]) Stage[A, C] {
	return &composeStage[A, B, C]{s1, s2}
}

type composeStage[A, B, C any] struct {
	s1 Stage[A, B]
	s2 Stage[B, C]
}

const composeStageName = "compose"

// ASTNode implements Stage.
func (sx *composeStage[A, B, C]) ASTNode() *SerializableASTNode {
	n1, n2 := sx.s1.ASTNode(), sx.s2.ASTNode()
	return &SerializableASTNode{
		StageName: composeStageName,
		Arguments: nil,
		Children:  []*SerializableASTNode{n1, n2},
	}
}

type composeLoader struct{}

// Load implements ASTLoaderRule.
func (*composeLoader) Load(loader *ASTLoader, node *LoadableASTNode) (RunnableASTNode, error) {
	if err := loader.LoadEmptyArguments(node); err != nil {
		return nil, err
	}
	if err := loader.RequireExactlyNumChildren(node, 2); err != nil {
		return nil, err
	}

	runnables, err := loader.LoadChildren(node)
	if err != nil {
		return nil, err
	}

	// Note: we Compose using `any` but we're not creating any Maybe[any] in the [composeStage.Run]
	// method and inner stages should create correctly-typed Maybes.
	runtimex.Assert(len(runnables) == 2, "expected exactly two children nodes")
	stage := Compose[any, any, any](runnables[0], runnables[1])
	return stage, nil
}

// StageName implements ASTLoaderRule.
func (*composeLoader) StageName() string {
	return composeStageName
}

// Run implements Stage.
func (sx *composeStage[A, B, C]) Run(ctx context.Context, rtx Runtime, input Maybe[A]) Maybe[C] {
	// Note: we cannot create any Maybe here because we may be a composeStage[any, any, any]
	// and hence creating a Maybe here would generate an incorrectly typed Maybe
	return sx.s2.Run(ctx, rtx, sx.s1.Run(ctx, rtx, input))
}

// Compose3 composes 3 [Stage] together.
func Compose3[
	T0,
	T1,
	T2,
	T3 any,
](
	s1 Stage[T0, T1],
	s2 Stage[T1, T2],
	s3 Stage[T2, T3],
) Stage[T0, T3] {
	return Compose(s1, Compose(s2, s3))
}

// Compose4 composes 4 [Stage] together.
func Compose4[
	T0,
	T1,
	T2,
	T3,
	T4 any,
](
	s1 Stage[T0, T1],
	s2 Stage[T1, T2],
	s3 Stage[T2, T3],
	s4 Stage[T3, T4],
) Stage[T0, T4] {
	return Compose(s1, Compose3(s2, s3, s4))
}

// Compose5 composes 5 [Stage] together.
func Compose5[
	T0,
	T1,
	T2,
	T3,
	T4,
	T5 any,
](
	s1 Stage[T0, T1],
	s2 Stage[T1, T2],
	s3 Stage[T2, T3],
	s4 Stage[T3, T4],
	s5 Stage[T4, T5],
) Stage[T0, T5] {
	return Compose(s1, Compose4(s2, s3, s4, s5))
}

// Compose6 composes 6 [Stage] together.
func Compose6[
	T0,
	T1,
	T2,
	T3,
	T4,
	T5,
	T6 any,
](
	s1 Stage[T0, T1],
	s2 Stage[T1, T2],
	s3 Stage[T2, T3],
	s4 Stage[T3, T4],
	s5 Stage[T4, T5],
	s6 Stage[T5, T6],
) Stage[T0, T6] {
	return Compose(s1, Compose5(s2, s3, s4, s5, s6))
}

// Compose7 composes 7 [Stage] together.
func Compose7[
	T0,
	T1,
	T2,
	T3,
	T4,
	T5,
	T6,
	T7 any,
](
	s1 Stage[T0, T1],
	s2 Stage[T1, T2],
	s3 Stage[T2, T3],
	s4 Stage[T3, T4],
	s5 Stage[T4, T5],
	s6 Stage[T5, T6],
	s7 Stage[T6, T7],
) Stage[T0, T7] {
	return Compose(s1, Compose6(s2, s3, s4, s5, s6, s7))
}

// Compose8 composes 8 [Stage] together.
func Compose8[
	T0,
	T1,
	T2,
	T3,
	T4,
	T5,
	T6,
	T7,
	T8 any,
](
	s1 Stage[T0, T1],
	s2 Stage[T1, T2],
	s3 Stage[T2, T3],
	s4 Stage[T3, T4],
	s5 Stage[T4, T5],
	s6 Stage[T5, T6],
	s7 Stage[T6, T7],
	s8 Stage[T7, T8],
) Stage[T0, T8] {
	return Compose(s1, Compose7(s2, s3, s4, s5, s6, s7, s8))
}

// Compose9 composes 9 [Stage] together.
func Compose9[
	T0,
	T1,
	T2,
	T3,
	T4,
	T5,
	T6,
	T7,
	T8,
	T9 any,
](
	s1 Stage[T0, T1],
	s2 Stage[T1, T2],
	s3 Stage[T2, T3],
	s4 Stage[T3, T4],
	s5 Stage[T4, T5],
	s6 Stage[T5, T6],
	s7 Stage[T6, T7],
	s8 Stage[T7, T8],
	s9 Stage[T8, T9],
) Stage[T0, T9] {
	return Compose(s1, Compose8(s2, s3, s4, s5, s6, s7, s8, s9))
}

// Compose10 composes 10 [Stage] together.
func Compose10[
	T0,
	T1,
	T2,
	T3,
	T4,
	T5,
	T6,
	T7,
	T8,
	T9,
	T10 any,
](
	s1 Stage[T0, T1],
	s2 Stage[T1, T2],
	s3 Stage[T2, T3],
	s4 Stage[T3, T4],
	s5 Stage[T4, T5],
	s6 Stage[T5, T6],
	s7 Stage[T6, T7],
	s8 Stage[T7, T8],
	s9 Stage[T8, T9],
	s10 Stage[T9, T10],
) Stage[T0, T10] {
	return Compose(s1, Compose9(s2, s3, s4, s5, s6, s7, s8, s9, s10))
}

// Compose11 composes 11 [Stage] together.
func Compose11[
	T0,
	T1,
	T2,
	T3,
	T4,
	T5,
	T6,
	T7,
	T8,
	T9,
	T10,
	T11 any,
](
	s1 Stage[T0, T1],
	s2 Stage[T1, T2],
	s3 Stage[T2, T3],
	s4 Stage[T3, T4],
	s5 Stage[T4, T5],
	s6 Stage[T5, T6],
	s7 Stage[T6, T7],
	s8 Stage[T7, T8],
	s9 Stage[T8, T9],
	s10 Stage[T9, T10],
	s11 Stage[T10, T11],
) Stage[T0, T11] {
	return Compose(s1, Compose10(s2, s3, s4, s5, s6, s7, s8, s9, s10, s11))
}

// Compose12 composes 12 [Stage] together.
func Compose12[
	T0,
	T1,
	T2,
	T3,
	T4,
	T5,
	T6,
	T7,
	T8,
	T9,
	T10,
	T11,
	T12 any,
](
	s1 Stage[T0, T1],
	s2 Stage[T1, T2],
	s3 Stage[T2, T3],
	s4 Stage[T3, T4],
	s5 Stage[T4, T5],
	s6 Stage[T5, T6],
	s7 Stage[T6, T7],
	s8 Stage[T7, T8],
	s9 Stage[T8, T9],
	s10 Stage[T9, T10],
	s11 Stage[T10, T11],
	s12 Stage[T11, T12],
) Stage[T0, T12] {
	return Compose(s1, Compose11(s2, s3, s4, s5, s6, s7, s8, s9, s10, s11, s12))
}

// Compose13 composes 13 [Stage] together.
func Compose13[
	T0,
	T1,
	T2,
	T3,
	T4,
	T5,
	T6,
	T7,
	T8,
	T9,
	T10,
	T11,
	T12,
	T13 any,
](
	s1 Stage[T0, T1],
	s2 Stage[T1, T2],
	s3 Stage[T2, T3],
	s4 Stage[T3, T4],
	s5 Stage[T4, T5],
	s6 Stage[T5, T6],
	s7 Stage[T6, T7],
	s8 Stage[T7, T8],
	s9 Stage[T8, T9],
	s10 Stage[T9, T10],
	s11 Stage[T10, T11],
	s12 Stage[T11, T12],
	s13 Stage[T12, T13],
) Stage[T0, T13] {
	return Compose(s1, Compose12(s2, s3, s4, s5, s6, s7, s8, s9, s10, s11, s12, s13))
}

// Compose14 composes 14 [Stage] together.
func Compose14[
	T0,
	T1,
	T2,
	T3,
	T4,
	T5,
	T6,
	T7,
	T8,
	T9,
	T10,
	T11,
	T12,
	T13,
	T14 any,
](
	s1 Stage[T0, T1],
	s2 Stage[T1, T2],
	s3 Stage[T2, T3],
	s4 Stage[T3, T4],
	s5 Stage[T4, T5],
	s6 Stage[T5, T6],
	s7 Stage[T6, T7],
	s8 Stage[T7, T8],
	s9 Stage[T8, T9],
	s10 Stage[T9, T10],
	s11 Stage[T10, T11],
	s12 Stage[T11, T12],
	s13 Stage[T12, T13],
	s14 Stage[T13, T14],
) Stage[T0, T14] {
	return Compose(s1, Compose13(s2, s3, s4, s5, s6, s7, s8, s9, s10, s11, s12, s13, s14))
}

// Compose15 composes 15 [Stage] together.
func Compose15[
	T0,
	T1,
	T2,
	T3,
	T4,
	T5,
	T6,
	T7,
	T8,
	T9,
	T10,
	T11,
	T12,
	T13,
	T14,
	T15 any,
](
	s1 Stage[T0, T1],
	s2 Stage[T1, T2],
	s3 Stage[T2, T3],
	s4 Stage[T3, T4],
	s5 Stage[T4, T5],
	s6 Stage[T5, T6],
	s7 Stage[T6, T7],
	s8 Stage[T7, T8],
	s9 Stage[T8, T9],
	s10 Stage[T9, T10],
	s11 Stage[T10, T11],
	s12 Stage[T11, T12],
	s13 Stage[T12, T13],
	s14 Stage[T13, T14],
	s15 Stage[T14, T15],
) Stage[T0, T15] {
	return Compose(s1, Compose14(s2, s3, s4, s5, s6, s7, s8, s9, s10, s11, s12, s13, s14, s15))
}

// Compose16 composes 16 [Stage] together.
func Compose16[
	T0,
	T1,
	T2,
	T3,
	T4,
	T5,
	T6,
	T7,
	T8,
	T9,
	T10,
	T11,
	T12,
	T13,
	T14,
	T15,
	T16 any,
](
	s1 Stage[T0, T1],
	s2 Stage[T1, T2],
	s3 Stage[T2, T3],
	s4 Stage[T3, T4],
	s5 Stage[T4, T5],
	s6 Stage[T5, T6],
	s7 Stage[T6, T7],
	s8 Stage[T7, T8],
	s9 Stage[T8, T9],
	s10 Stage[T9, T10],
	s11 Stage[T10, T11],
	s12 Stage[T11, T12],
	s13 Stage[T12, T13],
	s14 Stage[T13, T14],
	s15 Stage[T14, T15],
	s16 Stage[T15, T16],
) Stage[T0, T16] {
	return Compose(s1, Compose15(s2, s3, s4, s5, s6, s7, s8, s9, s10, s11, s12, s13, s14, s15, s16))
}

// Compose17 composes 17 [Stage] together.
func Compose17[
	T0,
	T1,
	T2,
	T3,
	T4,
	T5,
	T6,
	T7,
	T8,
	T9,
	T10,
	T11,
	T12,
	T13,
	T14,
	T15,
	T16,
	T17 any,
](
	s1 Stage[T0, T1],
	s2 Stage[T1, T2],
	s3 Stage[T2, T3],
	s4 Stage[T3, T4],
	s5 Stage[T4, T5],
	s6 Stage[T5, T6],
	s7 Stage[T6, T7],
	s8 Stage[T7, T8],
	s9 Stage[T8, T9],
	s10 Stage[T9, T10],
	s11 Stage[T10, T11],
	s12 Stage[T11, T12],
	s13 Stage[T12, T13],
	s14 Stage[T13, T14],
	s15 Stage[T14, T15],
	s16 Stage[T15, T16],
	s17 Stage[T16, T17],
) Stage[T0, T17] {
	return Compose(s1, Compose16(s2, s3, s4, s5, s6, s7, s8, s9, s10, s11, s12, s13, s14, s15, s16, s17))
}
