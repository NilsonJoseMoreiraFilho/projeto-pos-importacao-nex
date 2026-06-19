package outbound

import (
	"context"

	"projeto_pos/backend/internal/application"
	"projeto_pos/backend/internal/domain"
)

type ManagementSystemExporterPort interface {
	ExportApprovedPurchase(ctx context.Context, purchase domain.ApprovedPurchase) (application.ExportResult, error)
}
