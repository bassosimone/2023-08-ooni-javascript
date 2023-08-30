package dsl

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/ooni/probe-engine/pkg/model"
)

// ProgressMeter tracks progress.
type ProgressMeter interface {
	// IncrementProgress increments the progress meter by adding the given delta
	// to the current progress meter value. The progress meter value is a float
	// number where 0 means beginning and 1.0 means we are done.
	IncrementProgress(delta float64)
}

// NullProgressMeter is a [ProgressMeter] that does nothing. The zero
// value of this struct is ready to use.
type NullProgressMeter struct{}

var _ ProgressMeter = &NullProgressMeter{}

// IncrementProgress implements ProgressMeter.
func (pm *NullProgressMeter) IncrementProgress(value float64) {
	// nothing
}

// ProgressMeterExperimentCallbacks wraps [model.ExperimentCallbacks] and
// implements [ProgressMeter]. The zero value is not ready to use; you should
// construct using the [NewProgressMeterExperimentCallbacks] factory.
type ProgressMeterExperimentCallbacks struct {
	callbacks model.ExperimentCallbacks
	mu        sync.Mutex
	total     float64
}

// NewProgressMeterExperimentCallbacks constructs a new [ProgressMeterExperimentCallbacks].
func NewProgressMeterExperimentCallbacks(cb model.ExperimentCallbacks) *ProgressMeterExperimentCallbacks {
	return &ProgressMeterExperimentCallbacks{
		callbacks: cb,
		mu:        sync.Mutex{},
		total:     0,
	}
}

var _ ProgressMeter = &ProgressMeterExperimentCallbacks{}

// IncrementProgress implements ProgressMeter.
func (pm *ProgressMeterExperimentCallbacks) IncrementProgress(delta float64) {
	pm.mu.Lock()
	if delta >= 0 {
		pm.total += delta
		if pm.total > 1.0 {
			pm.total = 1.0
		}
	}
	total := pm.total
	pm.mu.Unlock()
	pm.callbacks.OnProgress(total, "")
}

// WrapWithProgress wraps a list of stages such that each stage increments the
// progress of running a measurement by an equal contribution.
func WrapWithProgress(input ...Stage[*Void, *Void]) (output []Stage[*Void, *Void]) {
	var delta float64
	if len(input) > 0 {
		delta = 1 / float64(len(input))
	}
	for _, stage := range input {
		output = append(output, &wrapWithProgressStage{delta, stage})
	}
	return output
}

type wrapWithProgressStage struct {
	delta float64
	stage Stage[*Void, *Void]
}

const wrapWithProgressStageName = "wrap_with_progress"

type wrapWithProgressStageArguments struct {
	Delta float64 `json:"delta"`
}

// ASTNode implements Stage.
func (sx *wrapWithProgressStage) ASTNode() *SerializableASTNode {
	// Note: we serialize the structure because this gives us forward compatibility (i.e., we
	// may add a field to a future version without breaking the AST structure and old probes will
	// be fine as long as the zero value of the new field is the default)
	return &SerializableASTNode{
		StageName: wrapWithProgressStageName,
		Arguments: &wrapWithProgressStageArguments{sx.delta},
		Children:  []*SerializableASTNode{sx.stage.ASTNode()},
	}
}

type wrapWithProgressLoader struct{}

// Load implements ASTLoaderRule.
func (*wrapWithProgressLoader) Load(loader *ASTLoader, node *LoadableASTNode) (RunnableASTNode, error) {
	var config wrapWithProgressStageArguments
	if err := json.Unmarshal(node.Arguments, &config); err != nil {
		return nil, err
	}
	runnables, err := loader.LoadChildren(node)
	if err != nil {
		return nil, err
	}
	if len(runnables) != 1 {
		return nil, ErrInvalidNumberOfChildren
	}
	runnables0 := &RunnableASTNodeStage[*Void, *Void]{runnables[0]}
	stage := &wrapWithProgressStage{config.Delta, runnables0}
	return &StageRunnableASTNode[*Void, *Void]{stage}, nil
}

// StageName implements ASTLoaderRule.
func (*wrapWithProgressLoader) StageName() string {
	return wrapWithProgressStageName
}

// Run implements Stage.
func (sx *wrapWithProgressStage) Run(ctx context.Context, rtx Runtime, input Maybe[*Void]) Maybe[*Void] {
	output := sx.stage.Run(ctx, rtx, input)
	rtx.ProgressMeter().IncrementProgress(sx.delta)
	return output
}
