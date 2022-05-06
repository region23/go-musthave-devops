package server

import (
	"net/http"
	"testing"

	"github.com/region23/go-musthave-devops/internal/server/storage"
)

func TestServer_UpdateHandler(t *testing.T) {
	type fields struct {
		repository storage.Repository
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
			}
			s.UpdateHandler(tt.args.w, tt.args.r)
		})
	}
}
