package outbound

import (
	"context"

	"projeto_pos/backend/internal/application"
)

type ProposalExporterPort interface {
	Export(ctx context.Context, proposal application.ImportProposal, path string) error
}
