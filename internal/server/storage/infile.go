package storage

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"sync"

	"github.com/region23/go-musthave-devops/internal/serializers"
)

type Producer struct {
	mu      sync.Mutex
	file    *os.File
	encoder *json.Encoder
}

func NewProducer(fileName string) (*Producer, error) {
	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE, 0777)
	if err != nil {
		return nil, err
	}
	return &Producer{
		file:    file,
		encoder: json.NewEncoder(file),
	}, nil
}

func (p *Producer) WriteMetrics(metrics map[string]serializers.Metric) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if metrics != nil {
		return p.encoder.Encode(metrics)
	}
	return errors.New("can't write metric to file from memory - object is empty")
}

func (p *Producer) Close() error {
	return p.file.Close()
}

type consumer struct {
	mu      sync.Mutex
	file    *os.File
	decoder *json.Decoder
}

func NewConsumer(fileName string) (*consumer, error) {
	file, err := os.OpenFile(fileName, os.O_RDONLY|os.O_CREATE, 0777)
	if err != nil {
		return nil, err
	}
	return &consumer{
		file:    file,
		decoder: json.NewDecoder(file),
	}, nil
}
func (c *consumer) ReadMetrics() (map[string]serializers.Metric, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	metrics := map[string]serializers.Metric{}
	if err := c.decoder.Decode(&metrics); err != nil {
		if !errors.Is(err, io.EOF) {
			return nil, err
		}
	}
	return metrics, nil
}
func (c *consumer) Close() error {
	return c.file.Close()
}
