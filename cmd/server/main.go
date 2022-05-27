package main

import (
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
	Address       string        `env:"ADDRESS" envDefault:"127.0.0.1:8080"`
	StoreInterval time.Duration `env:"STORE_INTERVAL" envDefault:"300s"`
	StoreFile     string        `env:"STORE_FILE" envDefault:"/tmp/devops-metrics-db.json"`
	Restore       bool          `env:"RESTORE" envDefault:"true"`
}

var cfg Config

func main() {
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
				metrics := repository.GetAll()
				producer.WriteMetrics(metrics)
			case <-osSigChan:
				metrics := repository.GetAll()
				producer.WriteMetrics(metrics)
				os.Exit(0)
			}
		}
	}()

	srv := server.New(repository)
	srv.MountHandlers()

	http.ListenAndServe(cfg.Address, srv.Router)
}
