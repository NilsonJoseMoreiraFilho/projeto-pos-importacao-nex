package imageocr

import (
	"context"
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

func hasObservation(observations []string, term string) bool {
	for _, observation := range observations {
		if strings.Contains(observation, term) {
			return true
		}
	}
	return false
}
