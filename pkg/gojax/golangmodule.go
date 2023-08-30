package gojax

import (
	"time"

	"github.com/dop251/goja"
	"github.com/ooni/probe-engine/pkg/runtimex"
)

// newModuleGolang creates the _golang module in JavaScript
func (vm *VM) newModuleGolang(gojaVM *goja.Runtime, mod *goja.Object) {
	runtimex.Assert(vm.vm == gojaVM, "gojax: unexpected gojaVM pointer value")
	exports := mod.Get("exports").(*goja.Object)
	exports.Set("timeNow", vm.golangTimeNow)
}

// golangTimeNow returns the current time using golang [time.Now]
func (vm *VM) golangTimeNow(call goja.FunctionCall) goja.Value {
	runtimex.Assert(len(call.Arguments) == 0, "gojax: _golang.timeNow expects zero arguments")
	return vm.vm.ToValue(time.Now())
}
