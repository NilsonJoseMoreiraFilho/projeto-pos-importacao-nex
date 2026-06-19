package csvexport

import (
	"context"
	"os"
	"time"

	"projeto_pos/backend/internal/adapters/export/internal/purchasewriter"
	"projeto_pos/backend/internal/application"
	"projeto_pos/backend/internal/domain"
)

type Exporter struct {
	Path string
}

func NewExporter() Exporter {
	return Exporter{}
}

func NewManagementExporter(path string) Exporter {
	return Exporter{Path: path}
}

func (e Exporter) Export(_ context.Context, proposal application.ImportProposal, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	return purchasewriter.WritePurchaseItemsCSV(file, proposal.PurchaseDraft)
}

func (e Exporter) ExportApprovedPurchase(_ context.Context, purchase domain.ApprovedPurchase) (application.ExportResult, error) {
	file, err := os.Create(e.Path)
	if err != nil {
		return application.ExportResult{}, err
	}
	defer file.Close()

	if err := purchasewriter.WritePurchaseItemsCSV(file, purchase.Purchase); err != nil {
		return application.ExportResult{}, err
	}
	return application.ExportResult{
		Status:      application.ExportStatusSuccess,
		Destination: "csv",
		Reference:   e.Path,
		CreatedAt:   time.Now().UTC(),
	}, nil
}
