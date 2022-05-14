package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/region23/go-musthave-devops/internal/server"
	"github.com/region23/go-musthave-devops/internal/server/storage"
)

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
		endpointUrl    string
		wantStatusCode int
	}{
		{
			name:           "update_invalid_type",
			endpointUrl:    "/update/unknown/testCounter/100",
			wantStatusCode: http.StatusNotImplemented,
		},
		{
			name:           "update_invalid_method",
			endpointUrl:    "/updater/counter/testCounter/100",
			wantStatusCode: http.StatusNotFound,
		},
	}

	// Create a New Server Struct
	repository := storage.NewInMemory()
	srv := server.New(repository)
	srv.MountHandlers()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, tt.endpointUrl, nil)
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
		endpointUrl    string
		wantStatusCode int
	}{
		{
			name:           "invalid_value",
			endpointUrl:    "/update/gauge/testGauge/none",
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "without_id",
			endpointUrl:    "/update/gauge/",
			wantStatusCode: http.StatusNotFound,
		},
		{
			name:           "update",
			endpointUrl:    "/update/gauge/testGauge/100",
			wantStatusCode: http.StatusOK,
		},
	}

	// Create a New Server Struct
	repository := storage.NewInMemory()
	srv := server.New(repository)
	srv.MountHandlers()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, tt.endpointUrl, nil)
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
		endpointUrl    string
		wantStatusCode int
	}{
		{
			name:           "invalid_value",
			endpointUrl:    "/update/counter/testCounter/none",
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "without_id",
			endpointUrl:    "/update/counter/",
			wantStatusCode: http.StatusNotFound,
		},
		{
			name:           "update",
			endpointUrl:    "/update/counter/testCounter/100",
			wantStatusCode: http.StatusOK,
		},
	}

	// Create a New Server Struct
	repository := storage.NewInMemory()
	srv := server.New(repository)
	srv.MountHandlers()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, tt.endpointUrl, nil)
			// Execute Request
			response := executeRequest(request, srv)

			// Check the response code
			checkResponseCode(t, tt.wantStatusCode, response.Code)
		})
	}
}
