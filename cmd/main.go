package main

import (
	"blocksyncer/config"
	"blocksyncer/service/syncer"
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var configPath = flag.String("config", "", "config file path")

type dep struct {
	config config.Config
}

func main() {
	flag.Parse()

	var dep dep
	dep.initInfra()

	ctx, cancel := context.WithCancel(context.Background())

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGKILL)
	go func() {
		s := <-sig
		log.Printf("received signal: %v, start to exit\n", s)
		go func() {
			time.Sleep(time.Second * 10)
			panic("shutdown timeout... force exit")
		}()
		cancel()
	}()

	syncerSvc := syncer.New(dep.config)
	syncerSvc.Sync(ctx)

	<-ctx.Done()

	log.Println("server exited without error :)")
}

func (d *dep) initInfra() {
	d.config = config.New(*configPath)

	log.Printf("succeed to load config, nodes count: %d, nodes: \n%s\n", len(d.config.Nodes), d.config.SprintNodeNames())
}
