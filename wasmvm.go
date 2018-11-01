package wagon

import (
	"github.com/tinychain/tiny-wasm/wagon/exec"
	"github.com/tinychain/tiny-wasm/wagon/wasm"
)

type WasmVM struct {
	context  EnvContext
	handlers map[string]interface{}
	vm       *exec.VM
}

func New(context EnvContext) *WasmVM {

	w := &WasmVM{
		context:  context,
		handlers: make(map[string]interface{}),
	}

	return w
}

func (w *WasmVM) register(name string, handler interface{}) {
	w.handlers[name] = handler
}

func (w *WasmVM) GetHandlers() map[string]interface{} {
	return w.handlers
}

func (w *WasmVM) GetHandler(name string) interface{} {
	return w.handlers[name]
}

func (w *WasmVM) importer(name string) (*wasm.Module, error) {
	if name == "env" {

	}
}
