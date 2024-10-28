package fractalbitcoin

import (
	bs "blocksyncer"
	"blocksyncer/config"
	"encoding/json"
	"fmt"
	"log"

	"github.com/btcsuite/btcd/rpcclient"
)

type Service struct {
	config    config.Node
	rpcClient *rpcclient.Client
}

func New(config config.Node) *Service {
	svc := &Service{
		config: config,
	}
	rpcClient, err := rpcclient.New(
		&rpcclient.ConnConfig{
			Host:         fmt.Sprintf("%s:%d", config.RPCHost, config.RPCPort),
			User:         config.RPCUser,
			Pass:         config.RPCPassword,
			HTTPPostMode: true,
			DisableTLS:   true,
		},
		nil,
	)
	if err != nil {
		log.Panicf("failed to create rpc client, error: %v", err)
	}
	err = rpcClient.Ping()
	if err != nil {
		log.Panicf("failed to ping rpc client, error: %v", err)
	}
	svc.rpcClient = rpcClient

	log.Printf("successfully created rpc client for nodeid: %d, nodename: %s\n", svc.config.ID, svc.config.Name)
	return svc
}

func (svc *Service) GetBlockHeadersAndBlocks() (int64, int64) {
	blockchainInfo, err := svc.rpcClient.GetBlockChainInfo()
	if err != nil {
		log.Panicf("failed to get block chain info, error: %v, nodename: %s", err, svc.GetNodeName())
	}

	return int64(blockchainInfo.Headers), int64(blockchainInfo.Blocks)
}

func (svc *Service) GetBlockHashByHeight(height int64) *bs.Hash {
	hash, err := svc.rpcClient.GetBlockHash(height)
	if err != nil {
		log.Panicf("failed to get block hash by height, error: %v, nodename: %s", err, svc.GetNodeName())
	}
	return hash
}

func (svc *Service) GetBlockByHeight(height int64) *bs.Block {
	hash, err := svc.rpcClient.GetBlockHash(height)
	if err != nil {
		log.Panicf("failed to get block hash to get block by height, error: %v, nodename: %s", err, svc.GetNodeName())
	}
	block, err := svc.rpcClient.GetBlock(hash)
	if err != nil {
		log.Panicf("failed to get block, error: %v", err)
	}

	return block
}

func (svc *Service) GetBlockHexByHeight(height int64) []byte {
	hash, err := svc.rpcClient.GetBlockHash(height)
	if err != nil {
		log.Panicf("failed to get block hex by height, error: %v, nodename: %s", err, svc.GetNodeName())
	}

	hashJSON, err := json.Marshal(hash)
	if err != nil {
		log.Panicf("failed to marshal block hash, error: %v, nodename: %s", err, svc.GetNodeName())
	}
	verboseJSON, err := json.Marshal(false)
	if err != nil {
		log.Panicf("failed to marshal verbose, error: %v, nodename: %s", err, svc.GetNodeName())
	}
	auxpowJSON, err := json.Marshal(true)
	if err != nil {
		log.Panicf("failed to marshal auxpow, error: %v, nodename: %s", err, svc.GetNodeName())
	}

	block, err := svc.rpcClient.RawRequest("getblock", []json.RawMessage{
		hashJSON, verboseJSON, auxpowJSON,
	})
	if err != nil {
		log.Panicf("failed to get block, error: %v, nodename: %s", err, svc.GetNodeName())
	}

	return block
}

func (svc *Service) SubmitBlocks(blocks ...[]byte) {
	for _, block := range blocks {
		resp, err := svc.rpcClient.RawRequest("submitblock", []json.RawMessage{
			block,
		})
		if err == nil {
			log.Printf("successfully submit block to nodename:%s, resp: %s", svc.GetNodeName(), resp)
			continue
		}
		if err.Error() == "duplicate" {
			log.Printf("submit block to nodename:%s duplicated, skip", svc.GetNodeName())
			continue
		}
		log.Panicf("failed to submit block to nodename:%s, error: %v, blockhex: %s", svc.GetNodeName(), err, string(block))
	}
}

func (svc *Service) GetNodeID() int {
	return svc.config.ID
}

func (svc *Service) GetNodeName() string {
	return svc.config.Name
}

func (svc *Service) Shutdown() {
	svc.rpcClient.Shutdown()
}
