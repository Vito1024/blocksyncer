package fractalbitcoin

import (
	"blocksyncer/config"
	"fmt"
	"log"
	"time"

	"github.com/go-resty/resty/v2"
)

const maxRetry = 10
const timeout = 10 * time.Second

type Service struct {
	config config.Node
	client *resty.Client
}

func New(config config.Node) *Service {
	svc := &Service{
		config: config,
		client: resty.New().
			SetBasicAuth(config.RPCUser, config.RPCPassword).
			SetHeader("Content-Type", "application/json").
			SetBaseURL(fmt.Sprintf("%s:%d", config.RPCHost, config.RPCPort)).
			SetTimeout(timeout).
			SetRetryCount(maxRetry).
			SetRetryWaitTime(200 * time.Millisecond).
			SetRetryMaxWaitTime(2 * time.Second),
	}

	log.Printf("succeed to create rpc client for nodeid: %d, nodename: %s\n", svc.config.ID, svc.config.Name)
	return svc
}

func (svc *Service) GetBlockHeadersAndBlocks() (int32, int32) {
	var headersBlocks struct {
		Headers int32 `json:"headers"`
		Blocks  int32 `json:"blocks"`
	}
	resp, err := svc.client.R().
		SetBody(map[string]string{
			"jsonrpc": "1.0",
			"method":  "getblockchaininfo",
		}).
		SetResult(&headersBlocks).
		Post("")
	if err != nil {
		log.Panicf("failed to getblockchaininfo, error: %v", err)
	}
	if resp.IsError() {
		log.Panicf("failed to getblockchaininfo, error: %v", resp.Error())
	}

	return headersBlocks.Headers, headersBlocks.Blocks
}

func (svc *Service) GetNodeID() int {
	return svc.config.ID
}
