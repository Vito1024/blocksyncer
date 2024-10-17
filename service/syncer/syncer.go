package syncer

import (
	bs "blocksyncer"
	"blocksyncer/config"
	"blocksyncer/fractalbitcoin"
	"context"
	"log"
	"sync"
	"time"
)

type node struct {
	*fractalbitcoin.Service
	headers int64
	blocks  int64
}

type rootBlock struct {
	*bs.Block
	height int64
}

type Service struct {
	config config.Config
	nodes  []*node

	rootBlock   *rootBlock
	longestNode *node
}

func New(config config.Config) *Service {
	if len(config.Nodes) <= 1 {
		log.Panicf("fb nodes count must be greater than 1, exiting")
	}

	svc := &Service{
		config: config,
	}
	svc.initAllNodes()

	return svc
}

func (svc *Service) Sync(ctx context.Context) {
	for range time.Tick(bs.IntervalOrDefault(svc.config.SyncInterval, 30*time.Second)) {
		select {
		case <-ctx.Done():
			log.Println("received ctx.Done, exit sync loop")
			return
		default:
			svc.sync()
		}
	}
}

func (svc *Service) shortestNode() *node {
	shortestNode := svc.nodes[0]
	for _, node := range svc.nodes {
		if node.blocks < shortestNode.blocks {
			shortestNode = node
		}
	}

	return shortestNode
}

// backtrack finds the root block by comparing all nodes blocks
func (svc *Service) backtrack() {
	log.Println("start backtrack")
	shortestNode := svc.shortestNode()

	for height := int64(shortestNode.blocks); height >= 0; height-- {
		shortestNodeHeightHash := shortestNode.GetBlockHashByHeight(height)

		isRootBlock := true
		var wg sync.WaitGroup
		var mu sync.Mutex
		for _, node := range svc.nodes {
			node := node
			if node == shortestNode {
				continue
			}

			wg.Add(1)
			go func() {
				defer wg.Done()
				hash := node.GetBlockHashByHeight(height)
				if !shortestNodeHeightHash.IsEqual(hash) {
					mu.Lock()
					isRootBlock = false
					mu.Unlock()
				}
			}()
		}
		wg.Wait()

		if isRootBlock {
			svc.rootBlock = &rootBlock{
				Block:  shortestNode.GetBlockByHeight(height),
				height: height,
			}

			log.Printf("backtrace found root block at height %d, hash: %s\n", height, svc.rootBlock.Header.BlockHash().String())
			break
		}
	}

	if svc.rootBlock == nil {
		log.Panicf("backtrace failed to find root block, exiting")
	}
}

func (svc *Service) initAllNodes() {
	var wg sync.WaitGroup
	var mu sync.Mutex
	for _, nodeConfig := range svc.config.Nodes {
		nodeConfig := nodeConfig

		wg.Add(1)
		go func() {
			defer wg.Done()
			node := node{
				Service: fractalbitcoin.New(nodeConfig),
			}
			mu.Lock()
			svc.nodes = append(svc.nodes, &node)
			mu.Unlock()
		}()
	}
	wg.Wait()

	log.Printf("successfully bundled %d fb nodes\n", len(svc.nodes))
}

// flushNodes flushes all nodes blocks and headers, and set the longest node
func (svc *Service) flushNodes() {
	log.Println("flushing nodes")
	var wg sync.WaitGroup
	var mu sync.Mutex
	for _, node := range svc.nodes {
		node := node
		wg.Add(1)
		go func() {
			defer wg.Done()

			node.headers, node.blocks = node.GetBlockHeadersAndBlocks()
			{
				mu.Lock()
				if svc.longestNode == nil || node.blocks > svc.longestNode.blocks {
					svc.longestNode = node
				}
				mu.Unlock()
			}
			log.Printf("node[%s] flushed: %d\n", node.GetNodeName(), node.blocks)
		}()
	}
	wg.Wait()
	log.Printf("successfully flushed all nodes. longest node[%s]: %d\n", svc.longestNode.GetNodeName(), svc.longestNode.blocks)
}

// sync flushes all nodes, finds the root block and syncs all delta blocks to all slower nodes
func (svc *Service) sync() {
	log.Println("start syncing")
	svc.flushNodes()
	if delta := svc.longestNode.blocks - svc.shortestNode().blocks; delta < 1 { // chose some value as the threshold to avoid unnecessary sync
		log.Printf("delta is %d, no need to sync\n\n", delta)
		return
	}
	svc.backtrack()

	delta := svc.longestNode.blocks - svc.rootBlock.height
	log.Printf("delta is %d, longest height: %d, rootblock height: %d\n", delta, svc.longestNode.blocks, svc.rootBlock.height)

	// get all delta blocks to prepare for sync
	deltaBlocks := make([][]byte, delta)
	{
		maxConcurrency := make(chan struct{}, 8)
		wg := sync.WaitGroup{}
		for height := svc.rootBlock.height + 1; height <= svc.longestNode.blocks; height++ {
			height := height
			maxConcurrency <- struct{}{}
			wg.Add(1)
			go func() {
				defer wg.Done()
				defer func() { <-maxConcurrency }()
				deltaBlocks[height-svc.rootBlock.height-1] = svc.longestNode.GetBlockHexByHeight(height)
			}()
		}
		wg.Wait()
	}
	log.Printf("successfully get %d delta blocks\n", len(deltaBlocks))

	// sync all delta blocks to all slower nodes
	log.Println("start syncing delta blocks")
	var wg sync.WaitGroup
	for _, node := range svc.nodes {
		node := node
		if node == svc.longestNode {
			continue
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			if node.blocks < svc.longestNode.blocks {
				node.SubmitBlocks(deltaBlocks...)
			}
		}()
	}
	wg.Wait()

	log.Printf("finished syncing\n\n")
}

func (svc *Service) Shutdown() {
	var wg sync.WaitGroup
	for _, node := range svc.nodes {
		node := node
		wg.Add(1)
		go func() {
			defer wg.Done()
			node.Shutdown()
		}()
	}
	wg.Wait()
}
