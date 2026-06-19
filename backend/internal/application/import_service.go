package application

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
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

type PurchaseRepository interface {
	SaveProposal(ctx context.Context, proposal ImportProposal) error
	FindProposalByID(ctx context.Context, id string) (ImportProposal, error)
	SaveApprovedPurchase(ctx context.Context, proposalID string, purchase domain.ApprovedPurchase) error
	FindApprovedPurchaseByProposalID(ctx context.Context, proposalID string) (domain.ApprovedPurchase, error)
	SaveExportResult(ctx context.Context, proposalID string, result ExportResult) error
}

type ProductCatalog interface {
	MatchProduct(ctx context.Context, item domain.PurchaseItem) (ProductMatchResult, error)
}

type ManagementSystemExporter interface {
	ExportApprovedPurchase(ctx context.Context, purchase domain.ApprovedPurchase) (ExportResult, error)
}

type ImportPurchaseService struct {
	reader     DocumentReader
	repository PurchaseRepository
	catalog    ProductCatalog
}

func NewImportPurchaseService(reader DocumentReader) ImportPurchaseService {
	return ImportPurchaseService{reader: reader}
}

func NewImportPurchaseServiceWithDependencies(reader DocumentReader, repository PurchaseRepository, catalog ProductCatalog) ImportPurchaseService {
	return ImportPurchaseService{reader: reader, repository: repository, catalog: catalog}
}

func (s ImportPurchaseService) Import(ctx context.Context, input ImportPurchaseInput) (ImportProposal, error) {
	raw, err := s.reader.Read(ctx, input.Path)
	if err != nil {
		return ImportProposal{}, err
	}

	if s.catalog != nil {
		for i, item := range raw.Purchase.Items {
			if item.MatchedInternalProductID != "" {
				continue
			}
			match, err := s.catalog.MatchProduct(ctx, item)
			if err != nil {
				return ImportProposal{}, err
			}
			if match.Matched {
				raw.Purchase.Items[i].MatchedInternalProductID = match.InternalProductID
			}
		}
	}

	results := domain.ValidatePurchase(raw.SourceDocument, raw.Purchase, input.RequiresHumanReview)
	status := ProposalProposed
	if domain.HasBlocking(results) {
		status = ProposalNeedsReview
	}

	proposal := ImportProposal{
		ID:                newProposalID(),
		SourceDocument:    raw.SourceDocument,
		ExtractedFields:   raw.ExtractedFields,
		PurchaseDraft:     raw.Purchase,
		ValidationResults: results,
		Status:            status,
		CreatedAt:         time.Now().UTC(),
	}

	if s.repository != nil {
		if err := s.repository.SaveProposal(ctx, proposal); err != nil {
			return ImportProposal{}, err
		}
	}

	return proposal, nil
}

func newProposalID() string {
	var bytes [8]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return fmt.Sprintf("proposal-%d", time.Now().UnixNano())
	}
	return "proposal-" + hex.EncodeToString(bytes[:])
}

type ReviewImportProposalService struct {
	repository PurchaseRepository
}

func NewReviewImportProposalService(repository PurchaseRepository) ReviewImportProposalService {
	return ReviewImportProposalService{repository: repository}
}

func (s ReviewImportProposalService) Review(ctx context.Context, proposalID string, decisions []ReviewDecision) (ImportProposal, error) {
	if s.repository == nil {
		return ImportProposal{}, errors.New("review service requires repository")
	}
	proposal, err := s.repository.FindProposalByID(ctx, proposalID)
	if err != nil {
		return ImportProposal{}, err
	}
	for _, decision := range decisions {
		for i, field := range proposal.ExtractedFields {
			if field.FieldPath != decision.FieldPath {
				continue
			}
			switch decision.Decision {
			case ReviewDecisionAccept:
				proposal.ExtractedFields[i].Status = FieldReviewed
			case ReviewDecisionCorrect:
				proposal.ExtractedFields[i].Status = FieldCorrected
				proposal.ExtractedFields[i].NormalizedValue = decision.CorrectedValue
				if err := applyCorrectedValue(&proposal.PurchaseDraft, decision.FieldPath, decision.CorrectedValue); err != nil {
					return ImportProposal{}, err
				}
			case ReviewDecisionReject:
				proposal.ExtractedFields[i].Status = FieldRejected
			}
		}
	}
	proposal.ValidationResults = domain.ValidatePurchase(proposal.SourceDocument, proposal.PurchaseDraft, false)
	proposal.ValidationResults = append(proposal.ValidationResults, validateReviewedFields(proposal.ExtractedFields)...)
	proposal.Status = ProposalProposed
	if domain.HasBlocking(proposal.ValidationResults) {
		proposal.Status = ProposalNeedsReview
	}
	if err := s.repository.SaveProposal(ctx, proposal); err != nil {
		return ImportProposal{}, err
	}
	return proposal, nil
}

