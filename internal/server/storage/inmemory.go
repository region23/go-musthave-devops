package storage

import (
	"github.com/region23/go-musthave-devops/internal/serializers"
)

type InMemory struct {
	m map[string]serializers.Metrics
}

func NewInMemory() Repository {
	return &InMemory{
		m: make(map[string]serializers.Metrics),
	}
}

func (s *InMemory) Get(key string) (serializers.Metrics, error) {
	if v, ok := s.m[key]; ok {
		return v, nil
	}
	return serializers.Metrics{}, ErrNotFound
}

func (s *InMemory) Put(metric serializers.Metrics) error {
	if curMetric, ok := s.m[metric.ID]; ok {
		if metric.MType == "counter" {
			*curMetric.Delta = *curMetric.Delta + *metric.Delta
			s.m[metric.ID] = curMetric
			return nil
		}
	}

	s.m[metric.ID] = metric
	return nil
}

// All values in map
func (s *InMemory) All() map[string]serializers.Metrics {
	return s.m
}
