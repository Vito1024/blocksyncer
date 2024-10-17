package blocksyncer

import (
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
)

type (
	Block = wire.MsgBlock
	Hash  = chainhash.Hash
)
