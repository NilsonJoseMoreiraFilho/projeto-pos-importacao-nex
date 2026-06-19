package xlsx

import (
	"context"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"

	"projeto_pos/backend/internal/application"
	"projeto_pos/backend/internal/domain"
)

type Reader struct {
	SheetName string
}

func NewReader() Reader {
	return Reader{}
}

func (r Reader) Read(_ context.Context, path string) (application.RawDocumentExtraction, error) {
	f, err := excelize.OpenFile(path)
	if err != nil {
		return application.RawDocumentExtraction{}, err
	}
	defer f.Close()

	sheet := r.SheetName
	if sheet == "" {
		sheets := f.GetSheetList()
		if len(sheets) == 0 {
			return application.RawDocumentExtraction{}, fmt.Errorf("xlsx has no sheets")
		}
		sheet = sheets[0]
	}
	rows, err := f.GetRows(sheet)
	if err != nil {
		return application.RawDocumentExtraction{}, err
	}
	if len(rows) < 2 {
		return application.RawDocumentExtraction{}, fmt.Errorf("xlsx sheet %q must include header and at least one item", sheet)
	}

	header := indexHeader(rows[0])
	purchase := domain.Purchase{
		ID:                 cell(rows[1], header, "purchase_id"),
		DocumentNumber:     cell(rows[1], header, "document_number"),
		SourceDocumentType: domain.SourceDocumentTypeSpreadsheet,
		PriceTable:         cell(rows[1], header, "price_table"),
		FreightMode:        cell(rows[1], header, "freight_mode"),
		Supplier: domain.Supplier{
			ID:                cell(rows[1], header, "supplier_id"),
			LegalName:         cell(rows[1], header, "supplier_legal_name"),
			TradeName:         cell(rows[1], header, "supplier_trade_name"),
			DocumentNumber:    cell(rows[1], header, "supplier_document_number"),
			StateRegistration: cell(rows[1], header, "supplier_state_registration"),
		},
	}
	if rawDate := cell(rows[1], header, "issue_date"); rawDate != "" {
		if parsed, err := time.Parse("2006-01-02", rawDate); err == nil {
			purchase.IssueDate = parsed
		}
	}

	var productsTotal domain.Money
	for i, row := range rows[1:] {
		if isBlank(row) {
			continue
		}
		packageQuantity, err := parseInt(cell(row, header, "package_quantity"))
		if err != nil && cell(row, header, "package_quantity") != "" {
			return application.RawDocumentExtraction{}, fmt.Errorf("invalid package_quantity on row %d: %w", i+2, err)
		}
		quantity, err := parseDecimal(cell(row, header, "quantity"))
		if err != nil {
			return application.RawDocumentExtraction{}, fmt.Errorf("invalid quantity on row %d: %w", i+2, err)
		}
		unitCost, err := parseMoney(cell(row, header, "unit_cost"))
		if err != nil {
			return application.RawDocumentExtraction{}, fmt.Errorf("invalid unit_cost on row %d: %w", i+2, err)
		}
		totalCost, err := parseMoney(cell(row, header, "total_cost"))
		if err != nil {
			return application.RawDocumentExtraction{}, fmt.Errorf("invalid total_cost on row %d: %w", i+2, err)
		}
		item := domain.PurchaseItem{
			LineNumber:               i + 1,
			SupplierProductCode:      cell(row, header, "supplier_product_code"),
			Barcode:                  cell(row, header, "barcode"),
			Reference:                cell(row, header, "reference"),
			Description:              cell(row, header, "description"),
			Unit:                     cell(row, header, "unit"),
			PackageQuantity:          packageQuantity,
			Quantity:                 quantity,
			UnitCost:                 unitCost,
			TotalCost:                totalCost,
			MatchedInternalProductID: cell(row, header, "matched_internal_product_id"),
		}
		if item.PackageQuantity == 0 {
			item.PackageQuantity = 1
		}
		purchase.Items = append(purchase.Items, item)
		productsTotal = productsTotal.Add(item.TotalCost)
	}

	totals, err := parseTotals(rows[1], header, productsTotal)
	if err != nil {
		return application.RawDocumentExtraction{}, err
	}
	purchase.Totals = totals

	return application.RawDocumentExtraction{
		SourceDocument: domain.SourceDocument{
			ID:                   filepath.Base(path),
			FileName:             filepath.Base(path),
			FileType:             "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
			PageCount:            1,
			CurrentPage:          1,
			StorageLocation:      path,
			QualityScore:         1,
			DetectedDocumentType: domain.SourceDocumentTypeSpreadsheet,
		},
		Purchase:         purchase,
		ExtractorName:    "xlsx-reader",
		ExtractorVersion: "0.1.0",
		ExtractedAt:      time.Now().UTC(),
	}, nil
}

