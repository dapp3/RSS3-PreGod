package main

import (
	"log"
	"time"

	"github.com/NaturalSelectionLabs/RSS3-PreGod/indexer/pkg/api/arweave"
	"github.com/NaturalSelectionLabs/RSS3-PreGod/indexer/pkg/api/gitcoin"
	"github.com/NaturalSelectionLabs/RSS3-PreGod/indexer/pkg/crawler"
	"github.com/NaturalSelectionLabs/RSS3-PreGod/indexer/pkg/db"
	"github.com/NaturalSelectionLabs/RSS3-PreGod/indexer/pkg/processor"
	"github.com/NaturalSelectionLabs/RSS3-PreGod/indexer/pkg/router"
	"github.com/NaturalSelectionLabs/RSS3-PreGod/shared/pkg/cache"
	"github.com/NaturalSelectionLabs/RSS3-PreGod/shared/pkg/config"
	"github.com/NaturalSelectionLabs/RSS3-PreGod/shared/pkg/logger"
	"github.com/NaturalSelectionLabs/RSS3-PreGod/shared/pkg/rss3uri"
	"github.com/NaturalSelectionLabs/RSS3-PreGod/shared/pkg/web"
	"github.com/RichardKnop/machinery/v1/tasks"
	jsoniter "github.com/json-iterator/go"
)

var jsoni = jsoniter.ConfigCompatibleWithStandardLibrary

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

	if err := processor.Setup(); err != nil {
		log.Fatalf("processor.Setup err: %v", err)
	}
}

func dispatchTasks(pauseDuration time.Duration) {
	// TODO: Get all accounts
	instances := []rss3uri.PlatformInstance{}
	for _, i := range instances {
		for _, n := range i.Platform.ID().GetNetwork() {
			time.Sleep(pauseDuration)

			param := crawler.WorkParam{Identity: i.Identity, PlatformID: i.Platform.ID(), NetworkID: n}

			param.LastIndexedTsp, _ = processor.GetLastIndexedTsp(&i)

			// marshal WorkParam to string so it's supported by machinery
			payload, err := jsoni.MarshalToString(param)

			if err != nil {
				logger.Errorf("dispatchTasks WorkParam mashalling error: %v", err)

				return
			}

			crawlerTask := tasks.Signature{
				// the name is defined by RegisterTasks() in processor/processor.go
				Name: "dispatch",
				Args: []tasks.Arg{
					{
						Type:  "string",
						Value: payload,
					},
				},
			}

			_, err = processor.SendTask(crawlerTask)

			if err != nil {
				processor.UpdateLastIndexedTsp(&i)
			}
		}
	}
}

func main() {
	srv := &web.Server{
		RunMode:      config.Config.Indexer.Server.RunMode,
		HttpPort:     config.Config.Indexer.Server.HttpPort,
		ReadTimeout:  config.Config.Indexer.Server.ReadTimeout,
		WriteTimeout: config.Config.Indexer.Server.WriteTimeout,
		Handler:      router.InitRouter(),
	}

	srv.Start()

	// arweave crawler
	ar := arweave.NewCrawler(arweave.MirrorUploader, arweave.DefaultCrawlConfig)
	ar.Start()

	// gitcoin crawler
	gc := gitcoin.NewCrawler(*gitcoin.DefaultEthConfig, *gitcoin.DefaultPolygonConfig, *gitcoin.DefaultZksyncConfig)
	go gc.PolygonStart()
	go gc.EthStart()
	go gc.ZkStart()

	defer logger.Logger.Sync()

	// TODO: adjust interval
	dispatchTasks(time.Minute)
}
