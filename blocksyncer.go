package blocksyncer

type FractalBitcoinService interface {
	GetBlockHeadersAndBlocks() (int32, int32) // return `headers` and `blocks` fields from the `getblockchaininfo` RPC call
	Shutdown()
}
