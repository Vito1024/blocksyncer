package syncer

import (
	bs "blocksyncer"
	"blocksyncer/config"
	"blocksyncer/fractalbitcoin"
	"context"
	"log"
	"time"
)

type Service struct {
	config  config.Config
	fbNodes []*fractalbitcoin.Service
}

func New(config config.Config) *Service {
	svc := &Service{
		config: config,
	}

	for _, node := range config.Nodes {
		svc.fbNodes = append(svc.fbNodes, fractalbitcoin.New(node))
	}
	log.Printf("successfully bundled %d fb nodes\n", len(svc.fbNodes))

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
