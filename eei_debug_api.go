package tiny_wasm

import "github.com/tinychain/tiny-wasm/wagon/exec"

type eeiDebugApi struct{}

func (*eeiDebugApi) printMem(p *exec.Process, w *WasmIntptr, offset, len int32) {

}

func (*eeiDebugApi) printMemHex(p *exec.Process, w *WasmIntptr, offset, len int32) {

}

func (*eeiDebugApi) printStorage(p *exec.Process, w *WasmIntptr, pathOffset int32) {

}

func (*eeiDebugApi) printStorageHex(p *exec.Process, w *WasmIntptr, pathOffset int32) {

}
