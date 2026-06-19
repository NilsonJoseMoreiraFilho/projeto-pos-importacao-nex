package application

import (
	"context"
	"fmt"
	"testing"
	"time"

	"projeto_pos/backend/internal/domain"
)

func TestImportServiceSavesProposal(t *testing.T) {
	repository := newFakeRepository()
	service := NewImportPurchaseServiceWithDependencies(stubReader{raw: validRawExtraction()}, repository, nil)

	proposal, err := service.Import(context.Background(), ImportPurchaseInput{})
	if err != nil {
		t.Fatalf("import failed: %v", err)
	}

	saved, err := repository.FindProposalByID(context.Background(), proposal.ID)
	if err != nil {
		t.Fatalf("proposal was not saved: %v", err)
	}
	if saved.PurchaseDraft.DocumentNumber != "123" {
		t.Fatalf("unexpected saved proposal: %#v", saved.PurchaseDraft)
	}
}

func TestApproveServiceBlocksInvalidProposal(t *testing.T) {
	repository := newFakeRepository()
	service := NewImportPurchaseServiceWithDependencies(stubReader{raw: invalidRawExtraction()}, repository, nil)
	proposal, err := service.Import(context.Background(), ImportPurchaseInput{})
	if err != nil {
		t.Fatalf("import failed: %v", err)
	}

	_, err = NewApprovePurchaseService(repository).Approve(context.Background(), proposal.ID, "tester")
	if err == nil {
		t.Fatal("expected approval to fail for proposal with blocking validation")
	}
}

func TestApproveServiceApprovesValidProposal(t *testing.T) {
	repository := newFakeRepository()
	service := NewImportPurchaseServiceWithDependencies(stubReader{raw: validRawExtraction()}, repository, nil)
	proposal, err := service.Import(context.Background(), ImportPurchaseInput{})
	if err != nil {
		t.Fatalf("import failed: %v", err)
	}

	approved, err := NewApprovePurchaseService(repository).Approve(context.Background(), proposal.ID, "tester")
	if err != nil {
		t.Fatalf("approval failed: %v", err)
	}
	if approved.ApprovedBy != "tester" {
		t.Fatalf("unexpected approver: %s", approved.ApprovedBy)
	}
}

type stubReader struct {
	raw RawDocumentExtraction
}

func (s stubReader) Read(context.Context, string) (RawDocumentExtraction, error) {
	return s.raw, nil
}

func validRawExtraction() RawDocumentExtraction {
	return RawDocumentExtraction{
		SourceDocument: domain.SourceDocument{
			ID:                   "doc-1",
			PageCount:            1,
			CurrentPage:          1,
			DetectedDocumentType: domain.SourceDocumentTypeManualJSON,
		},
		Purchase: domain.Purchase{
			ID:             "purchase-1",
			Supplier:       domain.Supplier{LegalName: "Fornecedor Teste"},
			DocumentNumber: "123",
			IssueDate:      time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC),
			Items: []domain.PurchaseItem{
				{
					LineNumber:               1,
					Description:              "Produto A",
					Unit:                     "UNID",
					Quantity:                 2,
					UnitCost:                 domain.NewMoneyFromFloat(10),
					TotalCost:                domain.NewMoneyFromFloat(20),
					MatchedInternalProductID: "prod-a",
				},
			},
			Totals: domain.PurchaseTotals{
				ProductsTotal: domain.NewMoneyFromFloat(20),
				GrandTotal:    domain.NewMoneyFromFloat(20),
			},
		},
	}
}

func invalidRawExtraction() RawDocumentExtraction {
	raw := validRawExtraction()
	raw.Purchase.Supplier = domain.Supplier{}
	return raw
}

type fakeRepository struct {
	proposals map[string]ImportProposal
	approved  map[string]domain.ApprovedPurchase
	exports   map[string]ExportResult
}

func newFakeRepository() *fakeRepository {
	return &fakeRepository{
		proposals: map[string]ImportProposal{},
		approved:  map[string]domain.ApprovedPurchase{},
		exports:   map[string]ExportResult{},
	}
}

func (r *fakeRepository) SaveProposal(_ context.Context, proposal ImportProposal) error {
	r.proposals[proposal.ID] = proposal
	return nil
}

func (r *fakeRepository) FindProposalByID(_ context.Context, id string) (ImportProposal, error) {
	proposal, ok := r.proposals[id]
	if !ok {
		return ImportProposal{}, fmt.Errorf("proposal not found: %s", id)
	}
	return proposal, nil
}

func (r *fakeRepository) SaveApprovedPurchase(_ context.Context, proposalID string, purchase domain.ApprovedPurchase) error {
	r.approved[proposalID] = purchase
	return nil
}

func (r *fakeRepository) FindApprovedPurchaseByProposalID(_ context.Context, proposalID string) (domain.ApprovedPurchase, error) {
	purchase, ok := r.approved[proposalID]
	if !ok {
		return domain.ApprovedPurchase{}, fmt.Errorf("approved purchase not found: %s", proposalID)
	}
	return purchase, nil
}

func (r *fakeRepository) SaveExportResult(_ context.Context, proposalID string, result ExportResult) error {
	r.exports[proposalID] = result
	return nil
}
