package nexapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"projeto_pos/backend/internal/domain"
)

func TestExportApprovedPurchasePostsToAPI(t *testing.T) {
	var called bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer token" {
			t.Fatalf("missing auth header")
		}
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	result, err := NewExporter(server.URL, "token").ExportApprovedPurchase(context.Background(), domain.ApprovedPurchase{})
	if err != nil {
		t.Fatalf("export failed: %v", err)
	}
	if !called {
		t.Fatal("expected API call")
	}
	if result.Destination != "nex_api" {
		t.Fatalf("unexpected destination: %s", result.Destination)
	}
}
