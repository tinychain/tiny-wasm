package tiny_wasm

func useGas(w *WasmVM, amount int64) {

}

func getAddress(w *WasmVM, resultOffset int32) {

}

func getExternalBalance(w *WasmVM, addressOffset, resultOffset int32) {

}

func getBlockHash(w *WasmVM, resultOffset int32) int32 {

}

func call(w *WasmVM, gas int64, addressOffset, valueOffset, dataOffset int32, dataLength int32) int32 {

}

func callDataCopy(w *WasmVM, resultOffset, dataOffset, length int32) {

}

func getCallDataSize(w *WasmVM) int32 {

}

func callCode(w *WasmVM, gas int64, addressOffset, valueOffset, dataOffset, dataLength int32) int32 {

}

func callDelegate(w *WasmVM, gas int64, addressOffset, dataOffset, dataLength int32) int32 {

}

func callStatic(w *WasmVM, gas int64, addressOffset, dataOffset, dataLength int32) int32 {

}

func storageStore(w *WasmVM, pathOffset, valueOffset int32) {

}

func storageLoad(w *WasmVM, pathOffset, resultOffset int32) {

}

func getCaller(w *WasmVM, resultOffset int32) {

}

func getCallValue(w *WasmVM, resultOffset int32) {

}

func codeCopy(w *WasmVM, resultOffset, codeOffset, length int32) {

}

func getCodeSize(w *WasmVM) int32 {

}

func getBlockCoinbase(w *WasmVM, resultOffset int32) {

}

func create(w *WasmVM, valueOffset, dataOffset, length, resultOffset int32) {

}

func getBlockDifficulty(w *WasmVM, resultOffset int32) {

}

func externalCodeCopy(w *WasmVM, addressOffset, resultOffset, codeOffset, length int32) {

}

func getExternalCodeSize(w *WasmVM, addressOffset int32) int32 {

}

func getGasLeft(w *WasmVM) int64 {

}

func getBlockGasLimit(w *WasmVM) int64 {

}

func getTxGasPrice(w *WasmVM, valueOffset int32) {

}

func log(w *WasmVM, dataOffset, length, numberOfTopics, topic1, topic2, topic3, topic4 int32) {

}

func getBlockNumber(w *WasmVM) int64 {

}

func getTxOrigin(w *WasmVM, resultOffset int32) {

}

func finish(w *WasmVM, dataOffset, length int32) {

}

func revert(w *WasmVM, dataOffset, length int32) {

}

func getReturnDataSize(w *WasmVM) int32 {

}

func returnDataCopy(w *WasmVM, resultOffset, dataOffset, length int32) {

}

func selfDestruct(w *WasmVM, addressOffset int32) {

}

func getBlockTimestamp(w *WasmVM) int64 {

}
