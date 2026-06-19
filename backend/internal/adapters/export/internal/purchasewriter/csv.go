package purchasewriter

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"

	"projeto_pos/backend/internal/domain"
)

func WritePurchaseItemsCSV(writer io.Writer, purchase domain.Purchase) error {
	csvWriter := csv.NewWriter(writer)
	defer csvWriter.Flush()

	if err := csvWriter.Write([]string{
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

	for _, item := range purchase.Items {
		if err := csvWriter.Write([]string{
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

	return csvWriter.Error()
}
