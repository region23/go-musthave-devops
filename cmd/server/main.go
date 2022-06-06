package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/region23/go-musthave-devops/internal/server"
	"github.com/region23/go-musthave-devops/internal/server/storage"
)

type Config struct {
	Address       string        `env:"ADDRESS"`
	StoreInterval time.Duration `env:"STORE_INTERVAL"`
	StoreFile     string        `env:"STORE_FILE"`
	Restore       bool          `env:"RESTORE"`
	Key           string        `env:"KEY"`
}

var cfg Config = Config{}

func init() {
	flag.StringVar(&cfg.Address, "a", "127.0.0.1:8080", "server address")
	flag.BoolVar(&cfg.Restore, "r", true, "restore metrics before start")
	flag.DurationVar(&cfg.StoreInterval, "i", 300*time.Second, "store interval")
	flag.StringVar(&cfg.StoreFile, "f", "/tmp/devops-metrics-db.json", "path to file for metrics store")
	flag.StringVar(&cfg.Key, "k", "", "key for hashing")
}

func main() {
	flag.Parse()

	if err := env.Parse(&cfg); err != nil {
		fmt.Printf("%+v\n", err)
	}

	osSigChan := make(chan os.Signal, 1)
	signal.Notify(osSigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	repository := storage.NewInMemory()

	consumer, err := storage.NewConsumer(cfg.StoreFile)
	if err != nil {
		log.Fatalf("%+v\n", err)
	}

	if cfg.Restore {
		metricsFromFile, err := consumer.ReadMetrics()
		if err != nil {
			log.Fatalf("%+v\n", err)
		}

		repository.UpdateAll(*metricsFromFile)
	}

	producer, err := storage.NewProducer(cfg.StoreFile)
	if err != nil {
		log.Fatalf("%+v\n", err)
	}

	storeIntervalTick := time.NewTicker(cfg.StoreInterval)
	go func() {
		for {
			select {
			case <-storeIntervalTick.C:
				metrics := repository.All()
				producer.WriteMetrics(metrics)
			case <-osSigChan:
				metrics := repository.All()
				producer.WriteMetrics(metrics)
				os.Exit(0)
			}
		}
	}()

	srv := server.New(repository, cfg.Key)
	srv.MountHandlers()

	http.ListenAndServe(cfg.Address, srv.Router)
}
