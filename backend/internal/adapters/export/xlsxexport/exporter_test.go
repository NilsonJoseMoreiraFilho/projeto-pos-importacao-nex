package xlsxexport

import (
	"context"
	"os"
	"testing"

	"projeto_pos/backend/internal/domain"
)

func TestExportApprovedPurchaseWritesXLSX(t *testing.T) {
	path := t.TempDir() + "/purchase.xlsx"
	_, err := NewExporter(path).ExportApprovedPurchase(context.Background(), domain.ApprovedPurchase{
		Purchase: domain.Purchase{
			DocumentNumber: "123",
			Supplier:       domain.Supplier{LegalName: "Fornecedor Teste"},
			Items: []domain.PurchaseItem{
				{
					LineNumber:               1,
					Description:              "Produto A",
					Unit:                     "UNID",
					Quantity:                 1,
					UnitCost:                 domain.NewMoneyFromFloat(5),
					TotalCost:                domain.NewMoneyFromFloat(5),
					MatchedInternalProductID: "prod-a",
				},
			},
			Totals: domain.PurchaseTotals{
				ProductsTotal: domain.NewMoneyFromFloat(5),
				GrandTotal:    domain.NewMoneyFromFloat(5),
			},
		},
	})
	if err != nil {
		t.Fatalf("export failed: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected xlsx file: %v", err)
	}
}
