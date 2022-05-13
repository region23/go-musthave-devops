package storage

import "errors"

var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
)

type Metric struct {
	MetricType  string
	MetricValue string
}

type Repository interface {
	Get(key string) (Metric, error)
	Put(key, metricType, value string) error
	All() map[string]Metric
}
