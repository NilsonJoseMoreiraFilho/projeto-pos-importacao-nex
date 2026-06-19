package manualjson

import (
	"context"
	"encoding/json"
	"os"

	"projeto_pos/backend/internal/application"
)

type Reader struct{}

func NewReader() Reader {
	return Reader{}
}

func (r Reader) Read(_ context.Context, path string) (application.RawDocumentExtraction, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return application.RawDocumentExtraction{}, err
	}
	var raw application.RawDocumentExtraction
	if err := json.Unmarshal(data, &raw); err != nil {
		return application.RawDocumentExtraction{}, err
	}
	return raw, nil
}
