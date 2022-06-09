package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/region23/go-musthave-devops/internal/server"
	"github.com/region23/go-musthave-devops/internal/server/storage"
	"github.com/region23/go-musthave-devops/internal/server/storage/database"
)

var dbpool *pgxpool.Pool

type Config struct {
	Address       string        `env:"ADDRESS"`
	StoreInterval time.Duration `env:"STORE_INTERVAL"`
	StoreFile     string        `env:"STORE_FILE"`
	Restore       bool          `env:"RESTORE"`
	Key           string        `env:"KEY"`
	DatabaseDSN   string        `env:"DATABASE_DSN"`
}

var cfg Config = Config{}

func init() {
	flag.StringVar(&cfg.Address, "a", "127.0.0.1:8080", "server address")
	flag.BoolVar(&cfg.Restore, "r", true, "restore metrics before start")
	flag.DurationVar(&cfg.StoreInterval, "i", 300*time.Second, "store interval")
	flag.StringVar(&cfg.StoreFile, "f", "/tmp/devops-metrics-db.json", "path to file for metrics store")
	flag.StringVar(&cfg.Key, "k", "", "key for hashing")
	flag.StringVar(&cfg.DatabaseDSN, "d", "", "database connection string")
}

func main() {
	flag.Parse()

	if err := env.Parse(&cfg); err != nil {
		fmt.Printf("%+v\n", err)
	}

	osSigChan := make(chan os.Signal, 1)
	signal.Notify(osSigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	var repository storage.Repository

	if cfg.DatabaseDSN == "" {
		fmt.Println("DSN пуст, используется файловое хранилище")
		repository = storage.NewInMemory()
		consumer, err := storage.NewConsumer(cfg.StoreFile)
		if err != nil {
			log.Fatalf("%+v\n", err)
		}

		if cfg.Restore {
			metricsFromFile, err := consumer.ReadMetrics()
			if err != nil {
				log.Fatalf("%+v\n", err)
			}

			repository.UpdateAll(metricsFromFile)
			if err != nil {
				log.Fatalf("%+v\n", err)
			}
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
					metrics, err := repository.All()

					if err != nil {
						log.Fatalf("%+v\n", err)
					}

					producer.WriteMetrics(metrics)
				case <-osSigChan:
					metrics, err := repository.All()

					if err != nil {
						log.Fatalf("%+v\n", err)
					}

					producer.WriteMetrics(metrics)
					os.Exit(0)
				}
			}
		}()
	} else {
		// Инициализируем подключение к базе данных
		var err error
		dbpool, err = pgxpool.Connect(context.Background(), cfg.DatabaseDSN)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to connection to database: %v\n", err)
			os.Exit(1)
		}

		log.Println("Подключение к базе данных успешно произведено")

		err = database.InitDB(dbpool)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to connection to database: %v\n", err)
			os.Exit(1)
		}

		defer dbpool.Close()

		go func() {
			<-osSigChan
			os.Exit(0)
		}()

		repository = database.NewInDatabase(dbpool, cfg.Key)

	}

	log.Println("Starting server...")

	srv := server.New(repository, cfg.Key, dbpool)
	srv.MountHandlers()

	http.ListenAndServe(cfg.Address, srv.Router)
}
