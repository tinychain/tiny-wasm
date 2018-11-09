package tiny_wasm

import (
	"bytes"
	"encoding/binary"
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

type WasmIntptr struct {
	// execution fields
	vm            *exec.VM
	contract      *Contract
	readonly      bool          // static mode
	evm           *EVM          // evm instance
	terminateType TerminateType // termination type of the execution
	returnData    []byte        // returning output data for the execution

	// module resolver components
	handlers map[string]interface{} // eei function handlers
	entries  []wasm.FunctionSig
	funcs    []wasm.Function
	exports  map[string]wasm.ExportEntry
}

func NewWasmIntptr(evm *EVM) *WasmIntptr {
	w := &WasmIntptr{
		evm:      evm,
		handlers: make(map[string]interface{}),
	}

	w.init()

	return w
}

func (w *WasmIntptr) init() {
	// eei function register
	api := &eeiApi{}
	rapi := reflect.TypeOf(api)
	fnCount := rapi.NumMethod()
	for i := 0; i < fnCount; i++ {
		fn := rapi.Method(i)
		w.register(fn.Name, fn)
	}

	w.initModule()
}

func (w *WasmIntptr) initModule() {
	i := 0
	for k, f := range w.handlers {
		rType := reflect.TypeOf(f)
		numIn := rType.NumIn() - 1
		args := make([]wasm.ValueType, numIn)
		for j := 0; j < numIn; j++ {
			args[j] = goType2WasmType(rType.In(j + 1).Kind())
		}

		numOut := rType.NumOut()
		returns := make([]wasm.ValueType, numOut)
		for j := 0; j < numIn; j++ {
			returns[j] = goType2WasmType(rType.Out(j).Kind())
		}

		w.entries[i] = wasm.FunctionSig{
			ParamTypes:  args,
			ReturnTypes: returns,
		}

		w.funcs[i] = wasm.Function{
			Sig:  &w.entries[i],
			Body: &wasm.FunctionBody{},
			Host: reflect.ValueOf(f),
		}

		w.exports[k] = wasm.ExportEntry{
			FieldStr: k,
			Kind:     wasm.ExternalFunction,
			Index:    uint32(i),
		}

		i++
	}
}

func (w *WasmIntptr) register(name string, handler interface{}) {
	w.handlers[name] = handler
}

func (w *WasmIntptr) GetHandlers() map[string]interface{} {
	return w.handlers
}

func (w *WasmIntptr) GetHandler(name string) interface{} {
	return w.handlers[name]
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

func (w *WasmIntptr) BlockHeight() uint64 {

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

	// TODO: input needs to specify action name

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
