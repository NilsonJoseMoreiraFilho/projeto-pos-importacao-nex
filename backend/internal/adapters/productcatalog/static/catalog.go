package static

import (
	"context"
	"strings"

	"projeto_pos/backend/internal/application"
	"projeto_pos/backend/internal/domain"
)

type Product struct {
	ID                  string
	Barcode             string
	SupplierProductCode string
	Reference           string
	Description         string
}

type Catalog struct {
	products []Product
}

func NewCatalog(products []Product) Catalog {
	return Catalog{products: products}
}

func (c Catalog) MatchProduct(_ context.Context, item domain.PurchaseItem) (application.ProductMatchResult, error) {
	for _, product := range c.products {
		if product.Barcode != "" && product.Barcode == item.Barcode {
			return matched(product.ID, 1, "barcode"), nil
		}
		if product.SupplierProductCode != "" && product.SupplierProductCode == item.SupplierProductCode {
			return matched(product.ID, 0.95, "supplier_product_code"), nil
		}
		if product.Reference != "" && product.Reference == item.Reference {
			return matched(product.ID, 0.85, "reference"), nil
		}
		if product.Description != "" && strings.EqualFold(product.Description, item.Description) {
			return matched(product.ID, 0.75, "description"), nil
		}
	}
	return application.ProductMatchResult{Matched: false, Reason: "no_match"}, nil
}

func matched(id string, confidence float64, reason string) application.ProductMatchResult {
	return application.ProductMatchResult{
		Matched:           true,
		InternalProductID: id,
		Confidence:        confidence,
		Reason:            reason,
	}
}
