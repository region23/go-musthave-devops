package storage

import "strconv"

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
		if metricType == "counter" {
			if newValue, err := strconv.ParseInt(value, 10, 64); err == nil {
				if curValue, err := strconv.ParseInt(curMetric.Value, 10, 64); err == nil {
					value = strconv.FormatInt(curValue+newValue, 10)
				} else {
					return err
				}
			} else {
				return err
			}
		}
	}

	metric := Metric{Type: metricType, Value: value}
	s.m[key] = metric
	return nil
}

// All values in map
func (s *InMemory) All() map[string]Metric {
	return s.m
}
