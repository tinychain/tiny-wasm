package tinywasm

import (
	"fmt"
	"github.com/tinychain/tiny-wasm/wagon/exec"
	"github.com/tinychain/tinychain/common"
	"github.com/tinychain/tinychain/core/types"
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
	GasSstoreClear        = 5000
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
	writeToMem(p, w.contract.Address().Bytes(), resultOffset)
}

func (*eeiApi) getExternalBalance(p *exec.Process, w *WasmIntptr, addressOffset, resultOffset int32) {
	w.useGas(GasCostBalance)
	addr := loadFromMem(p, addressOffset, common.AddressLength)
	balance := w.evm.StateDB.GetBalance(common.BytesToAddress(addr))
	writeToMem(p, balance.Bytes(), resultOffset)
}

// getBlockHash gets the hash of one of the 256 most recent completed blocks.
func (*eeiApi) getBlockHash(p *exec.Process, w *WasmIntptr, number int64, resultOffset int32) int32 {
	w.useGas(GasCostBlockHash)
	currHeight := w.evm.Context.BlockHeight.Uint64()
	if currHeight > 256 && currHeight-256 > uint64(number) {
		return ErrEEICallFailure
	}
	hash := w.evm.Context.GetHash(uint64(number))
	writeToMem(p, hash.Bytes(), resultOffset)
	return EEICallSuccess

}

func (*eeiApi) call(p *exec.Process, w *WasmIntptr, gas int64, addressOffset, valueOffset, dataOffset, dataLength int32) int32 {
	w.useGas(GasCostCall)

	addr, value, input := getCallParams(p, w, addressOffset, valueOffset, dataOffset, dataLength)

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

	return call(w, toContract, input, snapshot)
}

func (*eeiApi) callDataCopy(p *exec.Process, w *WasmIntptr, resultOffset, dataOffset, length int32) {
	w.useGas(GasCostVeryLow + GasCostCopy*uint64(length))
	writeToMem(p, w.contract.Input[dataOffset:dataOffset+length], resultOffset)
}

func (*eeiApi) getCallDataSize(p *exec.Process, w *WasmIntptr) int32 {
	w.useGas(GasCostBase)
	return int32(len(w.contract.Input))
}

func (*eeiApi) callCode(p *exec.Process, w *WasmIntptr, gas int64, addressOffset, valueOffset, dataOffset, dataLength int32) int32 {
	w.useGas(GasCostCall)
	addr, value, input := getCallParams(p, w, addressOffset, valueOffset, dataOffset, dataLength)

	if !w.evm.Context.CanTransfer(w.StateDB(), addr, value) {
		fmt.Printf("balance not enough: want to use %v, got %v\n", value, w.StateDB().GetBalance(addr))
	}

	snapshot := w.StateDB().Snapshot()
	toContract := NewContract(w.contract.caller, w.contract.caller, value, uint64(gas))
	toContract.SetCallCode(&addr, w.StateDB().GetCodeHash(addr), w.StateDB().GetCode(addr))

	return call(w, toContract, input, snapshot)
}

func (*eeiApi) callDelegate(p *exec.Process, w *WasmIntptr, gas int64, addressOffset, dataOffset, dataLength int32) int32 {
	w.useGas(GasCostCall)
	addr, _, input := getCallParams(p, w, addressOffset, -1, dataOffset, dataLength)

	snapshot := w.StateDB().Snapshot()
	toContract := NewContract(w.contract.caller, w.contract.caller, nil, uint64(gas)).AsDelegate()
	toContract.SetCallCode(&addr, w.StateDB().GetCodeHash(addr), w.StateDB().GetCode(addr))

	return call(w, toContract, input, snapshot)
}

func (*eeiApi) callStatic(p *exec.Process, w *WasmIntptr, gas int64, addressOffset, dataOffset, dataLength int32) int32 {
	w.useGas(GasCostCall)
	addr, _, input := getCallParams(p, w, addressOffset, -1, dataOffset, dataLength)

	if !w.IsReadOnly() {
		w.SetReadOnly(true)
		defer func() { w.SetReadOnly(false) }()
	}

	toContract := NewContract(w.contract.caller, AccountRef(addr), new(big.Int), uint64(gas))
	toContract.SetCallCode(&addr, w.StateDB().GetCodeHash(addr), w.StateDB().GetCode(addr))

	return call(w, toContract, input, w.StateDB().Snapshot())
}

func (*eeiApi) storageStore(p *exec.Process, w *WasmIntptr, pathOffset, valueOffset int32) {
	if w.IsReadOnly() {
		panic("Static mode violation in storageStore")
	}

	key := common.BytesToHash(loadFromMem(p, pathOffset, u256Len))
	val := loadFromMem(p, pathOffset, u256Len)

	oldVal := w.StateDB().GetState(w.contract.Address(), key)

	// This checks for 3 scenario's and calculates gas accordingly:
	//
	// 1. From a zero-value address to a non-zero value         (NEW VALUE)
	// 2. From a non-zero value address to a zero-value address (DELETE)
	// 3. From a non-zero to a non-zero                         (CHANGE)
	switch {
	case oldVal == nil && new(big.Int).SetBytes(val).Sign() != 0: // 0 => non 0
		w.useGas(GasCostSSet)
	case oldVal != nil && new(big.Int).SetBytes(val).Sign() == 0: // non 0 => 0
		w.useGas(GasSstoreClear)
	default: // non 0 => non 0 (or 0 => 0)
		w.useGas(GasCostSReset)
	}

	w.StateDB().SetState(w.contract.Address(), key, val)
}

