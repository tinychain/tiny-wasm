package tiny_wasm

import (
	"github.com/tinychain/tinychain/core/chain"
	"github.com/tinychain/tinychain/core/vm"
)

type EnvContext interface {
	StateDB() vm.StateDB
	Chain() *chain.Blockchain
}
