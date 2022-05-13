package storage

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
	if _, ok := s.m[key]; ok {
		return ErrAlreadyExists
	}
	metric := Metric{metricType: metricType, metricValue: value}
	s.m[key] = metric
	return nil
}
