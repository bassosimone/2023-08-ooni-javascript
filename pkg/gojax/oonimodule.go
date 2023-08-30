package gojax

import (
	"context"
	"encoding/json"
	"time"

	"github.com/bassosimone/2023-08-ooni-javascript/pkg/dsl"
	"github.com/dop251/goja"
	"github.com/ooni/probe-engine/pkg/runtimex"
)

// newModuleOONI creates the console module in JavaScript
func (vm *VM) newModuleOONI(gojaVM *goja.Runtime, mod *goja.Object) {
	runtimex.Assert(vm.vm == gojaVM, "gojax: unexpected gojaVM pointer value")
	exports := mod.Get("exports").(*goja.Object)
	exports.Set("runDSL", vm.ooniRunDSL)
}

func (vm *VM) ooniRunDSL(jsAST *goja.Object, zeroTime time.Time) (map[string]any, error) {
	// serialize the incoming JS object
	rawAST, err := jsAST.MarshalJSON()
	if err != nil {
		return nil, err
	}

	// parse the raw AST into the loadable AST format
	var loadableAST dsl.LoadableASTNode
	if err := json.Unmarshal(rawAST, &loadableAST); err != nil {
		return nil, err
	}

	// convert the loadable AST format into a runnable AST
	loader := dsl.NewASTLoader()
	runnableAST, err := loader.Load(&loadableAST)
	if err != nil {
		return nil, err
	}

	// TODO(bassosimone): we need to pass to this function the correct progressMeter

	// create the runtime objects required for interpreting a DSL
	metrics := dsl.NewAccountingMetrics()
	progressMeter := &dsl.NullProgressMeter{}
	rtx := dsl.NewMeasurexliteRuntime(vm.logger, metrics, progressMeter, zeroTime)
	input := dsl.NewValue(&dsl.Void{}).AsGeneric()

	// interpret the DSL and correctly route exceptions
	if err := dsl.Try(runnableAST.Run(context.Background(), rtx, input)); err != nil {
		return nil, err
	}

	// create a Go object to hold the results
	resultMap := map[string]any{
		"observations": dsl.ReduceObservations(rtx.ExtractObservations()...),
		"metrics":      metrics.Snapshot(),
	}

	// serialize the map to JSON
	resultRaw := runtimex.Try1(json.Marshal(resultMap))

	// create object holding the results
	var jsResult map[string]any
	if err := json.Unmarshal(resultRaw, &jsResult); err != nil {
		return nil, err
	}
	return jsResult, nil
}
