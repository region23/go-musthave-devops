package storage

import (
	"encoding/json"
	"os"

	"github.com/region23/go-musthave-devops/internal/serializers"
)

type Producer struct {
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

func (p *Producer) WriteMetrics(metrics *map[string]serializers.Metrics) error {
	return p.encoder.Encode(&metrics)
}

func (p *Producer) Close() error {
	return p.file.Close()
}

type consumer struct {
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
func (c *consumer) ReadMetrics() (*map[string]serializers.Metrics, error) {
	metrics := &map[string]serializers.Metrics{}
	if err := c.decoder.Decode(&metrics); err != nil {
		return nil, err
	}
	return metrics, nil
}
func (c *consumer) Close() error {
	return c.file.Close()
}
