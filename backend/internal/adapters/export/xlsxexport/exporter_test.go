package xlsxexport

import (
	"context"
	"os"
	"testing"

	"github.com/xuri/excelize/v2"

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
					SupplierProductCode:      "COD-1",
					Barcode:                  "789",
					Reference:                "REF-1",
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
	file, err := excelize.OpenFile(path)
	if err != nil {
		t.Fatalf("open xlsx: %v", err)
	}
	defer file.Close()
	headers, err := file.GetRows("purchase_items")
	if err != nil {
		t.Fatalf("read sheet: %v", err)
	}
	if len(headers) == 0 {
		t.Fatal("expected header row")
	}
	for _, header := range headers[0] {
		if header == "matched_internal_product_id" {
			t.Fatal("xlsx should not expose internal NEX product match column")
		}
	}
	code, err := file.GetCellValue("purchase_items", "D2")
	if err != nil {
		t.Fatalf("read supplier product code: %v", err)
	}
	if code != "COD-1" {
		t.Fatalf("unexpected supplier product code: %s", code)
	}
}
