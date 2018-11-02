package tiny_wasm

import "github.com/tinychain/tiny-wasm/wagon/wasm"

var eeiFunctionList = []string{
	"useGas",
	"getAddress",
	"getExternalBalance",
	"getBlockHash",
	"call",
	"callDataCopy",
	"getCallDataSize",
	"callCode",
	"callDelegate",
	"callStatic",
	"storageStore",
	"storageLoad",
	"getCaller",
	"getCallValue",
	"codeCopy",
	"getCodeSize",
	"getBlockCoinbase",
	"create",
	"getBlockDifficulty",
	"externalCodeCopy",
	"getExternalCodeSize",
	"getGasLeft",
	"getBlockGasLimit",
	"getTxGasPrice",
	"log",
	"getBlockNumber",
	"getTxOrigin",
	"finish",
	"revert",
	"getReturnDataSize",
	"returnDataCopy",
	"selfDestruct",
	"getBlockTimestamp",
}

func moduleResolver(w *WasmVM, name string) (*wasm.Module, error) {

}

func ModuleResolver(w *WasmVM) wasm.ResolveFunc {
	return func(name string) (*wasm.Module, error) {
		return moduleResolver(w, name)
	}
}
