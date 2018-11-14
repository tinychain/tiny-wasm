package tinywasm

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/tinychain/tinychain/core/vm"
	"reflect"

	"github.com/tinychain/tiny-wasm/wagon/exec"
	"github.com/tinychain/tiny-wasm/wagon/wasm"
)

type TerminateType int

const (
	TerminateFinish = iota
	TerminateRevert
	TerminateSuicide
	TerminateInvalid
)

type funcSet struct {
	entries []wasm.FunctionSig
	funcs   []wasm.Function
	exports map[string]wasm.ExportEntry
}

type WasmIntptr struct {
	// execution fields
	vm            *exec.VM
	contract      *Contract
	readonly      bool          // static mode
	evm           *EVM          // evm instance
	terminateType TerminateType // termination type of the execution
	returnData    []byte        // returning output data for the execution

	// module resolver components
	handlers      map[string]interface{} // eei function handlers
	debugHandlers map[string]interface{} // debug function handlers
	eeiFuncSet    *funcSet               // eei function set
	debugFuncSet  *funcSet               // debug function set
}

func NewWasmIntptr(evm *EVM) *WasmIntptr {
	w := &WasmIntptr{
		evm:        evm,
		handlers:   make(map[string]interface{}),
		eeiFuncSet: &funcSet{exports: make(map[string]wasm.ExportEntry)},
	}

	w.initEEIModule()
	if w.evm.vmConfig.Debug {
		w.initDebugModule()
	}

	return w
}

func (w *WasmIntptr) initEEIModule() {
	// eei function register
	api := &eeiApi{}
	rapi := reflect.TypeOf(api)
	fnCount := rapi.NumMethod()
	for i := 0; i < fnCount; i++ {
		fn := rapi.Method(i)
		w.handlers[fn.Name] = fn
	}

	i := 0
	for k, f := range w.handlers {
		rType := reflect.TypeOf(f)
		numIn := rType.NumIn() - 2
		args := make([]wasm.ValueType, numIn)
		for j := 0; j < numIn; j++ {
			args[j] = goType2WasmType(rType.In(j + 2).Kind())
		}

		numOut := rType.NumOut()
		returns := make([]wasm.ValueType, numOut)
		for j := 0; j < numIn; j++ {
			returns[j] = goType2WasmType(rType.Out(j).Kind())
		}

		w.eeiFuncSet.entries[i] = wasm.FunctionSig{
			ParamTypes:  args,
			ReturnTypes: returns,
		}

		w.eeiFuncSet.funcs[i] = wasm.Function{
			Sig:  &w.eeiFuncSet.entries[i],
			Body: &wasm.FunctionBody{},
			Host: reflect.ValueOf(f),
		}

		w.eeiFuncSet.exports[k] = wasm.ExportEntry{
			FieldStr: k,
			Kind:     wasm.ExternalFunction,
			Index:    uint32(i),
		}

		i++
	}
}

func (w *WasmIntptr) initDebugModule() {
	w.debugHandlers = make(map[string]interface{})
	w.debugFuncSet = &funcSet{exports: make(map[string]wasm.ExportEntry)}

	dapi := &eeiDebugApi{}
	rdapi := reflect.TypeOf(dapi)
	fnCount := rdapi.NumMethod()
	for i := 0; i < fnCount; i++ {
		fn := rdapi.Method(i)
		w.debugHandlers[fn.Name] = fn
	}

	i := 0
	for k, v := range w.debugHandlers {
		rType := reflect.TypeOf(v)
		numIn := rType.NumIn() - 2
		args := make([]wasm.ValueType, numIn)
		for j := 0; j < numIn; j++ {
			args[i] = goType2WasmType(rType.In(j + 2).Kind())
		}
		numOut := rType.NumOut()
		returns := make([]wasm.ValueType, numOut)
		for j := 0; j < numOut; j++ {
			returns[j] = goType2WasmType(rType.Out(j).Kind())
		}

		w.debugFuncSet.entries[i] = wasm.FunctionSig{
			ParamTypes:  args,
			ReturnTypes: returns,
		}

		w.debugFuncSet.funcs[i] = wasm.Function{
			Sig:  &w.debugFuncSet.entries[i],
			Body: &wasm.FunctionBody{},
			Host: reflect.ValueOf(v),
		}

		w.debugFuncSet.exports[k] = wasm.ExportEntry{
			FieldStr: k,
			Kind:     wasm.ExternalFunction,
			Index:    uint32(i),
		}

		i++
	}

}

func (w *WasmIntptr) GetHandlers() map[string]interface{} {
	return w.handlers
}

func (w *WasmIntptr) GetHandler(name string) interface{} {
	return w.handlers[name]
}

func (w *WasmIntptr) debug() bool {
	return w.evm.vmConfig.Debug
}

func (w *WasmIntptr) useGas(amount uint64) {
	if w.contract == nil {
		panic("contract is nil")
	}

	if amount > w.contract.Gas {
		panic("out of gas")
	}

	w.contract.Gas -= amount
}

func (w *WasmIntptr) StateDB() vm.StateDB {
	return w.evm.StateDB
}

func (w *WasmIntptr) Run(contract *Contract, input []byte) ([]byte, error) {
	w.evm.depth++
	w.contract = contract
	w.contract.Input = input

	defer func() {
		w.evm.depth--
	}()

	module, err := wasm.ReadModule(bytes.NewReader(contract.Code), ModuleResolver(w))
	if err != nil {
		return nil, err
	}

	if module.Start != nil {
		return nil, fmt.Errorf("A contract should not have a start function: found #%d", module.Start.Index)
	}

	vm, err := exec.NewVM(module)
	if err != nil {
		return nil, fmt.Errorf("failed to create vm: %v", err)
	}
	vm.RecoverPanic = true
	w.vm = vm

	for name, entry := range module.Export.Entries {
		if name == "main" && entry.Kind == wasm.ExternalFunction {
			// Check func signature and output types
			sig := module.FunctionIndexSpace[entry.Index].Sig
			if len(sig.ParamTypes) == 0 && len(sig.ReturnTypes) == 0 {
				_, err := vm.ExecCode(int64(entry.Index))
				if err != nil {
					w.terminateType = TerminateInvalid
					w.useGas(w.contract.Gas)
				}
				return w.returnData, err
			}
			break
		}
	}
	return nil, errors.New("could not find a valid 'main' function in the code")
}

// CanRun checks the binary for a WASM header and accepts the binary blob
// if it matches.
func (w *WasmIntptr) CanRun(file []byte) bool {
	// Check the header
	if len(file) <= 8 || string(file[:4]) != "\000asm" {
		return false
	}

	// Check the version
	ver := binary.LittleEndian.Uint32(file[4:])
	if ver != 1 {
		return false
	}

	return true
}

func (w *WasmIntptr) IsReadOnly() bool {
	return w.readonly
}

func (w *WasmIntptr) SetReadOnly(flag bool) {
	w.readonly = flag
}
