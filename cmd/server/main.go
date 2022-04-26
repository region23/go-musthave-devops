package main

import (
	"log"
	"net/http"

	"github.com/region23/go-musthave-devops/cmd/server/handlers"
)

func main() {
	http.HandleFunc("/", handlers.UpdateHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
