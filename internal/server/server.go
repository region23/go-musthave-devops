package server

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/region23/go-musthave-devops/internal/server/storage"
)

type Server struct {
	repository storage.Repository
}

func New(repository storage.Repository) *Server {
	return &Server{
		repository: repository,
	}
}

func (s *Server) UpdateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST requests are allowed!", http.StatusMethodNotAllowed)
		return
	}

	// extract metric from url
	url := r.URL.Path
	splittedUrl := strings.Split(url, "/")
	if len(splittedUrl) != 5 {
		http.Error(w, "Bad request. Bad count of params in URL", http.StatusBadRequest)
		return
	}
	if splittedUrl[1] == "update" {
		//процессим дальше
		if splittedUrl[2] == "gauge" || splittedUrl[2] == "counter" {
			err := s.repository.Put(splittedUrl[3], splittedUrl[2], splittedUrl[4])
			if err != nil {
				http.Error(w, fmt.Sprintf("Ошибка при сохранении метрики: %v", err.Error()), http.StatusBadRequest)
				return
			}
		}
	}

	// write metric to repository

	// response answer
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	// намеренно сделана ошибка в JSON
	w.Write([]byte(`Metric updated`))
}
