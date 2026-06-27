package imageocr

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSidecarJSONClientUsesFallbackWhenOCRSidecarIsMissing(t *testing.T) {
	imagePath := filepath.Join(t.TempDir(), "pedido.jpg")
	client := SidecarJSONClient{
		Suffix:        ".ocr.json",
		FallbackPaths: []string{filepath.Join("..", "..", "..", "..", "testdata", "sample-import-proposal.json")},
	}

	raw, err := client.Extract(context.Background(), imagePath)
	if err != nil {
		t.Fatalf("extract with fallback failed: %v", err)
	}

	if raw.Purchase.DocumentNumber != "11441701" {
		t.Fatalf("unexpected document number: %s", raw.Purchase.DocumentNumber)
	}
	if raw.SourceDocument.FileName != "pedido.jpg" {
		t.Fatalf("expected uploaded image file name, got %s", raw.SourceDocument.FileName)
	}
	if !hasObservation(raw.SourceDocument.Observations, "fallback") {
		t.Fatalf("expected fallback observation, got %#v", raw.SourceDocument.Observations)
	}
}

func TestOpenAIClientExtractsStructuredPurchase(t *testing.T) {
	imagePath := filepath.Join(t.TempDir(), "pedido.jpg")
	if err := os.WriteFile(imagePath, []byte("fake image"), 0644); err != nil {
		t.Fatalf("write image: %v", err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Fatalf("unexpected authorization header: %s", r.Header.Get("Authorization"))
		}
		var request map[string]any
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if request["model"] != "test-model" {
			t.Fatalf("unexpected model: %v", request["model"])
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"output": [{
				"type": "message",
				"content": [{
					"type": "output_text",
					"text": "{\"supplier\":{\"legalName\":\"Fornecedor IA\",\"tradeName\":\"\",\"documentNumber\":\"\",\"stateRegistration\":\"\"},\"documentNumber\":\"987\",\"issueDate\":\"2026-05-01\",\"priceTable\":\"2-ATACADO\",\"freightMode\":\"FOB\",\"pageCount\":1,\"currentPage\":1,\"items\":[{\"lineNumber\":1,\"supplierProductCode\":\"ABC\",\"barcode\":\"\",\"reference\":\"REF\",\"description\":\"Produto IA\",\"unit\":\"UNID\",\"packageQuantity\":1,\"quantity\":2,\"unitCost\":3.5,\"totalCost\":7}],\"totals\":{\"productsTotal\":7,\"discount\":0,\"addition\":0,\"ipi\":0,\"taxSubstitution\":0,\"fcpSt\":0,\"grandTotal\":7},\"warnings\":[]}"
				}]
			}]
		}`))
	}))
	defer server.Close()

	client := OpenAIClient{
		APIKey:     "test-key",
		Model:      "test-model",
		Endpoint:   server.URL,
		HTTPClient: server.Client(),
	}
	raw, err := client.Extract(context.Background(), imagePath)
	if err != nil {
		t.Fatalf("extract failed: %v", err)
	}
	if raw.Purchase.DocumentNumber != "987" {
		t.Fatalf("unexpected document number: %s", raw.Purchase.DocumentNumber)
	}
	if raw.Purchase.Items[0].TotalCost.Cents != 700 {
		t.Fatalf("unexpected total cents: %d", raw.Purchase.Items[0].TotalCost.Cents)
	}
	if raw.ExtractorName != "openai-vision-reader" {
		t.Fatalf("unexpected extractor: %s", raw.ExtractorName)
	}
}

func hasObservation(observations []string, term string) bool {
	for _, observation := range observations {
		if strings.Contains(observation, term) {
			return true
		}
	}
	return false
}
