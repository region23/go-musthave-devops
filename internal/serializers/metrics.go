package serializers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
)

type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
	Hash  string   `json:"hash,omitempty"`  // значение хеш-функции
}

func NewMetrics(id string, mtype string, val ...interface{}) Metrics {
	m := Metrics{ID: id, MType: mtype}
	if len(val) == 0 {
		return m
	}
	switch v := val[0].(type) {
	case int64:
		m.Delta = &v
	case int:
		i := int64(v)
		m.Delta = &i
	case float64:
		m.Value = &v
	case string:
		if mtype == "counter" {
			convertedV, err := strconv.ParseInt(v, 10, 64)
			if err == nil {
				m.Delta = &convertedV
			}
		} else if mtype == "gauge" {
			convertedV, err := strconv.ParseFloat(v, 64)
			if err == nil {
				m.Value = &convertedV
			}
		}
	default:
	}
	return m
}

func Hash(mType, mName, mValue, key string) string {
	str := fmt.Sprintf("%s:%s:%s", mName, mType, mValue)
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}
