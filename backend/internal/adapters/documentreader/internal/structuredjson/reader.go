package structuredjson

import (
	"context"
	"encoding/json"
	"os"

	"projeto_pos/backend/internal/application"
)

func ReadRawExtraction(_ context.Context, path string) (application.RawDocumentExtraction, error) {
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
