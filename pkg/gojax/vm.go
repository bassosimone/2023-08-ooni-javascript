package gojax

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/require"
	"github.com/dop251/goja_nodejs/util"
	"github.com/ooni/probe-engine/pkg/logx"
	"github.com/ooni/probe-engine/pkg/model"
)

// VMConfig contains configuration for creating a VM.
type VMConfig struct {
	// Logger is the MANDATORY logger to use.
	Logger model.Logger

	// ScriptBaseDir is the MANDATORY script base dir to use.
	ScriptBaseDir string
}

// errVMConfig indicates that some setting in the [*VMConfig] is invalid.
var errVMConfig = errors.New("gojax: invalid VMConfig")

// check returns an explanatory error if the [*VMConfig] is invalid.
func (cfg *VMConfig) check() error {
	if cfg.Logger == nil {
		return fmt.Errorf("%w: the Logger field is nil", errVMConfig)
	}

	if cfg.ScriptBaseDir == "" {
		return fmt.Errorf("%w: the ScriptBaseDir field is empty", errVMConfig)
	}

	return nil
}

// VM wraps the [*github.com/dop251/goja.Runtime]. The zero value of this
// struct is invalid; please, use [NewVM] to construct.
type VM struct {
	// logger is the logger to use.
	logger model.Logger

	// registry is the JavaScript package registry to use.
	registry *require.Registry

	// scriptBaseDir is the base directory containing scripts.
	scriptBaseDir string

	// util is a reference to goja's util model.
	util *goja.Object

	// vm is a reference to goja's runtime.
	vm *goja.Runtime
}

// NewVM creates a new instance of the [*VM].
func NewVM(config *VMConfig) (*VM, error) {
	// make sure the provided config is correct
	if err := config.check(); err != nil {
		return nil, err
	}

	// convert the script base dir to be an absolute path
	scriptBaseDir, err := filepath.Abs(config.ScriptBaseDir)
	if err != nil {
		return nil, err
	}

	// create package registry ("By default, a registry's global folders list is empty")
	registry := require.NewRegistry(require.WithGlobalFolders(scriptBaseDir))

	// create the goja virtual machine
	gojaVM := goja.New()

	// enable 'require' for the virtual machine
	registry.Enable(gojaVM)

	// make sure the JavaScript logger has a prefix
	logger := &logx.PrefixLogger{
		Prefix: "[JavaScriptConsole] ",
		Logger: config.Logger,
	}

	// create the virtual machine wrapper
	vm := &VM{
		logger:        logger,
		registry:      registry,
		scriptBaseDir: scriptBaseDir,
		util:          require.Require(gojaVM, util.ModuleName).(*goja.Object),
		vm:            gojaVM,
	}

	// register the console module in JavaScript
	registry.RegisterNativeModule("console", vm.newModuleConsole)

	// make sure the 'console' object exists in the VM before running scripts
	gojaVM.Set("console", require.Require(gojaVM, "console"))

	// register the _golang module in JavaScript
	registry.RegisterNativeModule("_golang", vm.newModuleGolang)

	// register the _ooni module in JavaScript
	registry.RegisterNativeModule("_ooni", vm.newModuleOONI)

	return vm, nil
}

// RunScript runs the given script file discarding its return value.
func (vm *VM) RunScript(fpath string) error {
	content, err := os.ReadFile(fpath)
	if err != nil {
		return err
	}
	_, err = vm.vm.RunScript(fpath, string(content))
	return err
}
