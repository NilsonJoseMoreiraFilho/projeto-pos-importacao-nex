package domain

import (
	"testing"
	"time"
)

func validPurchase() (SourceDocument, Purchase) {
	source := SourceDocument{
		ID:                   "doc-1",
		PageCount:            1,
		CurrentPage:          1,
		DetectedDocumentType: SourceDocumentTypeManualJSON,
	}
	purchase := Purchase{
		ID:             "purchase-1",
		Supplier:       Supplier{LegalName: "Fornecedor Teste"},
		DocumentNumber: "123",
		IssueDate:      time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC),
		Items: []PurchaseItem{
			{
				LineNumber:               1,
				Description:              "Produto A",
				Unit:                     "UNID",
				Quantity:                 2,
				UnitCost:                 NewMoneyFromFloat(10),
				TotalCost:                NewMoneyFromFloat(20),
				MatchedInternalProductID: "prod-a",
			},
			{
				LineNumber:               2,
				Description:              "Produto B",
				Unit:                     "UNID",
				Quantity:                 1,
				UnitCost:                 NewMoneyFromFloat(5.50),
				TotalCost:                NewMoneyFromFloat(5.50),
				MatchedInternalProductID: "prod-b",
			},
		},
		Totals: PurchaseTotals{
			ProductsTotal: NewMoneyFromFloat(25.50),
			GrandTotal:    NewMoneyFromFloat(25.50),
		},
	}
	return source, purchase
}

func TestValidatePurchaseAcceptsValidTotals(t *testing.T) {
	source, purchase := validPurchase()
	results := ValidatePurchase(source, purchase, false)
	if HasBlocking(results) {
		t.Fatalf("expected no blocking results, got %#v", results)
	}
}

func TestValidatePurchaseWarnsItemTotalMismatch(t *testing.T) {
	source, purchase := validPurchase()
	purchase.Items[0].TotalCost = NewMoneyFromFloat(19)
	results := ValidatePurchase(source, purchase, false)
	assertWarningCode(t, results, "ITEM_TOTAL_MISMATCH")
	if HasBlocking(results) {
		t.Fatalf("item total mismatch should not block human-approved import, got %#v", results)
	}
}

func TestValidatePurchaseWarnsProductsTotalMismatch(t *testing.T) {
	source, purchase := validPurchase()
	purchase.Totals.ProductsTotal = NewMoneyFromFloat(30)
	purchase.Totals.GrandTotal = NewMoneyFromFloat(30)
	results := ValidatePurchase(source, purchase, false)
	assertWarningCode(t, results, "PRODUCTS_TOTAL_MISMATCH")
	if HasBlocking(results) {
		t.Fatalf("products total mismatch should not block human-approved import, got %#v", results)
	}
}

func TestValidatePurchaseWarnsMissingPage(t *testing.T) {
	source, purchase := validPurchase()
	source.PageCount = 2
	source.CurrentPage = 1
	results := ValidatePurchase(source, purchase, false)
	assertWarningCode(t, results, "DOCUMENT_INCOMPLETE")
	if HasBlocking(results) {
		t.Fatalf("missing page should not block human-approved import, got %#v", results)
	}
}

func TestValidatePurchaseBlocksUnknownSupplier(t *testing.T) {
	source, purchase := validPurchase()
	purchase.Supplier = Supplier{}
	results := ValidatePurchase(source, purchase, false)
	assertCode(t, results, "SUPPLIER_NOT_IDENTIFIED")
}

func TestValidatePurchaseDoesNotBlockMissingProductMatch(t *testing.T) {
	source, purchase := validPurchase()
	purchase.Items[0].MatchedInternalProductID = ""
	results := ValidatePurchase(source, purchase, false)
	for _, result := range results {
		if result.Code == "PRODUCT_MATCH_REQUIRED" {
			t.Fatalf("product match should not block validation: %#v", results)
		}
	}
	if HasBlocking(results) {
		t.Fatalf("expected missing product match to remain valid, got %#v", results)
	}
}

func assertCode(t *testing.T, results []ValidationResult, code string) {
	t.Helper()
	for _, result := range results {
		if result.Code == code && result.Blocking {
			return
		}
	}
	t.Fatalf("expected blocking code %s, got %#v", code, results)
}

func assertWarningCode(t *testing.T, results []ValidationResult, code string) {
	t.Helper()
	for _, result := range results {
		if result.Code == code && result.Severity == ValidationWarning && !result.Blocking {
			return
		}
	}
	t.Fatalf("expected non-blocking warning code %s, got %#v", code, results)
}
