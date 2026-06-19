package csvexport

import (
	"context"
	"os"
	"strings"
	"testing"

	"projeto_pos/backend/internal/domain"
)

func TestExportApprovedPurchaseWritesCSV(t *testing.T) {
	path := t.TempDir() + "/purchase.csv"
	result, err := NewManagementExporter(path).ExportApprovedPurchase(context.Background(), approvedPurchase())
	if err != nil {
		t.Fatalf("export failed: %v", err)
	}
	if result.Reference != path {
		t.Fatalf("unexpected reference: %s", result.Reference)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read csv: %v", err)
	}
	if !strings.Contains(string(data), "Produto A") {
		t.Fatalf("csv does not contain item: %s", string(data))
	}
}

func approvedPurchase() domain.ApprovedPurchase {
	return domain.ApprovedPurchase{
		Purchase: domain.Purchase{
			DocumentNumber: "123",
			Supplier:       domain.Supplier{LegalName: "Fornecedor Teste"},
			Items: []domain.PurchaseItem{
				{
					LineNumber:               1,
					Barcode:                  "789",
					Description:              "Produto A",
					Unit:                     "UNID",
					Quantity:                 2,
					UnitCost:                 domain.NewMoneyFromFloat(10),
					TotalCost:                domain.NewMoneyFromFloat(20),
					MatchedInternalProductID: "prod-a",
				},
			},
			Totals: domain.PurchaseTotals{
				ProductsTotal: domain.NewMoneyFromFloat(20),
				GrandTotal:    domain.NewMoneyFromFloat(20),
			},
		},
		ApprovedBy: "tester",
	}
}
