package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/region23/go-musthave-devops/internal/serializers"
	"github.com/region23/go-musthave-devops/internal/server"
	"github.com/region23/go-musthave-devops/internal/server/storage"
	"github.com/stretchr/testify/require"
)

var key = "test"

func TestUnknownHandlersJSON(t *testing.T) {
	tests := []struct {
		name           string
		endpointURL    string
		metric         serializers.Metrics
		wantStatusCode int
	}{
		{
			name:           "update_invalid_type",
			endpointURL:    "/update",
			metric:         serializers.NewMetrics("testCounter", "unknown", 100),
			wantStatusCode: http.StatusNotImplemented,
		},
		{
			name:           "update_invalid_method",
			endpointURL:    "/updater",
			metric:         serializers.NewMetrics("testCounter", "counter", 100),
			wantStatusCode: http.StatusNotFound,
		},
	}

	// Create a New Server Struct
	repository := storage.NewInMemory()
	srv := server.New(repository, key)
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
	tests := []struct {
		name           string
		endpointURL    string
		metric         serializers.Metrics
		wantStatusCode int
	}{
		{
			name:           "invalid_value",
			endpointURL:    "/update",
			metric:         serializers.NewMetrics("testGauge", "gauge", "none"),
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "without_id",
			endpointURL:    "/update",
			metric:         serializers.NewMetrics("", "gauge"),
			wantStatusCode: http.StatusNotFound,
		},
		{
			name:           "update",
			endpointURL:    "/update",
			metric:         serializers.NewMetrics("testGauge", "gauge", 100),
			wantStatusCode: http.StatusOK,
		},
	}

	// Create a New Server Struct
	repository := storage.NewInMemory()
	srv := server.New(repository, key)
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
	tests := []struct {
		name           string
		endpointURL    string
		metric         serializers.Metrics
		wantStatusCode int
	}{
		{
			name:           "invalid_value",
			endpointURL:    "/update",
			metric:         serializers.NewMetrics("testCounter", "counter", "none"),
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "without_id",
			endpointURL:    "/update",
			metric:         serializers.NewMetrics("", "counter"),
			wantStatusCode: http.StatusNotFound,
		},
		{
			name:           "update",
			endpointURL:    "/update",
			metric:         serializers.NewMetrics("testCounter", "counter", 100),
			wantStatusCode: http.StatusOK,
		},
	}

	// Create a New Server Struct
	repository := storage.NewInMemory()
	srv := server.New(repository, key)
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
	srv := server.New(repository, key)
	srv.MountHandlers()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.onlyValue == false {
				// получаем текущее значение метрики
				beforeMetric := serializers.NewMetrics(tt.metricName, "counter")
				postBody, err := json.Marshal(beforeMetric)

				if err != nil {
					t.Fatal(err)
				}

				responseBody := bytes.NewBuffer(postBody)
				request0 := httptest.NewRequest(http.MethodPost, "/value", responseBody)
				request0.Header.Set("Content-Type", "application/json")
				response0 := executeRequest(request0, srv)
				var returnedMetric serializers.Metrics
				if response0.Code == http.StatusOK {
					// decode input or return error
					err = json.NewDecoder(response0.Body).Decode(&returnedMetric)
					if err != nil {
						t.Fatal(err)
					}
				}

				newMetric := serializers.NewMetrics(tt.metricName, "counter", tt.value)
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
					var afterMetric serializers.Metrics
					err = json.NewDecoder(response2.Body).Decode(&afterMetric)
					if err != nil {
						t.Fatal(err)
					}
					require.Equal(t, wantValue, *afterMetric.Delta)
				} else {
					t.Fatal("Value not found")
				}

			} else {
				metric := serializers.NewMetrics(tt.metricName, "counter")
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
	srv := server.New(repository, key)
	srv.MountHandlers()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.onlyValue {
				metric := serializers.NewMetrics(tt.metricName, "gauge", tt.value)
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
				var newMetric serializers.Metrics
				err = json.NewDecoder(response1.Body).Decode(&newMetric)
				if err != nil {
					t.Fatal(err)
				}
				require.Equal(t, tt.value, *newMetric.Value)
			} else {
				metric := serializers.NewMetrics(tt.metricName, "gauge")
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
