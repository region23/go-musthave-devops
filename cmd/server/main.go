package main

import (
	"log"
	"net/http"

	"github.com/region23/go-musthave-devops/internal/server"
	"github.com/region23/go-musthave-devops/internal/server/storage"
)

func main() {
	repository := storage.NewInMemory()
	srv := server.New(repository)

	http.HandleFunc("/", srv.UpdateHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
