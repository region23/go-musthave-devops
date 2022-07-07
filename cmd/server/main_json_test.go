package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/region23/go-musthave-devops/internal/serializers"
	"github.com/region23/go-musthave-devops/internal/server"
	"github.com/region23/go-musthave-devops/internal/server/storage"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/require"
)

func NewMetric(id string, mtype string, v string) serializers.Metric {
	metric := serializers.Metric{ID: id, MType: mtype}

	if v != "" && v != "none" {
		if mtype == "counter" {
			convertedV, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				log.Error().Err(err).Msg("ошибка при парсинге значения счетчика метрики")
				return metric
			}
			metric.Delta = &convertedV
		} else if mtype == "gauge" {
			convertedV, err := strconv.ParseFloat(v, 64)
			if err != nil {
				log.Error().Err(err).Msg("ошибка при парсинге значения метрики")
				return metric
			}
			metric.Value = &convertedV
		}
	}

	return metric
}

func TestUnknownHandlersJSON(t *testing.T) {
	m1 := NewMetric("testCounter", "unknown", "100")
	m2 := NewMetric("testCounter", "counter", "100")
	tests := []struct {
		name           string
		endpointURL    string
		metric         serializers.Metric
		wantStatusCode int
	}{
		{
			name:           "update_invalid_type",
			endpointURL:    "/update",
			metric:         m1,
			wantStatusCode: http.StatusNotImplemented,
		},
		{
			name:           "update_invalid_method",
			endpointURL:    "/updater",
			metric:         m2,
			wantStatusCode: http.StatusNotFound,
		},
	}

	// Create a New Server Struct
	repository := storage.NewInMemory()
	srv := server.New(repository, key, nil)
	srv.MountHandlers()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			postBody, err := json.Marshal(tt.metric)

			if err != nil {
				t.Error(err)
			}

			responseBody := bytes.NewBuffer(postBody)
			request := httptest.NewRequest(http.MethodPost, tt.endpointURL, responseBody)
			request.Header.Set("Content-Type", "application/json")
			// Execute Request
			response := executeRequest(request, srv)

			// Check the response code
			checkResponseCode(t, tt.wantStatusCode, response.Code)
		})
	}
}

func TestGaugeHandlersJSON(t *testing.T) {
	m1 := NewMetric("testGauge", "gauge", "none")
	m2 := NewMetric("", "gauge", "")
	m3 := NewMetric("testGauge", "gauge", "100")
	tests := []struct {
		name           string
		endpointURL    string
		metric         serializers.Metric
		wantStatusCode int
	}{
		{
			name:           "invalid_value",
			endpointURL:    "/update",
			metric:         m1,
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "without_id",
			endpointURL:    "/update",
			metric:         m2,
			wantStatusCode: http.StatusNotFound,
		},
		{
			name:           "update",
			endpointURL:    "/update",
			metric:         m3,
			wantStatusCode: http.StatusOK,
		},
	}

	// Create a New Server Struct
	repository := storage.NewInMemory()
	srv := server.New(repository, key, nil)
	srv.MountHandlers()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			postBody, err := json.Marshal(tt.metric)

			if err != nil {
				t.Error(err)
			}

			responseBody := bytes.NewBuffer(postBody)
			request := httptest.NewRequest(http.MethodPost, tt.endpointURL, responseBody)
			request.Header.Set("Content-Type", "application/json")
			// Execute Request
			response := executeRequest(request, srv)

			// Check the response code
			checkResponseCode(t, tt.wantStatusCode, response.Code)
		})
	}
}

func TestCounterHandlersJSON(t *testing.T) {
	m1 := NewMetric("testCounter", "counter", "none")
	m2 := NewMetric("", "counter", "")
	m3 := NewMetric("testCounter", "counter", "100")

	tests := []struct {
		name           string
		endpointURL    string
		metric         serializers.Metric
		wantStatusCode int
	}{
		{
			name:           "invalid_value",
			endpointURL:    "/update",
			metric:         m1,
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "without_id",
			endpointURL:    "/update",
			metric:         m2,
			wantStatusCode: http.StatusNotFound,
		},
		{
			name:           "update",
			endpointURL:    "/update",
			metric:         m3,
			wantStatusCode: http.StatusOK,
		},
	}

	// Create a New Server Struct
	repository := storage.NewInMemory()
	srv := server.New(repository, key, nil)
	srv.MountHandlers()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			postBody, err := json.Marshal(tt.metric)

			if err != nil {
				t.Error(err)
			}

			responseBody := bytes.NewBuffer(postBody)
			request := httptest.NewRequest(http.MethodPost, tt.endpointURL, responseBody)
			request.Header.Set("Content-Type", "application/json")
			// Execute Request
			response := executeRequest(request, srv)

			// Check the response code
			checkResponseCode(t, tt.wantStatusCode, response.Code)
		})
	}
}

