package server

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/region23/go-musthave-devops/internal/serializers"
	mw "github.com/region23/go-musthave-devops/internal/server/middleware"
	"github.com/region23/go-musthave-devops/internal/server/storage"
)

type Server struct {
	storage storage.Repository
	Router  *chi.Mux
	Key     string
}

func New(storage storage.Repository, key string) *Server {
	return &Server{
		storage: storage,
		Router:  chi.NewRouter(),
		Key:     key,
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
	s.Router.Post("/update", s.UpdateMetricJSON)
	s.Router.Post("/update/{metricType}/{metricName}/{metricValue}", s.UpdateMetric)
	s.Router.Post("/value", s.GetMetricJSON)
	s.Router.Get("/value/{metricType}/{metricName}", s.GetMetric)

}

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

	metric := serializers.NewMetrics(metricName, metricType, metricValue)

	// write metric to repository
	err := s.storage.Put(metric)
	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка при сохранении метрики: %v", err.Error()), http.StatusBadRequest)
		return
	}

	// response answer
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`Metric updated`))
}

// Ручка обновляющая значение метрики
func (s *Server) UpdateMetricJSON(w http.ResponseWriter, r *http.Request) {
	var metrics serializers.Metrics

	// decode input or return error
	err := json.NewDecoder(r.Body).Decode(&metrics)
	if err != nil {
		http.Error(w, "Decode error! please check your JSON formating.", http.StatusBadRequest)
		return
	}

	if metrics.ID == "" {
		http.Error(w, "Metric name can't be empty", http.StatusNotFound)
		return
	}

	if metrics.Value == nil && metrics.Delta == nil {
		http.Error(w, "Value can't be nil", http.StatusBadRequest)
		return
	}

	if metrics.MType != "gauge" && metrics.MType != "counter" {
		http.Error(w, "Не поддерживаемый тип метрики", http.StatusNotImplemented)
		return
	}

	// Если хэш не пустой, то сверяем хэши
	if metrics.Hash != "" && metrics.Hash != "none" && s.Key != "" {
		fmt.Println("KEY: ", s.Key)
		var serverGeneratedHash string
		if metrics.MType == "gauge" {
			serverGeneratedHash = serializers.Hash(metrics.MType, metrics.ID, fmt.Sprintf("%g", *metrics.Value), s.Key)
		} else if metrics.MType == "counter" {
			serverGeneratedHash = serializers.Hash(metrics.MType, metrics.ID, strconv.FormatInt(*metrics.Delta, 10), s.Key)
		}

		if metrics.Hash != serverGeneratedHash {
			fmt.Printf("Хэш от клиента %v не равен хэшу вычисленному на сервере %v\n", metrics.Hash, serverGeneratedHash)
			fmt.Printf("%v %v %g %v\n", metrics.MType, metrics.ID, *metrics.Value, s.Key)
			http.Error(w, "Hash is not valid", http.StatusBadRequest)
			return
		}
	}

	// write metric to repository
	err = s.storage.Put(metrics)
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
	var metrics serializers.Metrics
	// decode input or return error
	err := json.NewDecoder(r.Body).Decode(&metrics)
	if err != nil {
		http.Error(w, "Decode error! please check your JSON formating.", http.StatusBadRequest)
		return
	}

	fmt.Println(metrics)

	metrics, err = s.storage.Get(metrics.ID)

	// Если хэш не пустой, то сверяем хэши
	if s.Key != "" {
		fmt.Println("KEY: ", s.Key)

		var serverGeneratedHash string
		if metrics.MType == "gauge" {
			fmt.Println(*metrics.Value)
			serverGeneratedHash = serializers.Hash(metrics.MType, metrics.ID, fmt.Sprintf("%g", *metrics.Value), s.Key)
		} else if metrics.MType == "counter" {
			serverGeneratedHash = serializers.Hash(metrics.MType, metrics.ID, strconv.FormatInt(*metrics.Delta, 10), s.Key)
		}

		metrics.Hash = serverGeneratedHash
	}

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
	tmpl.Execute(w, s.storage.All())
}
