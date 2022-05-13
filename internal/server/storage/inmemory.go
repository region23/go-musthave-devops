package storage

import "strconv"

type Metric struct {
	metricType  string
	metricValue string
}

type InMemory struct {
	m map[string]Metric
}

func NewInMemory() Repository {
	return &InMemory{
		m: make(map[string]Metric),
	}
}

func (s *InMemory) Get(key string) (Metric, error) {
	if v, ok := s.m[key]; ok {
		return v, nil
	}
	return Metric{}, ErrNotFound
}

func (s *InMemory) Put(key string, metricType string, value string) error {
	if curMetric, ok := s.m[key]; ok {
		//
		if key == "PollCount" {
			if newValue, err := strconv.Atoi(value); err == nil {
				if curValue, err := strconv.Atoi(curMetric.metricValue); err == nil {
					value = strconv.Itoa(curValue + newValue)
				}
			}
		} else {
			return ErrAlreadyExists
		}
	}
	metric := Metric{metricType: metricType, metricValue: value}
	s.m[key] = metric
	return nil
}
