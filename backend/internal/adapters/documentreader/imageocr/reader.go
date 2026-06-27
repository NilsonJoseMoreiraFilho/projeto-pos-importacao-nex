package imageocr

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"projeto_pos/backend/internal/adapters/documentreader/internal/structuredjson"
	"projeto_pos/backend/internal/application"
)

type OCRClient interface {
	Extract(ctx context.Context, imagePath string) (application.RawDocumentExtraction, error)
}

type Reader struct {
	Client OCRClient
}

func NewReader(client OCRClient) Reader {
	return Reader{Client: client}
}

func (r Reader) Read(ctx context.Context, path string) (application.RawDocumentExtraction, error) {
	if !isImage(path) {
		return application.RawDocumentExtraction{}, fmt.Errorf("image OCR reader expects image file: %s", path)
	}
	if r.Client == nil {
		return application.RawDocumentExtraction{}, fmt.Errorf("image OCR reader requires OCR client")
	}
	return r.Client.Extract(ctx, path)
}

type SidecarJSONClient struct {
	Suffix        string
	FallbackPaths []string
}

func NewSidecarJSONClient() SidecarJSONClient {
	fallbacks := []string{
		os.Getenv("IMAGE_OCR_FALLBACK_JSON"),
		"frontend/demo-import-proposal.json",
		"../frontend/demo-import-proposal.json",
		"backend/testdata/sample-import-proposal.json",
		"../backend/testdata/sample-import-proposal.json",
		"testdata/sample-import-proposal.json",
	}
	return SidecarJSONClient{Suffix: ".ocr.json", FallbackPaths: fallbacks}
}

func (c SidecarJSONClient) Extract(ctx context.Context, imagePath string) (application.RawDocumentExtraction, error) {
	sidecar := imagePath + c.Suffix
	if _, err := os.Stat(sidecar); err == nil {
		return structuredjson.ReadRawExtraction(ctx, sidecar)
	}
	for _, fallback := range c.FallbackPaths {
		if strings.TrimSpace(fallback) == "" {
			continue
		}
		if _, err := os.Stat(fallback); err == nil {
			raw, err := structuredjson.ReadRawExtraction(ctx, fallback)
			if err != nil {
				return application.RawDocumentExtraction{}, err
			}
			raw.SourceDocument.FileName = filepath.Base(imagePath)
			raw.SourceDocument.FileType = "image/" + strings.TrimPrefix(strings.ToLower(filepath.Ext(imagePath)), ".")
			raw.SourceDocument.StorageLocation = imagePath
			raw.SourceDocument.Observations = append(raw.SourceDocument.Observations, "extracao demonstrativa por fallback; OCR real ainda nao implementado")
			return raw, nil
		}
	}
	return application.RawDocumentExtraction{}, fmt.Errorf("image OCR sidecar not found for %s and no fallback JSON is available", imagePath)
}

func isImage(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".jpg", ".jpeg", ".png", ".webp", ".tif", ".tiff":
		return true
	default:
		return false
	}
}
