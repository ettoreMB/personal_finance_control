package server_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ettoreMB/personal_finance_control/api/internal/server"
)

func TestHealthzReturnsOK(t *testing.T) {
	app := server.New()

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}
}
