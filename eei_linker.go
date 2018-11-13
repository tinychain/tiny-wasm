package tinywasm

import (
	"fmt"
	"github.com/tinychain/tiny-wasm/wagon/wasm"
)

func moduleResolver(w *WasmIntptr, name string) (*wasm.Module, error) {
	if name == "ethereum" {
		m := wasm.NewModule()
		m.Types.Entries = w.entries
		m.FunctionIndexSpace = w.funcs
		m.Export.Entries = w.exports

		return m, nil
	}
	return nil, fmt.Errorf("unknow module name %s", name)
}

func ModuleResolver(w *WasmIntptr) wasm.ResolveFunc {
	return func(name string) (*wasm.Module, error) {
		return moduleResolver(w, name)
	}
}
