package pdf

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"projeto_pos/backend/internal/adapters/documentreader/internal/structuredjson"
	"projeto_pos/backend/internal/application"
	"projeto_pos/backend/internal/domain"
)

type Reader struct {
	SidecarSuffix string
}

func NewReader() Reader {
	return Reader{SidecarSuffix: ".json"}
}

func (r Reader) Read(ctx context.Context, path string) (application.RawDocumentExtraction, error) {
	sidecar := path + r.SidecarSuffix
	if _, err := os.Stat(sidecar); err == nil {
		return structuredjson.ReadRawExtraction(ctx, sidecar)
	}
	if filepath.Ext(path) != ".pdf" {
		return application.RawDocumentExtraction{}, fmt.Errorf("pdf reader expects .pdf file: %s", path)
	}
	if _, err := os.Stat(path); err != nil {
		return application.RawDocumentExtraction{}, err
	}
	return application.RawDocumentExtraction{
		SourceDocument: domain.SourceDocument{
			ID:                   filepath.Base(path),
			FileName:             filepath.Base(path),
			FileType:             "application/pdf",
			StorageLocation:      path,
			DetectedDocumentType: domain.SourceDocumentTypePDF,
			Observations:         []string{"pdf adapter registered document; structured extraction requires sidecar JSON or OCR/LLM pipeline"},
		},
		ExtractorName:    "pdf-sidecar-reader",
		ExtractorVersion: "0.1.0",
		ExtractedAt:      time.Now().UTC(),
	}, nil
}
