package application

import (
	"context"
	"fmt"
	"time"

	"projeto_pos/backend/internal/domain"
)

type ImportPurchaseInput struct {
	Path                string
	RequiresHumanReview bool
}

type DocumentReader interface {
	Read(ctx context.Context, path string) (RawDocumentExtraction, error)
}

type ImportPurchaseService struct {
	reader DocumentReader
}

func NewImportPurchaseService(reader DocumentReader) ImportPurchaseService {
	return ImportPurchaseService{reader: reader}
}

func (s ImportPurchaseService) Import(ctx context.Context, input ImportPurchaseInput) (ImportProposal, error) {
	raw, err := s.reader.Read(ctx, input.Path)
	if err != nil {
		return ImportProposal{}, err
	}

	results := domain.ValidatePurchase(raw.SourceDocument, raw.Purchase, input.RequiresHumanReview)
	status := ProposalProposed
	if domain.HasBlocking(results) {
		status = ProposalNeedsReview
	}

	return ImportProposal{
		ID:                fmt.Sprintf("proposal-%d", time.Now().Unix()),
		SourceDocument:    raw.SourceDocument,
		ExtractedFields:   raw.ExtractedFields,
		PurchaseDraft:     raw.Purchase,
		ValidationResults: results,
		Status:            status,
		CreatedAt:         time.Now().UTC(),
	}, nil
}
