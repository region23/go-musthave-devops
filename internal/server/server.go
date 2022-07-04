package server

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/region23/go-musthave-devops/internal/serializers"
	mw "github.com/region23/go-musthave-devops/internal/server/middleware"
	"github.com/region23/go-musthave-devops/internal/server/storage"
	"github.com/region23/go-musthave-devops/internal/server/storage/database"
)

type Server struct {
	storage storage.Repository
	Router  *chi.Mux
	Key     string
	DBPool  *pgxpool.Pool
}

func New(storage storage.Repository, key string, dbpool *pgxpool.Pool) *Server {
	return &Server{
		storage: storage,
		Router:  chi.NewRouter(),
		Key:     key,
		DBPool:  dbpool,
	}
}

func (s *Server) MountHandlers() {
	// Mount all Middleware here
	s.Router.Use(middleware.Logger)
	s.Router.Use(middleware.StripSlashes)
	//s.Router.Use(middleware.Compress(5))
	s.Router.Use(mw.GZipHandle)
	// Mount all handlers here
	s.Router.Get("/", s.AllMetrics)
	s.Router.Post("/updates", s.UpdateBatchMetricsJSON)
	s.Router.Post("/update", s.UpdateMetricJSON)
	s.Router.Post("/update/{metricType}/{metricName}/{metricValue}", s.UpdateMetric)
	s.Router.Post("/value", s.GetMetricJSON)
	s.Router.Get("/value/{metricType}/{metricName}", s.GetMetric)
	s.Router.Get("/ping", s.Ping)

}

func (s *Server) UpdateMetric(w http.ResponseWriter, r *http.Request) {
	// extract metric from url
	metricType := chi.URLParam(r, "metricType")
	metricName := chi.URLParam(r, "metricName")
	metricValue := chi.URLParam(r, "metricValue")

	// if metricValue == "none" {
	// 	http.Error(w, "Неверное значение метрики", http.StatusBadRequest)
	// 	return
	// }

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

	metric := serializers.NewMetrics(metricName, metricType, metricValue)

	// write metric to repository
	err := s.storage.Put(&metric)
	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка при сохранении метрики: %v", err.Error()), http.StatusBadRequest)
		return
	}

	// response answer
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`Metric updated`))
}

// Ручка обновляющая пачку метрик
func (s *Server) UpdateBatchMetricsJSON(w http.ResponseWriter, r *http.Request) {
	var metrics []serializers.Metrics

	// decode input or return error
	err := json.NewDecoder(r.Body).Decode(&metrics)
	if err != nil {
		http.Error(w, "Decode error! please check your JSON formating.", http.StatusBadRequest)
		return
	}

	if len(metrics) == 0 {
		http.Error(w, "Metric name can't be empty", http.StatusBadRequest)
		return
	}

	for _, metric := range metrics {
		if metric.Value == nil && metric.Delta == nil {
			http.Error(w, "Value can't be nil", http.StatusBadRequest)
			return
		}

		if metric.MType != "gauge" && metric.MType != "counter" {
			http.Error(w, "Не поддерживаемый тип метрики", http.StatusNotImplemented)
			return
		}

		// Если хэш не пустой, то сверяем хэши
		checkHash(s.Key, &metric, w)

		// write metric to repository
		err = s.storage.Put(&metric)
		if err != nil {
			http.Error(w, fmt.Sprintf("Ошибка при сохранении метрики: %v", err.Error()), http.StatusBadRequest)
			return
		}

	}

	// response answer
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`Metrics updated`))
}

// Ручка обновляющая значение метрики
func (s *Server) UpdateMetricJSON(w http.ResponseWriter, r *http.Request) {
	var metric *serializers.Metrics

	// decode input or return error
	err := json.NewDecoder(r.Body).Decode(&metric)
	if err != nil {
		http.Error(w, "Decode error! please check your JSON formating.", http.StatusBadRequest)
		return
	}

	if metric.ID == "" {
		http.Error(w, "Metric name can't be empty", http.StatusNotFound)
		return
	}

	if metric.Value == nil && metric.Delta == nil {
		http.Error(w, "Value can't be nil", http.StatusBadRequest)
		return
	}

	if metric.MType != "gauge" && metric.MType != "counter" {
		http.Error(w, "Не поддерживаемый тип метрики", http.StatusNotImplemented)
		return
	}

	// Если хэш не пустой, то сверяем хэши
	checkHash(s.Key, metric, w)

	// write metric to repository
	err = s.storage.Put(metric)
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
	metricName := chi.URLParam(r, "metricName")

	metric, err := s.storage.Get(metricName)

	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка при получении метрики: %v", err.Error()), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if metric.MType == "counter" {
		w.Write([]byte(strconv.FormatInt(*metric.Delta, 10)))
	} else if metric.MType == "gauge" {
		w.Write([]byte(fmt.Sprintf("%g", *metric.Value)))
	}
}

// Ручка возвращающая значение метрики
func (s *Server) GetMetricJSON(w http.ResponseWriter, r *http.Request) {
	metrics := &serializers.Metrics{}
	// decode input or return error
	err := json.NewDecoder(r.Body).Decode(metrics)
	if err != nil {
		http.Error(w, "Decode error! please check your JSON formating.", http.StatusBadRequest)
		return
	}

	metrics, err = s.storage.Get(metrics.ID)

	if metrics == nil {
		http.Error(w, "Metric not found", http.StatusNotFound)
		return
	}

	// Если хэш не пустой, то сверяем хэши
	metrics.Hash = checkHash(s.Key, metrics, w)

	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка при получении метрики: %v", err.Error()), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(metrics)
	if err != nil {
		http.Error(w, "Encode error! please check your JSON formating.", http.StatusBadRequest)
		return
	}
}

// Ручка возвращающая все имеющиеся метрики и их значения в виде HTML-страницы
func (s *Server) AllMetrics(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/index.go.html")
	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка при парсинге html-шаблона: %v", err.Error()), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	metrics, err := s.storage.All()
	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка при получении метрик: %v", err.Error()), http.StatusInternalServerError)
	}

	tmpl.Execute(w, metrics)
}

// Проверяем соединение с базой данных
func (s *Server) Ping(w http.ResponseWriter, r *http.Request) {
	err := database.Ping(s.DBPool)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Ping OK"))
}

// Сверяем хэши, а если пустой, то генерим новый
func checkHash(key string, metric *serializers.Metrics, w http.ResponseWriter) (hash string) {
	if key != "" {
		var serverGeneratedHash string

		if metric.Value != nil {
			serverGeneratedHash = serializers.Hash(metric.MType, metric.ID, fmt.Sprintf("%f", *metric.Value), key)
		}
		if metric.Delta != nil {
			serverGeneratedHash = serializers.Hash(metric.MType, metric.ID, fmt.Sprintf("%d", *metric.Delta), key)
		}

		if metric.Hash != "" && metric.Hash != "none" && metric.Hash != serverGeneratedHash {
			http.Error(w, "Hash is not valid", http.StatusBadRequest)
			return
		}

		return serverGeneratedHash
	}

	return ""
}
