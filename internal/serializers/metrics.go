package serializers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"sync"

	"github.com/rs/zerolog/log"
)

type Gauge float64
type Counter int64

type Metric struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
	Hash  string   `json:"hash,omitempty"`  // значение хеш-функции
}

type Metrics struct {
	collection map[string]Metric
	key        string
	mu         sync.Mutex
}

func InitMetrics(key string) *Metrics {
	return &Metrics{
		collection: make(map[string]Metric),
		key:        key,
	}
}

func (m *Metrics) Get(id string) (metric Metric, exist bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	v, exist := m.collection[id]
	if exist {
		return v, true
	}

	return Metric{ID: id}, false
}

func (m *Metrics) GetAll() []Metric {
	m.mu.Lock()
	defer m.mu.Unlock()

	values := []Metric{}

	for _, value := range m.collection {
		values = append(values, value)
	}

	return values
}

func NewMetric(id string, mtype string, val ...interface{}) (Metric, error) {
	metric := Metric{ID: id, MType: mtype}
	if len(val) == 0 {
		return metric, errors.New("value for metric is absent")
	}

	switch v := val[0].(type) {
	case int64:
		if mtype == "counter" {
			metric.Delta = &v
		} else if mtype == "gauge" {
			f := float64(v)
			metric.Value = &f
		}
	case uint64:
		if mtype == "counter" {
			i := int64(v)
			metric.Delta = &i
		} else if mtype == "gauge" {
			f := float64(v)
			metric.Value = &f
		}
	case uint32:
		if mtype == "counter" {
			i := int64(v)
			metric.Delta = &i
		} else if mtype == "gauge" {
			f := float64(v)
			metric.Value = &f
		}
	case int:
		if mtype == "counter" {
			i := int64(v)
			metric.Delta = &i
		} else if mtype == "gauge" {
			f := float64(v)
			metric.Value = &f
		}
	case float64:
		if mtype == "counter" {
			i := int64(v)
			metric.Delta = &i
		} else if mtype == "gauge" {
			metric.Value = &v
		}
	case string:
		if v != "none" {
			if mtype == "counter" {
				convertedV, err := strconv.ParseInt(v, 10, 64)
				if err != nil {
					log.Error().Err(err).Msg("ошибка при парсинге значения счетчика метрики")
					return metric, errors.New("ошибка при парсинге значения счетчика метрики")
				}
				metric.Delta = &convertedV
			} else if mtype == "gauge" {
				convertedV, err := strconv.ParseFloat(v, 64)
				if err != nil {
					log.Error().Err(err).Msg("ошибка при парсинге значения метрики")
					return metric, errors.New("ошибка при парсинге значения метрики")
				}
				metric.Value = &convertedV
			}
		}
	default:
		log.Error().Msg(fmt.Sprintf("не поддерживаемый тип метрики %v", v))
		return metric, errors.New("не поддерживаемый тип метрики")
	}

	return metric, nil
}

// Добавление метрики в коллекцию
func (m *Metrics) Add(id string, mtype string, val interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	metric, err := NewMetric(id, mtype, val)
	if err != nil {
		return err
	}

	if m.key != "" {
		if metric.Value != nil {
			metric.Hash = Hash(m.key, metric.ID, metric.MType, fmt.Sprintf("%f", *metric.Value))
		}
		if metric.Delta != nil {
			metric.Hash = Hash(m.key, metric.ID, metric.MType, fmt.Sprintf("%d", *metric.Delta))
		}
	}

	m.collection[id] = metric

	return nil
}

func Hash(key, id, mType, val string) string {
	str := fmt.Sprintf("%s:%s:%s", id, mType, val)
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}
