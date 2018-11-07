package tiny_wasm

import "github.com/tinychain/tiny-wasm/wagon/exec"

const (
	// EEICallSuccess is the return value in case of a successful contract execution
	EEICallSuccess = 0
	// ErrEEICallFailure is the return value in case of a contract execution failture
	ErrEEICallFailure = 1
	// ErrEEICallRevert is the return value in case a contract calls `revert`
	ErrEEICallRevert = 2
)

// List of gas costs
const (
	GasCostZero           = 0
	GasCostBase           = 2
	GasCostVeryLow        = 3
	GasCostLow            = 5
	GasCostMid            = 8
	GasCostHigh           = 10
	GasCostExtCode        = 700
	GasCostBalance        = 400
	GasCostSLoad          = 200
	GasCostJumpDest       = 1
	GasCostSSet           = 20000
	GasCostSReset         = 5000
	GasRefundSClear       = 15000
	GasRefundSelfDestruct = 24000
	GasCostCreate         = 32000
	GasCostCall           = 700
	GasCostCallValue      = 9000
	GasCostCallStipend    = 2300
	GasCostLog            = 375
	GasCostLogData        = 8
	GasCostLogTopic       = 375
	GasCostCopy           = 3
	GasCostBlockHash      = 800
)

type eeiApi struct{}

func (*eeiApi) useGas(p *exec.Process, w *WasmIntptr, amount int64) {

}

func (*eeiApi) getAddress(p *exec.Process, w *WasmIntptr, resultOffset int32) {

}

func (*eeiApi) getExternalBalance(p *exec.Process, w *WasmIntptr, addressOffset, resultOffset int32) {

}

func (*eeiApi) getBlockHash(p *exec.Process, w *WasmIntptr, resultOffset int32) int32 {

}

func (*eeiApi) call(p *exec.Process, w *WasmIntptr, gas int64, addressOffset, valueOffset, dataOffset int32, dataLength int32) int32 {

}

func (*eeiApi) callDataCopy(p *exec.Process, w *WasmIntptr, resultOffset, dataOffset, length int32) {

}

func (*eeiApi) getCallDataSize(p *exec.Process, w *WasmIntptr) int32 {

}

func (*eeiApi) callCode(p *exec.Process, w *WasmIntptr, gas int64, addressOffset, valueOffset, dataOffset, dataLength int32) int32 {

}

func (*eeiApi) callDelegate(p *exec.Process, w *WasmIntptr, gas int64, addressOffset, dataOffset, dataLength int32) int32 {

}

func (*eeiApi) callStatic(p *exec.Process, w *WasmIntptr, gas int64, addressOffset, dataOffset, dataLength int32) int32 {

}

func (*eeiApi) storageStore(p *exec.Process, w *WasmIntptr, pathOffset, valueOffset int32) {

}

func (*eeiApi) storageLoad(p *exec.Process, w *WasmIntptr, pathOffset, resultOffset int32) {

}

func (*eeiApi) getCaller(p *exec.Process, w *WasmIntptr, resultOffset int32) {

}

func (*eeiApi) getCallValue(p *exec.Process, w *WasmIntptr, resultOffset int32) {

}

func (*eeiApi) codeCopy(p *exec.Process, w *WasmIntptr, resultOffset, codeOffset, length int32) {

}

func (*eeiApi) getCodeSize(p *exec.Process, w *WasmIntptr) int32 {

}

func (*eeiApi) getBlockCoinbase(p *exec.Process, w *WasmIntptr, resultOffset int32) {

}

func (*eeiApi) create(p *exec.Process, w *WasmIntptr, valueOffset, dataOffset, length, resultOffset int32) {

}

func (*eeiApi) getBlockDifficulty(p *exec.Process, w *WasmIntptr, resultOffset int32) {

}

func (*eeiApi) externalCodeCopy(p *exec.Process, w *WasmIntptr, addressOffset, resultOffset, codeOffset, length int32) {

}

func (*eeiApi) getExternalCodeSize(p *exec.Process, w *WasmIntptr, addressOffset int32) int32 {

}

func (*eeiApi) getGasLeft(p *exec.Process, w *WasmIntptr) int64 {

}

func (*eeiApi) getBlockGasLimit(p *exec.Process, w *WasmIntptr) int64 {

}

func (*eeiApi) getTxGasPrice(p *exec.Process, w *WasmIntptr, valueOffset int32) {

}

func (*eeiApi) log(p *exec.Process, w *WasmIntptr, dataOffset, length, numberOfTopics, topic1, topic2, topic3, topic4 int32) {

}

func (*eeiApi) getBlockNumber(p *exec.Process, w *WasmIntptr) int64 {

}

func (*eeiApi) getTxOrigin(p *exec.Process, w *WasmIntptr, resultOffset int32) {

}

func (*eeiApi) finish(p *exec.Process, w *WasmIntptr, dataOffset, length int32) {

}

func (*eeiApi) revert(p *exec.Process, w *WasmIntptr, dataOffset, length int32) {

}

func (*eeiApi) getReturnDataSize(p *exec.Process, w *WasmIntptr) int32 {

}

func (*eeiApi) returnDataCopy(p *exec.Process, w *WasmIntptr, resultOffset, dataOffset, length int32) {

}

func (*eeiApi) selfDestruct(p *exec.Process, w *WasmIntptr, addressOffset int32) {

}

func (*eeiApi) getBlockTimestamp(p *exec.Process, w *WasmIntptr) int64 {

}
