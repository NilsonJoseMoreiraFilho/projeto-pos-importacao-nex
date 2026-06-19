package application

import "time"

type ReviewDecisionType string

const (
	ReviewDecisionAccept  ReviewDecisionType = "ACCEPT"
	ReviewDecisionCorrect ReviewDecisionType = "CORRECT"
	ReviewDecisionReject  ReviewDecisionType = "REJECT"
)

type ReviewDecision struct {
	FieldPath      string             `json:"fieldPath"`
	Decision       ReviewDecisionType `json:"decision"`
	CorrectedValue any                `json:"correctedValue,omitempty"`
	ReviewedBy     string             `json:"reviewedBy"`
	ReviewedAt     time.Time          `json:"reviewedAt"`
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
