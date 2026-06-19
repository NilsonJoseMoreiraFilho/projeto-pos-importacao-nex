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
	Suffix string
}

func NewSidecarJSONClient() SidecarJSONClient {
	return SidecarJSONClient{Suffix: ".ocr.json"}
}

func (c SidecarJSONClient) Extract(ctx context.Context, imagePath string) (application.RawDocumentExtraction, error) {
	sidecar := imagePath + c.Suffix
	if _, err := os.Stat(sidecar); err != nil {
		return application.RawDocumentExtraction{}, err
	}
	return structuredjson.ReadRawExtraction(ctx, sidecar)
}

func isImage(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".jpg", ".jpeg", ".png", ".webp", ".tif", ".tiff":
		return true
	default:
		return false
	}
}