func (*eeiApi) storageLoad(p *exec.Process, w *WasmIntptr, pathOffset, resultOffset int32) {
	w.useGas(GasCostSLoad)
	key := common.BytesToHash(loadFromMem(p, pathOffset, u256Len))
	val := w.StateDB().GetState(w.contract.Address(), key)
	writeToMem(p, val, resultOffset)
}

func (*eeiApi) getCaller(p *exec.Process, w *WasmIntptr, resultOffset int32) {
	w.useGas(GasCostBase)
	addr := w.contract.CallerAddress
	writeToMem(p, addr.Bytes(), resultOffset)
}

func (*eeiApi) getCallValue(p *exec.Process, w *WasmIntptr, resultOffset int32) {
	w.useGas(GasCostBase)
	writeToMem(p, w.contract.Value().Bytes(), resultOffset)
}

func (*eeiApi) codeCopy(p *exec.Process, w *WasmIntptr, resultOffset, codeOffset, length int32) {
	w.useGas(GasCostVeryLow + GasCostCopy*uint64(length))
	writeToMem(p, w.contract.Code[codeOffset:codeOffset+length], resultOffset)
}

func (*eeiApi) getCodeSize(p *exec.Process, w *WasmIntptr) int32 {
	w.useGas(GasCostBase)
	return int32(len(w.contract.Code))
}

func (*eeiApi) getBlockCoinbase(p *exec.Process, w *WasmIntptr, resultOffset int32) {
	w.useGas(GasCostBase)
	writeToMem(p, w.evm.Coinbase().Bytes(), resultOffset)
}

func (*eeiApi) create(p *exec.Process, w *WasmIntptr, valueOffset, dataOffset, length, resultOffset int32) int32 {
	w.useGas(GasCostCreate)

	oldVM := w.vm
	oldContract := w.contract
	defer func() {
		w.vm = oldVM
		w.contract = oldContract
	}()

	w.terminateType = TerminateInvalid

	if int(valueOffset)+u128Len > len(w.vm.Memory()) {
		return ErrEEICallFailure
	}

	if int(dataOffset+length) > len(w.vm.Memory()) {
		return ErrEEICallFailure
	}

	code := loadFromMem(p, dataOffset, length)
	val := loadFromMem(p, valueOffset, u128Len)

	w.terminateType = TerminateFinish

	// EIP150 says that the calling contract should keep 1/64th of the
	// leftover gas.
	gas := w.contract.Gas - w.contract.Gas/64
	_, addr, leftGas, _ := w.evm.Create(w.contract, code, gas, new(big.Int).SetBytes(val))

	switch w.terminateType {
	case TerminateFinish:
		oldContract.Gas += leftGas
		p.WriteAt(addr.Bytes(), int64(resultOffset))
		return EEICallSuccess
	case TerminateRevert:
		oldContract.Gas += gas
		return ErrEEICallRevert
	default:
		oldContract.Gas += leftGas
		return ErrEEICallFailure
	}
}

func (*eeiApi) getBlockDifficulty(p *exec.Process, w *WasmIntptr, resultOffset int32) {

}

func (*eeiApi) externalCodeCopy(p *exec.Process, w *WasmIntptr, addressOffset, resultOffset, codeOffset, length int32) {
	addr := common.BytesToAddress(loadFromMem(p, addressOffset, common.AddressLength))
	code := w.StateDB().GetCode(addr)

	w.useGas(GasCostVeryLow + GasCostCopy*uint64(len(code)))
	writeToMem(p, code[codeOffset:codeOffset+length], resultOffset)
}

func (*eeiApi) getExternalCodeSize(p *exec.Process, w *WasmIntptr, addressOffset int32) int32 {
	w.useGas(GasCostExtCode)
	addr := common.BytesToAddress(loadFromMem(p, addressOffset, common.AddressLength))
	return int32(w.StateDB().GetCodeSize(addr))
}

func (*eeiApi) getGasLeft(p *exec.Process, w *WasmIntptr) int64 {
	w.useGas(GasCostBase)
	return int64(w.contract.Gas)
}

func (*eeiApi) getBlockGasLimit(p *exec.Process, w *WasmIntptr) int64 {
	w.useGas(GasCostBase)
	return int64(w.evm.GasLimit)
}

func (*eeiApi) getTxGasPrice(p *exec.Process, w *WasmIntptr, valueOffset int32) {
	w.useGas(GasCostBase)
	writeToMem(p, w.evm.GasPrice.Bytes(), valueOffset)
}

