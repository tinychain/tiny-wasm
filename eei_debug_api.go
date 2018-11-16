package tinywasm

import (
	"fmt"
	"github.com/tinychain/tiny-wasm/wagon/exec"
	"github.com/tinychain/tinychain/common"
)

type eeiDebugApi struct{}

func (*eeiDebugApi) print32(p *exec.Process, w *WasmIntptr, value int32) {
	fmt.Println(value)
}

func (*eeiDebugApi) print64(p *exec.Process, w *WasmIntptr, value int64) {
	fmt.Println(value)
}

func (*eeiDebugApi) printMem(p *exec.Process, w *WasmIntptr, offset, len int32) {
	fmt.Println(loadFromMem(p, offset, len))
}

func (*eeiDebugApi) printMemHex(p *exec.Process, w *WasmIntptr, offset, len int32) {
	fmt.Println(common.Hex(loadFromMem(p, offset, len)))
}

func (*eeiDebugApi) printStorage(p *exec.Process, w *WasmIntptr, pathOffset int32) {
	key := common.BytesToHash(loadFromMem(p, pathOffset, u256Len))
	fmt.Println(w.StateDB().GetState(w.contract.Address(), key))
}

func (*eeiDebugApi) printStorageHex(p *exec.Process, w *WasmIntptr, pathOffset int32) {
	key := common.BytesToHash(loadFromMem(p, pathOffset, u256Len))
	fmt.Println(common.Hex(w.StateDB().GetState(w.contract.Address(), key)))
}
