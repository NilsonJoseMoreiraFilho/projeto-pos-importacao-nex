package csvexport

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"strconv"

	"projeto_pos/backend/internal/application"
)

type Exporter struct{}

func NewExporter() Exporter {
	return Exporter{}
}

func (e Exporter) Export(_ context.Context, proposal application.ImportProposal, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	if err := writer.Write([]string{
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
	}); err != nil {
		return err
	}

	purchase := proposal.PurchaseDraft
	for _, item := range purchase.Items {
		if err := writer.Write([]string{
			purchase.DocumentNumber,
			purchase.Supplier.LegalName,
			strconv.Itoa(item.LineNumber),
			item.Barcode,
			item.Reference,
			item.Description,
			item.Unit,
			fmt.Sprintf("%.2f", item.Quantity),
			fmt.Sprintf("%.2f", item.UnitCost.Float64()),
			fmt.Sprintf("%.2f", item.TotalCost.Float64()),
			item.MatchedInternalProductID,
		}); err != nil {
			return err
		}
	}

	return writer.Error()
}
