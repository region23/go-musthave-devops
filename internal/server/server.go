package server

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/region23/go-musthave-devops/internal/serializers"
	"github.com/region23/go-musthave-devops/internal/server/storage"
)

type Server struct {
	repository storage.Repository
	Router     *chi.Mux
}

func New(repository storage.Repository) *Server {
	return &Server{
		repository: repository,
		Router:     chi.NewRouter(),
	}
}

func (s *Server) MountHandlers() {
	// Mount all Middleware here
	s.Router.Use(middleware.Logger)
	s.Router.Use(middleware.StripSlashes)

	// Mount all handlers here
	s.Router.Get("/", s.AllMetrics)
	s.Router.Post("/value", s.GetMetric)
	s.Router.Post("/update", s.UpdateMetric)
}

// Ручка обновляющая значение метрики
func (s *Server) UpdateMetric(w http.ResponseWriter, r *http.Request) {
	var metrics serializers.Metrics

	// decode input or return error
	err := json.NewDecoder(r.Body).Decode(&metrics)
	if err != nil {
		http.Error(w, "Decode error! please check your JSON formating.", http.StatusBadRequest)
		return
	}

	if metrics.MType != "gauge" && metrics.MType != "counter" {
		http.Error(w, "Не поддерживаемый тип метрики", http.StatusNotImplemented)
		return
	}

	// write metric to repository
	err = s.repository.Put(metrics)
	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка при сохранении метрики: %v", err.Error()), http.StatusBadRequest)
		return
	}

	// response answer
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`Metric updated`))
}

// Ручка возвращающая значение метрики
func (s *Server) GetMetric(w http.ResponseWriter, r *http.Request) {
	var metrics serializers.Metrics

	// decode input or return error
	err := json.NewDecoder(r.Body).Decode(&metrics)
	if err != nil {
		http.Error(w, "Decode error! please check your JSON formating.", http.StatusBadRequest)
		return
	}

	metric, err := s.repository.Get(metrics.ID)

	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка при получении метрики: %v", err.Error()), http.StatusNotFound)
		return
	}

	metricMarshaled, err := json.Marshal(metric)

	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка при маршалинге: %v", err.Error()), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(metricMarshaled)
}

// Ручка возвращающая все имеющиеся метрики и их значения в виде HTML-страницы
func (s *Server) AllMetrics(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/index.go.html")
	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка при парсинге html-шаблона: %v", err.Error()), http.StatusInternalServerError)
		return
	}

	tmpl.Execute(w, s.repository.All())
}
