package manualjson

import (
	"context"

	"projeto_pos/backend/internal/adapters/documentreader/internal/structuredjson"
	"projeto_pos/backend/internal/application"
)

type Reader struct{}

func NewReader() Reader {
	return Reader{}
}

func (r Reader) Read(ctx context.Context, path string) (application.RawDocumentExtraction, error) {
	return structuredjson.ReadRawExtraction(ctx, path)
}
