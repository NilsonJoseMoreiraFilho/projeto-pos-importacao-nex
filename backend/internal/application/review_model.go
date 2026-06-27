package application

import "time"

type ReviewEdit struct {
	FieldPath string    `json:"fieldPath"`
	Value     any       `json:"value"`
	EditedBy  string    `json:"editedBy"`
	EditedAt  time.Time `json:"editedAt"`
}

type ExportStatus string

const (
	ExportStatusSuccess ExportStatus = "SUCCESS"
	ExportStatusFailed  ExportStatus = "FAILED"
)

type ExportResult struct {
	Status      ExportStatus `json:"status"`
	Destination string       `json:"destination"`
	Reference   string       `json:"reference"`
	CreatedAt   time.Time    `json:"createdAt"`
}
