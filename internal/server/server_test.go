package server

import (
	"net/http"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/region23/go-musthave-devops/internal/server/storage"
)

func TestServer_UpdateMetricHandler(t *testing.T) {
	type fields struct {
		repository storage.Repository
		Router     *chi.Mux
	}
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				repository: tt.fields.repository,
				Router:     tt.fields.Router,
			}
			s.UpdateMetricHandler(tt.args.w, tt.args.r)
		})
	}
}

func TestServer_GetMetricHandler(t *testing.T) {
	type fields struct {
		repository storage.Repository
		Router     *chi.Mux
	}
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				repository: tt.fields.repository,
				Router:     tt.fields.Router,
			}
			s.GetMetricHandler(tt.args.w, tt.args.r)
		})
	}
}

func TestServer_AllMetricsHandler(t *testing.T) {
	type fields struct {
		repository storage.Repository
		Router     *chi.Mux
	}
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				repository: tt.fields.repository,
				Router:     tt.fields.Router,
			}
			s.AllMetricsHandler(tt.args.w, tt.args.r)
		})
	}
}
