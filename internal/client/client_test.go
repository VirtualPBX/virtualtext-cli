package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetMetrics(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/ops/metrics/response" {
			t.Fatalf("path %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer secret" {
			t.Fatalf("auth %s", r.Header.Get("Authorization"))
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"metrics": map[string]int{"count": 2}})
	}))
	defer srv.Close()

	raw, err := New(srv.URL, "secret").Get("/api/ops/metrics/response", nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(raw) == 0 {
		t.Fatal("empty body")
	}
}

func TestMissingAuth(t *testing.T) {
	_, err := New("", "").Get("/api/ops/metrics/response", nil)
	if err == nil {
		t.Fatal("expected error")
	}
}
