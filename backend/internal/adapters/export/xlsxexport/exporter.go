package xlsxexport

import (
	"context"
	"strconv"
	"time"

	"github.com/xuri/excelize/v2"

	"projeto_pos/backend/internal/application"
	"projeto_pos/backend/internal/domain"
)

type Exporter struct {
	Path string
}

func NewExporter(path string) Exporter {
	return Exporter{Path: path}
}

func (e Exporter) ExportApprovedPurchase(_ context.Context, purchase domain.ApprovedPurchase) (application.ExportResult, error) {
	file := excelize.NewFile()
	defer file.Close()

	sheet := "purchase_items"
	file.SetSheetName("Sheet1", sheet)
	headers := []string{
		"document_number",
		"supplier",
		"line",
		"barcode",
		"reference",
		"description",
		"unit",
		"quantity",
		"unit_cost",
		"total_cost",
		"matched_internal_product_id",
	}
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		file.SetCellValue(sheet, cell, header)
	}

	for row, item := range purchase.Purchase.Items {
		values := []any{
			purchase.Purchase.DocumentNumber,
			purchase.Purchase.Supplier.LegalName,
			item.LineNumber,
			item.Barcode,
			item.Reference,
			item.Description,
			item.Unit,
			item.Quantity,
			item.UnitCost.Float64(),
			item.TotalCost.Float64(),
			item.MatchedInternalProductID,
		}
		for col, value := range values {
			cell, _ := excelize.CoordinatesToCellName(col+1, row+2)
			file.SetCellValue(sheet, cell, value)
		}
	}

	totalRow := len(purchase.Purchase.Items) + 3
	file.SetCellValue(sheet, "I"+strconv.Itoa(totalRow), "products_total")
	file.SetCellValue(sheet, "J"+strconv.Itoa(totalRow), purchase.Purchase.Totals.ProductsTotal.Float64())
	file.SetCellValue(sheet, "I"+strconv.Itoa(totalRow+1), "grand_total")
	file.SetCellValue(sheet, "J"+strconv.Itoa(totalRow+1), purchase.Purchase.Totals.GrandTotal.Float64())

	if err := file.SaveAs(e.Path); err != nil {
		return application.ExportResult{}, err
	}
	return application.ExportResult{
		Status:      application.ExportStatusSuccess,
		Destination: "xlsx",
		Reference:   e.Path,
		CreatedAt:   time.Now().UTC(),
	}, nil
}
