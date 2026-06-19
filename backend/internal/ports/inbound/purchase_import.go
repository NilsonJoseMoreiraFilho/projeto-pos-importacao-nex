package inbound

import (
	"context"

	"projeto_pos/backend/internal/application"
	"projeto_pos/backend/internal/domain"
)

type ImportPurchaseProposalPort interface {
	Import(ctx context.Context, input application.ImportPurchaseInput) (application.ImportProposal, error)
}

type ReviewImportProposalPort interface {
	Review(ctx context.Context, proposalID string, decisions []application.ReviewDecision) (application.ImportProposal, error)
}

type ApprovePurchasePort interface {
	Approve(ctx context.Context, proposalID string, reviewer string) (domain.ApprovedPurchase, error)
}

type ExportApprovedPurchasePort interface {
	Export(ctx context.Context, proposalID string) (application.ExportResult, error)
}
