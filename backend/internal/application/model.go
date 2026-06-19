package application

import (
	"time"

	"projeto_pos/backend/internal/domain"
)

type ExtractedFieldStatus string

const (
	FieldExtracted     ExtractedFieldStatus = "EXTRACTED"
	FieldLowConfidence ExtractedFieldStatus = "LOW_CONFIDENCE"
	FieldReviewed      ExtractedFieldStatus = "REVIEWED"
	FieldCorrected     ExtractedFieldStatus = "CORRECTED"
	FieldRejected      ExtractedFieldStatus = "REJECTED"
)

type ExtractedField struct {
	FieldPath       string               `json:"fieldPath"`
	RawText         string               `json:"rawText"`
	NormalizedValue any                  `json:"normalizedValue"`
	Confidence      float64              `json:"confidence"`
	Page            int                  `json:"page"`
	Status          ExtractedFieldStatus `json:"status"`
}

type RawDocumentExtraction struct {
	SourceDocument   domain.SourceDocument `json:"sourceDocument"`
	Purchase         domain.Purchase       `json:"purchase"`
	ExtractedFields  []ExtractedField      `json:"extractedFields"`
	ExtractorName    string                `json:"extractorName"`
	ExtractorVersion string                `json:"extractorVersion"`
	ExtractedAt      time.Time             `json:"extractedAt"`
}

type ImportProposalStatus string

const (
	ProposalProposed    ImportProposalStatus = "PROPOSED"
	ProposalNeedsReview ImportProposalStatus = "NEEDS_REVIEW"
)

type ImportProposal struct {
	ID                string                    `json:"id"`
	SourceDocument    domain.SourceDocument     `json:"sourceDocument"`
	ExtractedFields   []ExtractedField          `json:"extractedFields"`
	PurchaseDraft     domain.Purchase           `json:"purchaseDraft"`
	ValidationResults []domain.ValidationResult `json:"validationResults"`
	Status            ImportProposalStatus      `json:"status"`
	CreatedAt         time.Time                 `json:"createdAt"`
}
