package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/caarlos0/env/v6"
	"github.com/region23/go-musthave-devops/internal/server"
	"github.com/region23/go-musthave-devops/internal/server/storage"
)

type Config struct {
	Address string `env:"ADDRESS" envDefault:"127.0.0.1:8080"`
}

var cfg Config

func main() {
	if err := env.Parse(&cfg); err != nil {
		fmt.Printf("%+v\n", err)
	}

	repository := storage.NewInMemory()
	srv := server.New(repository)
	srv.MountHandlers()

	log.Fatal(http.ListenAndServe(cfg.Address, srv.Router))
}
