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

// funcSet wraps the necessary fields of an importing wasm module
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
	handlers      map[string]reflect.Value // eei function handlers
	debugHandlers map[string]reflect.Value // debug function handlers
	eeiFuncSet    *funcSet                 // eei function set
	debugFuncSet  *funcSet                 // debug function set

	// meter
	metering bool
}

func NewWasmIntptr(evm *EVM) *WasmIntptr {
	w := &WasmIntptr{
		evm:        evm,
		handlers:   make(map[string]reflect.Value),
		eeiFuncSet: &funcSet{exports: make(map[string]wasm.ExportEntry)},
	}

	w.initEEIModule()
	if w.debug() {
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
		w.handlers[fn.Name] = fn.Func
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
			Host: f,
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
	w.debugHandlers = make(map[string]reflect.Value)
	w.debugFuncSet = &funcSet{exports: make(map[string]wasm.ExportEntry)}

	dapi := &eeiDebugApi{}
	rdapi := reflect.TypeOf(dapi)
	fnCount := rdapi.NumMethod()
	for i := 0; i < fnCount; i++ {
		fn := rdapi.Method(i)
		w.debugHandlers[fn.Name] = fn.Func
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
			Host: v,
		}

		w.debugFuncSet.exports[k] = wasm.ExportEntry{
			FieldStr: k,
			Kind:     wasm.ExternalFunction,
			Index:    uint32(i),
		}

		i++
	}

}

func (w *WasmIntptr) GetHandlers() map[string]reflect.Value {
	return w.handlers
}

func (w *WasmIntptr) GetHandler(name string) (reflect.Value, bool) {
	val, ok := w.handlers[name]
	return val, ok
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

	mainIndex, err := w.verifyModule(module)
	if err != nil {
		return nil, err
	}

	vm, err := exec.NewVM(module)
	if err != nil {
		return nil, fmt.Errorf("failed to create vm: %v", err)
	}
	vm.RecoverPanic = true
	w.vm = vm

	sig := module.FunctionIndexSpace[mainIndex].Sig
	if len(sig.ParamTypes) == 0 && len(sig.ReturnTypes) == 0 {
		_, err := vm.ExecCode(int64(mainIndex))
		if err != nil {
			w.terminateType = TerminateInvalid
		}

		if w.StateDB().HasSuicided(contract.Address()) {
			err = nil
		}
		return w.returnData, err
	}

	w.terminateType = TerminateInvalid
	return nil, errors.New("could not find a valid 'main' function in the code")
}

// verifyModule validates the wasm module resolved by the wagon, check `main` and `memory`
// export and import valid `eei` api. It returns the index of `main` export function and an error.
func (w *WasmIntptr) verifyModule(m *wasm.Module) (int, error) {
	if m.Start != nil {
		return -1, fmt.Errorf("A contract should not have a start function: found #%d", m.Start.Index)
	}

	if m.Export == nil {
		return -1, fmt.Errorf("module has no exports `main` and `memory`")
	}

	if c := len(m.Export.Entries); c != 2 {
		return -1, fmt.Errorf("module has %d exports instead of 2", c)
	}

	// Check the existence of the `main` and `memory` exports
	mainIndex := -1
	for name, entry := range m.Export.Entries {
		if name == "main" {
			if entry.Kind != wasm.ExternalFunction {
				return -1, fmt.Errorf("`main` is not a function in module")
			}
			mainIndex = int(entry.Index)
		} else if name == "memory" {
			if entry.Kind != wasm.ExternalMemory {
				return -1, fmt.Errorf("`memory` is not a memory in module")
			}
		}
	}

	// Validate whether the function imported from `ethereum` module are in the list of eei_api or not
	if m.Import != nil {
		for _, entry := range m.Import.Entries {
			if entry.ModuleName == "ethereum" && entry.Type.Kind() == wasm.ExternalFunction {
				if _, exist := w.GetHandler(entry.FieldName); !exist {
					return -1, fmt.Errorf("%s is not found in eei api list", entry.FieldName)
				}
			}
		}
	}

	return mainIndex, nil
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