func (*eeiApi) log(p *exec.Process, w *WasmIntptr, dataOffset, dataLength, numberOfTopics, topic1, topic2, topic3, topic4 int32) {
	w.useGas(GasCostLog + GasCostLogData*uint64(dataLength) + GasCostLogTopic*uint64(numberOfTopics))

	if numberOfTopics > 4 || numberOfTopics < 0 {
		w.terminateType = TerminateInvalid
		p.Terminate()
	}

	data := loadFromMem(p, dataOffset, dataLength)
	topics := make([]common.Hash, numberOfTopics)

	switch numberOfTopics {
	case 4:
		topics[3] = common.BytesToHash(loadFromMem(p, topic4, u256Len))
		fallthrough
	case 3:
		topics[2] = common.BytesToHash(loadFromMem(p, topic3, u256Len))
		fallthrough
	case 2:
		topics[1] = common.BytesToHash(loadFromMem(p, topic2, u256Len))
		fallthrough
	case 1:
		topics[0] = common.BytesToHash(loadFromMem(p, topic1, u256Len))
	default:
		return
	}

	w.StateDB().AddLog(&types.Log{
		Address:     w.contract.Address(),
		Topics:      topics,
		Data:        data,
		BlockHeight: w.evm.BlockHeight.Uint64(),
	})
}

func (*eeiApi) getBlockNumber(p *exec.Process, w *WasmIntptr) int64 {
	w.useGas(GasCostBase)
	return w.evm.BlockHeight.Int64()
}

func (*eeiApi) getTxOrigin(p *exec.Process, w *WasmIntptr, resultOffset int32) {
	w.useGas(GasCostBase)
	writeToMem(p, w.evm.Origin.Bytes(), resultOffset)
}

func (*eeiApi) finish(p *exec.Process, w *WasmIntptr, dataOffset, length int32) {
	w.returnData = loadFromMem(p, dataOffset, length)
	w.terminateType = TerminateFinish
	p.Terminate()
}

func (*eeiApi) revert(p *exec.Process, w *WasmIntptr, dataOffset, length int32) {
	w.returnData = loadFromMem(p, dataOffset, length)
	w.terminateType = TerminateRevert
	p.Terminate()
}

func (*eeiApi) getReturnDataSize(p *exec.Process, w *WasmIntptr) int32 {
	w.useGas(GasCostBase)
	return int32(len(w.returnData))
}

func (*eeiApi) returnDataCopy(p *exec.Process, w *WasmIntptr, resultOffset, dataOffset, length int32) {
	w.useGas(GasCostCopy * uint64(length))
	writeToMem(p, w.returnData[dataOffset:dataOffset+length], resultOffset)
}

func (*eeiApi) selfDestruct(p *exec.Process, w *WasmIntptr, addressOffset int32) {
	addr := common.BytesToAddress(loadFromMem(p, addressOffset, common.AddressLength))
	balance := w.StateDB().GetBalance(w.contract.Address())

	totalGas := GasCostSuicide
	// If the target address dose not exist, add the account creation cost
	if !w.StateDB().Exist(addr) {
		totalGas += GasCostCreateBySuicide
	}
	w.StateDB().AddBalance(addr, balance)
	w.useGas(uint64(totalGas))
	w.StateDB().Suicide(w.contract.Address())

	w.terminateType = TerminateSuicide
	p.Terminate()
}

func (*eeiApi) getBlockTimestamp(p *exec.Process, w *WasmIntptr) int64 {
	w.useGas(GasCostBase)
	return w.evm.Time.Int64()
}

// swapEndian swap big endian to little endian or reverse.
func swapEndian(src []byte) []byte {
	rect := make([]byte, len(src))
	for i, b := range src {
		rect[len(src)-i-1] = b
	}
	return rect
}

func loadFromMem(p *exec.Process, offset int32, size int32) []byte {
	b := make([]byte, size)
	p.ReadAt(b, int64(offset))
	return swapEndian(b)
}

func writeToMem(p *exec.Process, data []byte, offset int32) (int, error) {
	return p.WriteAt(swapEndian(data), int64(offset))
}

func getCallParams(p *exec.Process, w *WasmIntptr, addressOffset, valueOffset, dataOffset, dataLength int32) (addr common.Address, value *big.Int, input []byte) {
	// Get the address from mem
	addr = common.BytesToAddress(loadFromMem(p, addressOffset, common.AddressLength))

	// Get the value from mem
	if valueOffset == -1 {
		value = w.contract.value
	} else {
		value = big.NewInt(0).SetBytes(loadFromMem(p, valueOffset, u128Len))
	}
	if value.Cmp(big.NewInt(0)) != 0 {
		w.useGas(GasCostCallValue)
	}

	// Get the input data from mem
	input = loadFromMem(p, dataOffset, dataLength)

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

	// Check terminateType from execution
	switch w.terminateType {
	case TerminateFinish:
		return EEICallSuccess
	case TerminateRevert:
		w.StateDB().RevertToSnapshot(snapshot)
		return ErrEEICallRevert
	default:
		w.StateDB().RevertToSnapshot(snapshot)
		w.useGas(w.contract.Gas)
		return ErrEEICallFailure
	}
}
