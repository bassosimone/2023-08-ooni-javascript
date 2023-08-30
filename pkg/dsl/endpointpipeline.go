package dsl

import (
	"context"

	"github.com/ooni/probe-engine/pkg/runtimex"
)

// NewEndpointPipeline returns a stage that measures each endpoint given in input in
// parallel using a pool of background goroutines.
func NewEndpointPipeline(stage Stage[*Endpoint, *Void]) Stage[[]*Endpoint, *Void] {
	return &newEndpointPipelineStage{stage}
}

type newEndpointPipelineStage struct {
	sx Stage[*Endpoint, *Void]
}

const newEndpointPipelineStageName = "new_endpoint_pipeline"

func (sx *newEndpointPipelineStage) ASTNode() *SerializableASTNode {
	node := sx.sx.ASTNode()
	return &SerializableASTNode{
		StageName: newEndpointPipelineStageName,
		Arguments: nil,
		Children:  []*SerializableASTNode{node},
	}
}

type newEndpointPipelineLoader struct{}

// Load implements ASTLoaderRule.
func (*newEndpointPipelineLoader) Load(loader *ASTLoader, node *LoadableASTNode) (RunnableASTNode, error) {
	if err := loader.LoadEmptyArguments(node); err != nil {
		return nil, err
	}
	runnables, err := loader.LoadChildren(node)
	if err != nil {
		return nil, err
	}
	if len(runnables) != 1 {
		return nil, ErrInvalidNumberOfChildren
	}
	children := RunnableASTNodeListToStageList[*Endpoint, *Void](runnables[0])
	runtimex.Assert(len(children) == 1, "unexpected number of children")
	stage := NewEndpointPipeline(children[0])
	return &StageRunnableASTNode[[]*Endpoint, *Void]{stage}, nil
}

// StageName implements ASTLoaderRule.
func (*newEndpointPipelineLoader) StageName() string {
	return newEndpointPipelineStageName
}

func (sx *newEndpointPipelineStage) Run(ctx context.Context, rtx Runtime, input Maybe[[]*Endpoint]) Maybe[*Void] {
	if input.Error != nil {
		return NewError[*Void](input.Error)
	}

	// create list of workers
	var workers []Worker[Maybe[*Void]]
	for _, endpoint := range input.Value {
		workers = append(workers, &newEndpointPipelineWorker{rtx: rtx, sx: sx.sx, input: endpoint})
	}

	// perform the measurement in parallel
	const parallelism = 2
	results := ParallelRun(ctx, parallelism, workers...)

	// route exceptions
	if err := catch(results...); err != nil {
		return NewError[*Void](err)
	}

	return NewValue(&Void{})
}

// newEndpointPipelineWorker is the [Worker] used by [newEndpointPipelineStage].
type newEndpointPipelineWorker struct {
	input *Endpoint
	rtx   Runtime
	sx    Stage[*Endpoint, *Void]
}

func (w *newEndpointPipelineWorker) Produce(ctx context.Context) Maybe[*Void] {
	return w.sx.Run(ctx, w.rtx, NewValue(w.input))
}
