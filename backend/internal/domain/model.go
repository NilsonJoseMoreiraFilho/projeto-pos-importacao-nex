package domain

import "time"

type SourceDocumentType string

const (
	SourceDocumentTypePhoto       SourceDocumentType = "PHOTO"
	SourceDocumentTypeManualJSON  SourceDocumentType = "MANUAL_JSON"
	SourceDocumentTypeNFeXML      SourceDocumentType = "NFE_XML"
	SourceDocumentTypeSpreadsheet SourceDocumentType = "SPREADSHEET"
	SourceDocumentTypePDF         SourceDocumentType = "PDF"
	SourceDocumentTypeUnknown     SourceDocumentType = "UNKNOWN"
)

type Supplier struct {
	ID                string `json:"id"`
	LegalName         string `json:"legalName"`
	TradeName         string `json:"tradeName"`
	DocumentNumber    string `json:"documentNumber"`
	StateRegistration string `json:"stateRegistration"`
}

type PurchaseItem struct {
	LineNumber               int     `json:"lineNumber"`
	SupplierProductCode      string  `json:"supplierProductCode"`
	Barcode                  string  `json:"barcode"`
	Reference                string  `json:"reference"`
	Description              string  `json:"description"`
	Unit                     string  `json:"unit"`
	PackageQuantity          int     `json:"packageQuantity"`
	Quantity                 float64 `json:"quantity"`
	UnitCost                 Money   `json:"unitCost"`
	TotalCost                Money   `json:"totalCost"`
	MatchedInternalProductID string  `json:"matchedInternalProductId"`
}

type PurchaseTotals struct {
	ProductsTotal   Money `json:"productsTotal"`
	Discount        Money `json:"discount"`
	Addition        Money `json:"addition"`
	IPI             Money `json:"ipi"`
	TaxSubstitution Money `json:"taxSubstitution"`
	FCPST           Money `json:"fcpSt"`
	GrandTotal      Money `json:"grandTotal"`
}

type Purchase struct {
	ID                 string             `json:"id"`
	Supplier           Supplier           `json:"supplier"`
	DocumentNumber     string             `json:"documentNumber"`
	IssueDate          time.Time          `json:"issueDate"`
	SourceDocumentType SourceDocumentType `json:"sourceDocumentType"`
	PriceTable         string             `json:"priceTable"`
	FreightMode        string             `json:"freightMode"`
	Items              []PurchaseItem     `json:"items"`
	Totals             PurchaseTotals     `json:"totals"`
}

type ApprovedPurchase struct {
	Purchase      Purchase  `json:"purchase"`
	ApprovedBy    string    `json:"approvedBy"`
	ApprovedAt    time.Time `json:"approvedAt"`
	ApprovalNotes string    `json:"approvalNotes"`
}

type SourceDocument struct {
	ID                   string             `json:"id"`
	FileName             string             `json:"fileName"`
	FileType             string             `json:"fileType"`
	PageCount            int                `json:"pageCount"`
	CurrentPage          int                `json:"currentPage"`
	Checksum             string             `json:"checksum"`
	StorageLocation      string             `json:"storageLocation"`
	QualityScore         float64            `json:"qualityScore"`
	DetectedDocumentType SourceDocumentType `json:"detectedDocumentType"`
	Observations         []string           `json:"observations"`
}
