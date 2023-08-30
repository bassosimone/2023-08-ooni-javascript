package main

import (
	"os"

	"github.com/apex/log"
	"github.com/bassosimone/2023-08-ooni-javascript/pkg/gojax"
	"github.com/ooni/probe-engine/pkg/runtimex"
)

func main() {
	config := &gojax.VMConfig{
		Logger:        log.Log,
		ScriptBaseDir: "./javascript",
	}
	vm := runtimex.Try1(gojax.NewVM(config))
	runtimex.Try0(vm.RunScript(os.Args[1]))
}
