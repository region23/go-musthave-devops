package main

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/region23/go-musthave-devops/internal/server"
	"github.com/region23/go-musthave-devops/internal/server/storage"
	"github.com/stretchr/testify/require"
)

var key = "test"

// executeRequest, creates a new ResponseRecorder
// then executes the request by calling ServeHTTP in the router
// after which the handler writes the response to the response recorder
// which we can then inspect.
func executeRequest(req *http.Request, s *server.Server) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	s.Router.ServeHTTP(rr, req)

	return rr
}

// checkResponseCode is a simple utility to check the response code
// of the response
func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected response code %d. Got %d\n", expected, actual)
	}
}

func TestUnknownHandlers(t *testing.T) {
	tests := []struct {
		name           string
		endpointURL    string
		wantStatusCode int
	}{
		{
			name:           "update_invalid_type",
			endpointURL:    "/update/unknown/testCounter/100",
			wantStatusCode: http.StatusNotImplemented,
		},
		{
			name:           "update_invalid_method",
			endpointURL:    "/updater/counter/testCounter/100",
			wantStatusCode: http.StatusNotFound,
		},
	}

	// Create a New Server Struct
	repository := storage.NewInMemory()
	srv := server.New(repository, key, nil)
	srv.MountHandlers()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, tt.endpointURL, nil)
			// Execute Request
			response := executeRequest(request, srv)

			// Check the response code
			checkResponseCode(t, tt.wantStatusCode, response.Code)
		})
	}
}

func TestGaugeHandlers(t *testing.T) {
	tests := []struct {
		name           string
		endpointURL    string
		wantStatusCode int
	}{
		{
			name:           "invalid_value",
			endpointURL:    "/update/gauge/testGauge/none",
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "without_id",
			endpointURL:    "/update/gauge/",
			wantStatusCode: http.StatusNotFound,
		},
		{
			name:           "update",
			endpointURL:    "/update/gauge/testGauge/100",
			wantStatusCode: http.StatusOK,
		},
	}

	// Create a New Server Struct
	repository := storage.NewInMemory()
	srv := server.New(repository, key, nil)
	srv.MountHandlers()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, tt.endpointURL, nil)
			// Execute Request
			response := executeRequest(request, srv)

			// Check the response code
			checkResponseCode(t, tt.wantStatusCode, response.Code)
		})
	}
}

func TestCounterHandlers(t *testing.T) {
	tests := []struct {
		name           string
		endpointURL    string
		wantStatusCode int
	}{
		{
			name:           "invalid_value",
			endpointURL:    "/update/counter/testCounter/none",
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "without_id",
			endpointURL:    "/update/counter/",
			wantStatusCode: http.StatusNotFound,
		},
		{
			name:           "update",
			endpointURL:    "/update/counter/testCounter/100",
			wantStatusCode: http.StatusOK,
		},
	}

	// Create a New Server Struct
	repository := storage.NewInMemory()
	srv := server.New(repository, key, nil)
	srv.MountHandlers()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, tt.endpointURL, nil)
			// Execute Request
			response := executeRequest(request, srv)

			// Check the response code
			checkResponseCode(t, tt.wantStatusCode, response.Code)
		})
	}
}

func TestCounter(t *testing.T) {
	tests := []struct {
		name       string
		metricName string
		value      string
		onlyValue  bool
	}{
		{
			name:       "update_sequence #1",
			metricName: "testSetGet33",
			value:      "527",
			onlyValue:  false,
		},
		{
			name:       "update_sequence #2",
			metricName: "testSetGet33",
			value:      "455",
			onlyValue:  false,
		},
		{
			name:       "update_sequence #3",
			metricName: "testSetGet33",
			value:      "187",
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
				var curValue int64
				request0 := httptest.NewRequest(http.MethodGet, "/value/counter/"+tt.metricName, nil)
				response0 := executeRequest(request0, srv)
				curValue, err := strconv.ParseInt(response0.Body.String(), 10, 64)
				if err != nil {
					curValue = 0
				}

				if newValue, err := strconv.ParseInt(tt.value, 10, 64); err == nil {
					wantValue := curValue + newValue
					request := httptest.NewRequest(http.MethodPost, "/update/counter/"+tt.metricName+"/"+tt.value, nil)
					// Execute Request
					executeRequest(request, srv)
					// Check the response code
					request2 := httptest.NewRequest(http.MethodGet, "/value/counter/"+tt.metricName, nil)
					response := executeRequest(request2, srv)
					serverValue := response.Body.String()
					require.Equal(t, strconv.FormatInt(wantValue, 10), serverValue)
				} else {
					t.Fail()
				}
			} else {
				request2 := httptest.NewRequest(http.MethodGet, "/value/counter/"+tt.metricName, nil)
				response := executeRequest(request2, srv)
				checkResponseCode(t, http.StatusNotFound, response.Code)
			}

		})
	}
}

func TestGauge(t *testing.T) {
	tests := []struct {
		name       string
		metricName string
		value      string
		onlyValue  bool
	}{
		{
			name:       "update_sequence #1",
			metricName: "testSetGet134",
			value:      "65637.019",
			onlyValue:  false,
		},
		{
			name:       "update_sequence #2",
			metricName: "testSetGet134",
			value:      "156519.255",
			onlyValue:  false,
		},
		{
			name:       "update_sequence #3",
			metricName: "testSetGet134",
			value:      "96969.519",
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
				request := httptest.NewRequest(http.MethodPost, "/update/gauge/"+tt.metricName+"/"+tt.value, nil)
				// Execute Request
				executeRequest(request, srv)
				// Check the response code
				request2 := httptest.NewRequest(http.MethodGet, "/value/gauge/"+tt.metricName, nil)
				response := executeRequest(request2, srv)
				require.Equal(t, tt.value, response.Body.String())
			} else {
				request2 := httptest.NewRequest(http.MethodGet, "/value/gauge/"+tt.metricName, nil)
				response := executeRequest(request2, srv)
				checkResponseCode(t, http.StatusNotFound, response.Code)
			}

		})
	}
}