func TestCounterJSON(t *testing.T) {
	tests := []struct {
		name       string
		metricName string
		value      int64
		onlyValue  bool
	}{
		{
			name:       "update_sequence #1",
			metricName: "testSetGet33",
			value:      527,
			onlyValue:  false,
		},
		{
			name:       "update_sequence #2",
			metricName: "testSetGet33",
			value:      455,
			onlyValue:  false,
		},
		{
			name:       "update_sequence #3",
			metricName: "testSetGet33",
			value:      187,
			onlyValue:  false,
		},
		{
			name:       "get_unknown",
			metricName: "testUnknown129",
			onlyValue:  true,
		},
	}

	// Create a New Server Struct
	repository := storage.NewInMemory()
	srv := server.New(repository, key, nil)
	srv.MountHandlers()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.onlyValue == false {
				// получаем текущее значение метрики
				beforeMetric := NewMetric(tt.metricName, "counter", "")

				postBody, err := json.Marshal(beforeMetric)

				if err != nil {
					t.Fatal(err)
				}

				responseBody := bytes.NewBuffer(postBody)
				request0 := httptest.NewRequest(http.MethodPost, "/value", responseBody)
				request0.Header.Set("Content-Type", "application/json")
				response0 := executeRequest(request0, srv)
				var returnedMetric serializers.Metric
				if response0.Code == http.StatusOK {
					// decode input or return error
					err = json.NewDecoder(response0.Body).Decode(&returnedMetric)
					if err != nil {
						t.Fatal(err)
					}
				}

				newMetric := NewMetric(tt.metricName, "counter", fmt.Sprintf("%d", tt.value))
				if err != nil {
					t.Fatal(err)
				}

				var wantValue int64
				if returnedMetric.Delta != nil {
					wantValue = *returnedMetric.Delta + *newMetric.Delta
				} else {
					wantValue = *newMetric.Delta
				}
				postBody, err = json.Marshal(newMetric)

				if err != nil {
					t.Fatal(err)
				}

				responseBody = bytes.NewBuffer(postBody)

				request1 := httptest.NewRequest(http.MethodPost, "/update", responseBody)
				request1.Header.Set("Content-Type", "application/json")
				response1 := executeRequest(request1, srv)
				if response1.Code == http.StatusOK {
					responseBody := bytes.NewBuffer(postBody)
					request2 := httptest.NewRequest(http.MethodPost, "/value", responseBody)
					request2.Header.Set("Content-Type", "application/json")
					response2 := executeRequest(request2, srv)
					var afterMetric serializers.Metric
					err = json.NewDecoder(response2.Body).Decode(&afterMetric)
					if err != nil {
						t.Fatal(err)
					}
					require.Equal(t, wantValue, *afterMetric.Delta)
				} else {
					t.Fatal("Value not found")
				}

			} else {
				metric := NewMetric(tt.metricName, "counter", "")

				postBody, err := json.Marshal(metric)

				if err != nil {
					t.Fatal(err)
				}

				responseBody := bytes.NewBuffer(postBody)
				request0 := httptest.NewRequest(http.MethodPost, "/value", responseBody)
				request0.Header.Set("Content-Type", "application/json")
				response0 := executeRequest(request0, srv)

				checkResponseCode(t, http.StatusNotFound, response0.Code)
			}

		})
	}
}

func TestGaugeJSON(t *testing.T) {
	tests := []struct {
		name       string
		metricName string
		value      float64
		onlyValue  bool
	}{
		{
			name:       "update_sequence #1",
			metricName: "testSetGet134",
			value:      65637.019,
			onlyValue:  false,
		},
		{
			name:       "update_sequence #2",
			metricName: "testSetGet134",
			value:      156519.255,
			onlyValue:  false,
		},
		{
			name:       "update_sequence #3",
			metricName: "testSetGet134",
			value:      96969.519,
			onlyValue:  false,
		},
		{
			name:       "get_unknown",
			metricName: "testUnknown164",
			onlyValue:  true,
		},
	}

	// Create a New Server Struct
	repository := storage.NewInMemory()
	srv := server.New(repository, key, nil)
	srv.MountHandlers()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.onlyValue {
				metric := NewMetric(tt.metricName, "gauge", fmt.Sprintf("%f", tt.value))

				postBody, err := json.Marshal(metric)

				if err != nil {
					t.Fatal(err)
				}

				responseBody := bytes.NewBuffer(postBody)
				request := httptest.NewRequest(http.MethodPost, "/update", responseBody)
				request.Header.Set("Content-Type", "application/json")

				// Execute Request
				executeRequest(request, srv)
				// Check the response code
				responseBody = bytes.NewBuffer(postBody)
				request1 := httptest.NewRequest(http.MethodPost, "/value", responseBody)
				request1.Header.Set("Content-Type", "application/json")
				response1 := executeRequest(request1, srv)
				var newMetric serializers.Metric
				err = json.NewDecoder(response1.Body).Decode(&newMetric)
				if err != nil {
					t.Fatal(err)
				}
				require.Equal(t, tt.value, *newMetric.Value)
			} else {
				metric := NewMetric(tt.metricName, "gauge", "")

				postBody, err := json.Marshal(metric)

				if err != nil {
					t.Fatal(err)
				}

				responseBody := bytes.NewBuffer(postBody)

				request2 := httptest.NewRequest(http.MethodPost, "/value", responseBody)
				request2.Header.Set("Content-Type", "application/json")
				response2 := executeRequest(request2, srv)
				checkResponseCode(t, http.StatusNotFound, response2.Code)
			}

		})
	}
}
