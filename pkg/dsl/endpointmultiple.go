package dsl

import (
	"context"
)

// MeasureMultipleEndpoints returns a stage that runs several endpoint measurement
// pipelines in parallel using a pool of background goroutines.
func MeasureMultipleEndpoints(stages ...Stage[*DNSLookupResult, *Void]) Stage[*DNSLookupResult, *Void] {
	return &measureMultipleEndpointsStage{stages}
}

type measureMultipleEndpointsStage struct {
	stages []Stage[*DNSLookupResult, *Void]
}

const measureMultipleEndpointsStageName = "measure_multiple_endpoints"

// ASTNode implements Stage.
func (sx *measureMultipleEndpointsStage) ASTNode() *SerializableASTNode {
	var nodes []*SerializableASTNode
	for _, stage := range sx.stages {
		nodes = append(nodes, stage.ASTNode())
	}
	return &SerializableASTNode{
		StageName: measureMultipleEndpointsStageName,
		Arguments: nil,
		Children:  nodes,
	}
}

type measureMultipleEndpointsLoader struct{}

// Load implements ASTLoaderRule.
func (*measureMultipleEndpointsLoader) Load(loader *ASTLoader, node *LoadableASTNode) (RunnableASTNode, error) {
	if err := loader.LoadEmptyArguments(node); err != nil {
		return nil, err
	}
	runnables, err := loader.LoadChildren(node)
	if err != nil {
		return nil, err
	}
	children := RunnableASTNodeListToStageList[*DNSLookupResult, *Void](runnables...)
	stage := MeasureMultipleEndpoints(children...)
	return &StageRunnableASTNode[*DNSLookupResult, *Void]{stage}, nil
}

// StageName implements ASTLoaderRule.
func (*measureMultipleEndpointsLoader) StageName() string {
	return measureMultipleEndpointsStageName
}

// Run implements stage.
func (sx *measureMultipleEndpointsStage) Run(ctx context.Context, rtx Runtime, input Maybe[*DNSLookupResult]) Maybe[*Void] {
	if input.Error != nil {
		return NewError[*Void](input.Error)
	}

	// initialize the workers
	var workers []Worker[Maybe[*Void]]
	for _, stage := range sx.stages {
		workers = append(workers, &measureMultipleEndpointsWorker{input: input, rtx: rtx, sx: stage})
	}

	// parallel run
	const parallelism = 2
	results := ParallelRun(ctx, parallelism, workers...)

	// route exceptions
	if err := catch(results...); err != nil {
		return NewError[*Void](err)
	}

	return NewValue(&Void{})
}

type measureMultipleEndpointsWorker struct {
	input Maybe[*DNSLookupResult]
	rtx   Runtime
	sx    Stage[*DNSLookupResult, *Void]
}

// Produce implements Worker.
func (w *measureMultipleEndpointsWorker) Produce(ctx context.Context) Maybe[*Void] {
	return w.sx.Run(ctx, w.rtx, w.input)
}
