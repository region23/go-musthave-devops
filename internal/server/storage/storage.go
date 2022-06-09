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
	Get(key string) (serializers.Metrics, error)
	Put(metric *serializers.Metrics) error
	All() (map[string]serializers.Metrics, error)
	UpdateAll(m map[string]serializers.Metrics) error
}
