package tiny_wasm

import (
	"fmt"
	"github.com/tinychain/tiny-wasm/wagon/exec"
	"github.com/tinychain/tinychain/common"
	"math/big"
)

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

	GasCostExtcodeSize = 700
	GasCostExtcodeCopy = 700
	GasCostExtcodeHash = 400
	GasCostCalls       = 700
	GasCostSuicide     = 5000
	GasCostExpByte     = 50

	GasCostCreateBySuicide = 25000
)

type eeiApi struct{}

func (*eeiApi) useGas(p *exec.Process, w *WasmIntptr, amount int64) {
	w.useGas(uint64(amount))
}

func (*eeiApi) getAddress(p *exec.Process, w *WasmIntptr, resultOffset int32) {
	w.useGas(GasCostBase)
	p.WriteAt(w.contract.Address().Bytes(), int64(resultOffset))
}

func (*eeiApi) getExternalBalance(p *exec.Process, w *WasmIntptr, addressOffset, resultOffset int32) {
	w.useGas(GasCostBalance)
	addr := loadFromMem(p, addressOffset, common.AddressLength)
	balance := w.evm.StateDB.GetBalance(common.BytesToAddress(addr))
	p.WriteAt(balance.Bytes(), int64(resultOffset))
}

// getBlockHash gets the hash of one of the 256 most recent completed blocks.
func (*eeiApi) getBlockHash(p *exec.Process, w *WasmIntptr, number int64, resultOffset int32) int32 {
	w.useGas(GasCostBlockHash)
	currHeight := w.evm.Context.BlockHeight.Uint64()
	if currHeight > 256 && currHeight-256 > uint64(number) {
		return ErrEEICallFailure
	}
	hash := w.evm.Context.GetHash(uint64(number))
	p.WriteAt(hash.Bytes(), int64(resultOffset))
	return EEICallSuccess

}

func (*eeiApi) call(p *exec.Process, w *WasmIntptr, gas int64, addressOffset, valueOffset, dataOffset, dataLength int32) int32 {
	w.useGas(GasCostCall)

	addr, value, input := getCallParams(p, addressOffset, valueOffset, dataOffset, dataLength)

	if !w.evm.Context.CanTransfer(w.StateDB(), addr, value) {
		fmt.Printf("balance not enough: want to use %v, got %v\n", value, w.StateDB().GetBalance(addr))
	}

	if !w.StateDB().Exist(addr) {
		w.StateDB().CreateAccount(addr)
	}

	snapshot := w.StateDB().Snapshot()
	// Transfer value
	w.evm.Transfer(w.StateDB(), w.contract.caller.Address(), addr, value)

	// Load the contract in a new VM
	toContract := NewContract(w.contract.caller, AccountRef(addr), value, uint64(gas))
	toContract.SetCallCode(&addr, w.StateDB().GetCodeHash(addr), w.StateDB().GetCode(addr))

	return call(w, toContract, input, value, snapshot, gas)
}

func (*eeiApi) callDataCopy(p *exec.Process, w *WasmIntptr, resultOffset, dataOffset, length int32) {
	w.useGas(GasCostCopy * uint64(length))
	p.WriteAt(w.contract.Input[dataOffset:dataOffset+length], int64(resultOffset))
}

func (*eeiApi) getCallDataSize(p *exec.Process, w *WasmIntptr) int32 {
	w.useGas(GasCostBase)
	return int32(len(w.contract.Input))
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

func loadFromMem(p *exec.Process, offset int32, size int) []byte {
	b := make([]byte, size)
	p.ReadAt(b, int64(offset))
	return b
}

func getCallParams(p *exec.Process, w *WasmIntptr, addressOffset, valueOffset, dataOffset, dataLength int32) (addr common.Address, value *big.Int, input []byte) {
	// Get the address from mem
	addr = common.BytesToAddress(loadFromMem(p, addressOffset, common.AddressLength))

	// Get the value from mem
	value = big.NewInt(0).SetBytes(loadFromMem(p, valueOffset, u128Len))
	if value.Cmp(big.NewInt(0)) != 0 {
		w.useGas(GasCostCallValue)
	}

	// Get the input data from mem
	input = loadFromMem(p, dataOffset, int(dataLength))

	return
}

func call(w *WasmIntptr, toContract *Contract, input []byte, snapshot int) int32 {
	if w.evm.depth > maxCallDepth {
		// Clear all gas of contract
		w.useGas(w.contract.Gas)
		return ErrEEICallFailure
	}

	beforeVM := w.vm
	beforeContract := w.contract

	_, err := w.Run(toContract, input)

	w.vm = beforeVM
	w.contract = beforeContract

	if err != nil {
		w.StateDB().RevertToSnapshot(snapshot)
		// TODO: need to clear all gas?
		return ErrEEICallFailure
	}

	// TODO: termination type
	return EEICallSuccess
}
