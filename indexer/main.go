package main

import (
	"log"
	"time"

	"github.com/NaturalSelectionLabs/RSS3-PreGod/indexer/pkg/crawler"
	"github.com/NaturalSelectionLabs/RSS3-PreGod/indexer/pkg/db"
	"github.com/NaturalSelectionLabs/RSS3-PreGod/indexer/pkg/processor"
	"github.com/NaturalSelectionLabs/RSS3-PreGod/indexer/pkg/router"
	"github.com/NaturalSelectionLabs/RSS3-PreGod/shared/pkg/cache"
	"github.com/NaturalSelectionLabs/RSS3-PreGod/shared/pkg/config"
	"github.com/NaturalSelectionLabs/RSS3-PreGod/shared/pkg/logger"
	"github.com/NaturalSelectionLabs/RSS3-PreGod/shared/pkg/rss3uri"
	"github.com/NaturalSelectionLabs/RSS3-PreGod/shared/pkg/web"
)

func init() {
	if err := config.Setup(); err != nil {
		log.Fatalf("config.Setup err: %v", err)
	}

	if err := logger.Setup(); err != nil {
		log.Fatalf("config.Setup err: %v", err)
	}

	if err := cache.Setup(); err != nil {
		log.Fatalf("cache.Setup err: %v", err)
	}

	if err := db.Setup(); err != nil {
		log.Fatalf("db.Setup err: %v", err)
	}
}

func dispatchTasks(q chan *crawler.WorkParam, ti time.Duration) {
	// TODO: Get all accounts
	instances := []rss3uri.PlatformInstance{}
	for _, i := range instances {
		for _, n := range i.Platform.ID().GetNetwork() {
			time.Sleep(ti)
			q <- &crawler.WorkParam{Identity: i.Identity, PlatformID: i.Platform.ID(), NetworkID: n}
		}
	}
}

func pollTasks(q chan *crawler.WorkParam) {
	for {
		dispatchTasks(q, time.Minute)
		time.Sleep(24 * time.Hour)
	}
}

func main() {
	lowQ := crawler.NewTaskQueue()
	highQ := crawler.NewTaskQueue()

	proc := processor.NewProcessor(lowQ, highQ)
	go proc.ListenAndServe()

	// TODO: listen tasks from mq
	// TODO: gracefully exit
	go pollTasks(lowQ)

	srv := &web.Server{
		RunMode:      config.Config.Indexer.Server.RunMode,
		HttpPort:     config.Config.Indexer.Server.HttpPort,
		ReadTimeout:  config.Config.Indexer.Server.ReadTimeout,
		WriteTimeout: config.Config.Indexer.Server.WriteTimeout,
		Handler:      router.InitRouter(),
	}

	srv.Start()

	defer logger.Logger.Sync()
}
