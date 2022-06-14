package storage

import (
	"sync"

	"github.com/region23/go-musthave-devops/internal/serializers"
)

type InMemory struct {
	mu sync.Mutex
	m  map[string]serializers.Metrics
}

func NewInMemory() Repository {
	return &InMemory{
		m: make(map[string]serializers.Metrics),
	}
}

func (s *InMemory) Get(key string) (*serializers.Metrics, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if v, ok := s.m[key]; ok {
		return &v, nil
	}
	return nil, ErrNotFound
}

func (s *InMemory) Put(metric *serializers.Metrics) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if curMetric, ok := s.m[metric.ID]; ok {
		if metric.MType == "counter" {
			*curMetric.Delta = *curMetric.Delta + *metric.Delta
			s.m[metric.ID] = curMetric
			return nil
		}
	}

	s.m[metric.ID] = *metric
	return nil
}

// All values in map
func (s *InMemory) All() (map[string]serializers.Metrics, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.m, nil
}

// Обновляет мапу с метриками в памяти снэпшотом данных из файла
func (s *InMemory) UpdateAll(m map[string]serializers.Metrics) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.m = m
	return nil
}
