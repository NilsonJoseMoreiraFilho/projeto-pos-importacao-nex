package outbound

import (
	"context"

	"projeto_pos/backend/internal/application"
)

type DocumentReaderPort interface {
	Read(ctx context.Context, path string) (application.RawDocumentExtraction, error)
}