type ApprovePurchaseService struct {
	repository PurchaseRepository
}

func NewApprovePurchaseService(repository PurchaseRepository) ApprovePurchaseService {
	return ApprovePurchaseService{repository: repository}
}

func (s ApprovePurchaseService) Approve(ctx context.Context, proposalID string, reviewer string) (domain.ApprovedPurchase, error) {
	if s.repository == nil {
		return domain.ApprovedPurchase{}, errors.New("approve service requires repository")
	}
	proposal, err := s.repository.FindProposalByID(ctx, proposalID)
	if err != nil {
		return domain.ApprovedPurchase{}, err
	}
	if domain.HasBlocking(proposal.ValidationResults) {
		return domain.ApprovedPurchase{}, errors.New("proposal has blocking validation results")
	}
	if hasRejectedField(proposal.ExtractedFields) {
		return domain.ApprovedPurchase{}, errors.New("proposal has rejected extracted fields")
	}
	approved := domain.ApprovedPurchase{
		Purchase:   proposal.PurchaseDraft,
		ApprovedBy: reviewer,
		ApprovedAt: time.Now().UTC(),
	}
	if err := s.repository.SaveApprovedPurchase(ctx, proposalID, approved); err != nil {
		return domain.ApprovedPurchase{}, err
	}
	return approved, nil
}

func validateReviewedFields(fields []ExtractedField) []domain.ValidationResult {
	var results []domain.ValidationResult
	for _, field := range fields {
		if field.Status == FieldRejected {
			results = append(results, domain.ValidationResult{
				Severity: domain.ValidationError,
				Code:     "FIELD_REJECTED",
				Field:    field.FieldPath,
				Message:  "campo extraido foi rejeitado na revisao humana",
				Blocking: true,
			})
		}
	}
	return results
}

func hasRejectedField(fields []ExtractedField) bool {
	for _, field := range fields {
		if field.Status == FieldRejected {
			return true
		}
	}
	return false
}

func applyCorrectedValue(purchase *domain.Purchase, fieldPath string, value any) error {
	switch fieldPath {
	case "purchase.documentNumber":
		purchase.DocumentNumber = fmt.Sprint(value)
	case "purchase.supplier.legalName":
		purchase.Supplier.LegalName = fmt.Sprint(value)
	case "purchase.supplier.documentNumber":
		purchase.Supplier.DocumentNumber = fmt.Sprint(value)
	default:
		return nil
	}
	return nil
}

type ExportApprovedPurchaseService struct {
	repository PurchaseRepository
	exporter   ManagementSystemExporter
}

func NewExportApprovedPurchaseService(repository PurchaseRepository, exporter ManagementSystemExporter) ExportApprovedPurchaseService {
	return ExportApprovedPurchaseService{repository: repository, exporter: exporter}
}

func (s ExportApprovedPurchaseService) Export(ctx context.Context, proposalID string) (ExportResult, error) {
	if s.repository == nil || s.exporter == nil {
		return ExportResult{}, errors.New("export service requires repository and exporter")
	}
	approved, err := s.repository.FindApprovedPurchaseByProposalID(ctx, proposalID)
	if err != nil {
		return ExportResult{}, err
	}
	result, err := s.exporter.ExportApprovedPurchase(ctx, approved)
	if err != nil {
		return ExportResult{}, err
	}
	if err := s.repository.SaveExportResult(ctx, proposalID, result); err != nil {
		return ExportResult{}, err
	}
	return result, nil
}
