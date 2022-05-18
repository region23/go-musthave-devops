package storage

import "errors"

var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
)

type Metric struct {
	Type  string
	Value string
}

type Repository interface {
	Get(key string) (Metric, error)
	Put(key, metricType, value string) error
	All() map[string]Metric
}