func indexHeader(row []string) map[string]int {
	index := make(map[string]int, len(row))
	for i, value := range row {
		key := strings.ToLower(strings.TrimSpace(value))
		key = strings.ReplaceAll(key, " ", "_")
		index[key] = i
	}
	return index
}

func cell(row []string, header map[string]int, name string) string {
	i, ok := header[name]
	if !ok || i >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[i])
}

func parseDecimal(value string) (float64, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, nil
	}
	value = normalizeDecimal(value)
	return strconv.ParseFloat(value, 64)
}

func parseInt(value string) (int, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, nil
	}
	return strconv.Atoi(value)
}

func parseMoney(value string) (domain.Money, error) {
	parsed, err := parseDecimal(value)
	if err != nil {
		return domain.Money{}, err
	}
	return domain.NewMoneyFromFloat(parsed), nil
}

func moneyOrDefault(value string, fallback domain.Money) (domain.Money, error) {
	if strings.TrimSpace(value) == "" {
		return fallback, nil
	}
	return parseMoney(value)
}

func parseTotals(row []string, header map[string]int, productsTotal domain.Money) (domain.PurchaseTotals, error) {
	parsedProductsTotal, err := moneyOrDefault(cell(row, header, "products_total"), productsTotal)
	if err != nil {
		return domain.PurchaseTotals{}, fmt.Errorf("invalid products_total: %w", err)
	}
	grandTotal, err := moneyOrDefault(cell(row, header, "grand_total"), productsTotal)
	if err != nil {
		return domain.PurchaseTotals{}, fmt.Errorf("invalid grand_total: %w", err)
	}
	discount, err := parseMoney(cell(row, header, "discount"))
	if err != nil {
		return domain.PurchaseTotals{}, fmt.Errorf("invalid discount: %w", err)
	}
	addition, err := parseMoney(cell(row, header, "addition"))
	if err != nil {
		return domain.PurchaseTotals{}, fmt.Errorf("invalid addition: %w", err)
	}
	ipi, err := parseMoney(cell(row, header, "ipi"))
	if err != nil {
		return domain.PurchaseTotals{}, fmt.Errorf("invalid ipi: %w", err)
	}
	taxSubstitution, err := parseMoney(cell(row, header, "tax_substitution"))
	if err != nil {
		return domain.PurchaseTotals{}, fmt.Errorf("invalid tax_substitution: %w", err)
	}
	fcpST, err := parseMoney(cell(row, header, "fcp_st"))
	if err != nil {
		return domain.PurchaseTotals{}, fmt.Errorf("invalid fcp_st: %w", err)
	}
	return domain.PurchaseTotals{
		ProductsTotal:   parsedProductsTotal,
		Discount:        discount,
		Addition:        addition,
		IPI:             ipi,
		TaxSubstitution: taxSubstitution,
		FCPST:           fcpST,
		GrandTotal:      grandTotal,
	}, nil
}

func normalizeDecimal(value string) string {
	value = strings.ReplaceAll(value, " ", "")
	value = strings.ReplaceAll(value, "R$", "")
	if strings.Contains(value, ",") {
		value = strings.ReplaceAll(value, ".", "")
		value = strings.ReplaceAll(value, ",", ".")
	}
	return value
}

func isBlank(row []string) bool {
	for _, value := range row {
		if strings.TrimSpace(value) != "" {
			return false
		}
	}
	return true
}
