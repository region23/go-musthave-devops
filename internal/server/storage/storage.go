package storage

import (
	"errors"

	"github.com/region23/go-musthave-devops/internal/serializers"
)

var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
)

type Repository interface {
	Get(key string) (*serializers.Metric, error)
	Put(metric serializers.Metric) error
	All() (map[string]serializers.Metric, error)
	UpdateAll(m map[string]serializers.Metric) error
}
