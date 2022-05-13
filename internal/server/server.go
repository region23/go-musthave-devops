package server

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
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

	// Mount all handlers here
	s.Router.Get("/", s.AllMetricsHandler)
	s.Router.Get("/value/{metricType}/{metricName}", s.GetMetricHandler)
	s.Router.Post("/update/{metricType}/{metricName}/{metricValue}", s.UpdateMetricHandler)
}

// Ручка обновляющая значение метрики
func (s *Server) UpdateMetricHandler(w http.ResponseWriter, r *http.Request) {
	// extract metric from url
	metricType := chi.URLParam(r, "metricType")
	metricName := chi.URLParam(r, "metricName")
	metricValue := chi.URLParam(r, "metricValue")

	// write metric to repository
	if metricType == "gauge" || metricType == "counter" {
		err := s.repository.Put(metricName, metricType, metricValue)
		if err != nil {
			http.Error(w, fmt.Sprintf("Ошибка при сохранении метрики: %v", err.Error()), http.StatusBadRequest)
			return
		}
	}

	// response answer
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`Metric updated`))
}

// Ручка возвращающая значение метрики
func (s *Server) GetMetricHandler(w http.ResponseWriter, r *http.Request) {
	// extract metric from url
	//metricType := chi.URLParam(r, "metricType")
	metricName := chi.URLParam(r, "metricName")

	metric, err := s.repository.Get(metricName)

	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка при получении метрики: %v", err.Error()), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(metric.MetricValue))
}

// Ручка возвращающая все имеющиеся метрики и их значения в виде HTML-страницы
func (s *Server) AllMetricsHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/index.go.html")
	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка при парсинге html-шаблона: %v", err.Error()), http.StatusInternalServerError)
		return
	}

	tmpl.Execute(w, s.repository.All())
}
