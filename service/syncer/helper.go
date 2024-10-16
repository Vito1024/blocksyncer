package syncer

import (
	"log"
	"sync"
)

type nodeHeadersBlocks struct {
	nodeID  int
	headers int32
	blocks  int32
}

type nodesInfo struct {
	longestNodeID      int
	nodesHeadersBlocks map[int]nodeHeadersBlocks
}

func (svc *Service) sync() {
	log.Println("start to sync")
	allNodesInfo := svc.getAllNodesInfo()

	var wg sync.WaitGroup
	for _, node := range allNodesInfo.nodesHeadersBlocks {
		if node.nodeID == allNodesInfo.longestNodeID {
			continue
		}
		longestNode, ok := allNodesInfo.nodesHeadersBlocks[allNodesInfo.longestNodeID]
		if !ok {
			panic("failed to find longest node after getting all nodes info")
		}
		if node.blocks == longestNode.blocks {
			continue
		}
		if node.headers < longestNode.headers {
			// how to sync?

			continue
		}
		if node.blocks < longestNode.blocks {
			// sync

			continue
		}
		log.Printf("unexpect node %d status: headers %d, blocks %d", node.nodeID, node.headers, node.blocks)
	}
	wg.Wait()
	log.Println("finished sync")
}

func (svc *Service) getAllNodesInfo() nodesInfo {
	var nodesInfo nodesInfo
	nodesInfo.nodesHeadersBlocks = make(map[int]nodeHeadersBlocks, len(svc.fbNodes))

	var mu sync.Mutex
	var wg sync.WaitGroup
	for _, node := range svc.fbNodes {
		node := node

		wg.Add(1)
		go func() {
			defer wg.Done()

			headers, blocks := node.GetBlockHeadersAndBlocks()
			mu.Lock()
			nodesInfo.nodesHeadersBlocks[node.GetNodeID()] = nodeHeadersBlocks{
				nodeID:  node.GetNodeID(),
				headers: headers,
				blocks:  blocks,
			}
			if longestNode, ok := nodesInfo.nodesHeadersBlocks[nodesInfo.longestNodeID]; ok && blocks > longestNode.blocks {
				nodesInfo.longestNodeID = node.GetNodeID()
			}
			mu.Unlock()
		}()
	}
	wg.Wait()

	log.Printf("succeed to get %d nodes info", len(nodesInfo.nodesHeadersBlocks))
	return nodesInfo
}
