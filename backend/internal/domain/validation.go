package domain

import "fmt"

type ValidationSeverity string

const (
	ValidationInfo    ValidationSeverity = "INFO"
	ValidationWarning ValidationSeverity = "WARNING"
	ValidationError   ValidationSeverity = "ERROR"
)

type ValidationResult struct {
	Severity ValidationSeverity `json:"severity"`
	Code     string             `json:"code"`
	Field    string             `json:"field"`
	Message  string             `json:"message"`
	Blocking bool               `json:"blocking"`
}

func ValidatePurchase(source SourceDocument, purchase Purchase, requiresHumanReview bool) []ValidationResult {
	var results []ValidationResult

	if source.PageCount > 0 && source.CurrentPage > 0 && source.CurrentPage < source.PageCount {
		results = append(results, errorResult("DOCUMENT_INCOMPLETE", "sourceDocument.pageCount", "documento possui páginas faltantes"))
	}
	if purchase.Supplier.LegalName == "" && purchase.Supplier.DocumentNumber == "" {
		results = append(results, errorResult("SUPPLIER_NOT_IDENTIFIED", "supplier", "fornecedor não identificado"))
	}
	if len(purchase.Items) == 0 {
		results = append(results, errorResult("PURCHASE_WITHOUT_ITEMS", "items", "compra não possui itens"))
	}
	if requiresHumanReview {
		results = append(results, errorResult("HUMAN_REVIEW_REQUIRED", "review", "origem por imagem/OCR exige revisão humana antes da aprovação"))
	}

	var sum Money
	for i, item := range purchase.Items {
		field := fmt.Sprintf("items[%d]", i)
		if item.Quantity <= 0 {
			results = append(results, errorResult("INVALID_QUANTITY", field+".quantity", "quantidade deve ser maior que zero"))
		}
		if item.UnitCost.Cents < 0 {
			results = append(results, errorResult("INVALID_UNIT_COST", field+".unitCost", "valor unitário não pode ser negativo"))
		}
		expectedTotal := NewMoneyFromFloat(item.Quantity * item.UnitCost.Float64())
		if !item.TotalCost.WithinCents(expectedTotal, 1) {
			results = append(results, errorResult("ITEM_TOTAL_MISMATCH", field+".totalCost", "total do item diverge de quantidade x valor unitário"))
		}
		sum = sum.Add(item.TotalCost)
	}

	if !purchase.Totals.ProductsTotal.WithinCents(sum, 1) {
		results = append(results, errorResult("PRODUCTS_TOTAL_MISMATCH", "totals.productsTotal", "soma dos itens diverge do total de produtos"))
	}
	expectedGrandTotal := purchase.Totals.ProductsTotal.
		Add(purchase.Totals.Addition).
		Add(purchase.Totals.IPI).
		Add(purchase.Totals.TaxSubstitution).
		Add(purchase.Totals.FCPST).
		Sub(purchase.Totals.Discount)
	if !purchase.Totals.GrandTotal.WithinCents(expectedGrandTotal, 1) {
		results = append(results, errorResult("GRAND_TOTAL_MISMATCH", "totals.grandTotal", "total geral diverge dos totais calculados"))
	}

	return results
}

func HasBlocking(results []ValidationResult) bool {
	for _, result := range results {
		if result.Blocking {
			return true
		}
	}
	return false
}

func errorResult(code, field, message string) ValidationResult {
	return ValidationResult{
		Severity: ValidationError,
		Code:     code,
		Field:    field,
		Message:  message,
		Blocking: true,
	}
}
