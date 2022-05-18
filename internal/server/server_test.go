package server

import (
	"net/http"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/region23/go-musthave-devops/internal/server/storage"
)

func TestServer_UpdateMetric(t *testing.T) {
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
			s.UpdateMetric(tt.args.w, tt.args.r)
		})
	}
}

func TestServer_GetMetric(t *testing.T) {
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
			s.GetMetric(tt.args.w, tt.args.r)
		})
	}
}

func TestServer_AllMetrics(t *testing.T) {
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
			s.AllMetrics(tt.args.w, tt.args.r)
		})
	}
}
