package server

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"

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
	s.Router.Use(middleware.StripSlashes)

	// Mount all handlers here
	s.Router.Get("/", s.AllMetrics)
	s.Router.Get("/value/{metricType}/{metricName}", s.GetMetric)
	s.Router.Post("/update/{metricType}/{metricName}/{metricValue}", s.UpdateMetric)
}

// Ручка обновляющая значение метрики
func (s *Server) UpdateMetric(w http.ResponseWriter, r *http.Request) {
	// extract metric from url
	metricType := chi.URLParam(r, "metricType")
	metricName := chi.URLParam(r, "metricName")
	metricValue := chi.URLParam(r, "metricValue")

	if metricType != "gauge" && metricType != "counter" {
		http.Error(w, "Не поддерживаемый тип метрики", http.StatusNotImplemented)
		return
	}

	if metricType == "gauge" {
		if _, err := strconv.ParseFloat(metricValue, 64); err != nil {
			http.Error(w, fmt.Sprintf("Неверное значение метрики: %v", err.Error()), http.StatusBadRequest)
			return
		}
	}

	if metricType == "counter" {
		if _, err := strconv.ParseInt(metricValue, 10, 64); err != nil {
			http.Error(w, fmt.Sprintf("Неверное значение метрики: %v", err.Error()), http.StatusBadRequest)
			return
		}
	}

	// write metric to repository
	err := s.repository.Put(metricName, metricType, metricValue)
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
	w.Write([]byte(metric.Value))
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
