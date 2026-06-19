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
		item := domain.PurchaseItem{
			LineNumber:               i + 1,
			SupplierProductCode:      cell(row, header, "supplier_product_code"),
			Barcode:                  cell(row, header, "barcode"),
			Reference:                cell(row, header, "reference"),
			Description:              cell(row, header, "description"),
			Unit:                     cell(row, header, "unit"),
			PackageQuantity:          parseInt(cell(row, header, "package_quantity")),
			Quantity:                 parseFloat(cell(row, header, "quantity")),
			UnitCost:                 domain.NewMoneyFromFloat(parseFloat(cell(row, header, "unit_cost"))),
			TotalCost:                domain.NewMoneyFromFloat(parseFloat(cell(row, header, "total_cost"))),
			MatchedInternalProductID: cell(row, header, "matched_internal_product_id"),
		}
		if item.PackageQuantity == 0 {
			item.PackageQuantity = 1
		}
		purchase.Items = append(purchase.Items, item)
		productsTotal = productsTotal.Add(item.TotalCost)
	}

	purchase.Totals = domain.PurchaseTotals{
		ProductsTotal:   moneyOrDefault(cell(rows[1], header, "products_total"), productsTotal),
		Discount:        domain.NewMoneyFromFloat(parseFloat(cell(rows[1], header, "discount"))),
		Addition:        domain.NewMoneyFromFloat(parseFloat(cell(rows[1], header, "addition"))),
		IPI:             domain.NewMoneyFromFloat(parseFloat(cell(rows[1], header, "ipi"))),
		TaxSubstitution: domain.NewMoneyFromFloat(parseFloat(cell(rows[1], header, "tax_substitution"))),
		FCPST:           domain.NewMoneyFromFloat(parseFloat(cell(rows[1], header, "fcp_st"))),
		GrandTotal:      moneyOrDefault(cell(rows[1], header, "grand_total"), productsTotal),
	}

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

func parseFloat(value string) float64 {
	value = strings.ReplaceAll(strings.TrimSpace(value), ",", ".")
	parsed, _ := strconv.ParseFloat(value, 64)
	return parsed
}

func parseInt(value string) int {
	parsed, _ := strconv.Atoi(strings.TrimSpace(value))
	return parsed
}

func moneyOrDefault(value string, fallback domain.Money) domain.Money {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return domain.NewMoneyFromFloat(parseFloat(value))
}

func isBlank(row []string) bool {
	for _, value := range row {
		if strings.TrimSpace(value) != "" {
			return false
		}
	}
	return true
}
