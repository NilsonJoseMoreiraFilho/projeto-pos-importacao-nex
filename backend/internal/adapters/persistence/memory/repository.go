package memory

import (
	"context"
	"fmt"
	"sync"

	"projeto_pos/backend/internal/application"
	"projeto_pos/backend/internal/domain"
)

type Repository struct {
	mu        sync.RWMutex
	proposals map[string]application.ImportProposal
	approved  map[string]domain.ApprovedPurchase
	exports   map[string]application.ExportResult
}

func NewRepository() *Repository {
	return &Repository{
		proposals: make(map[string]application.ImportProposal),
		approved:  make(map[string]domain.ApprovedPurchase),
		exports:   make(map[string]application.ExportResult),
	}
}

func (r *Repository) SaveProposal(_ context.Context, proposal application.ImportProposal) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.proposals[proposal.ID] = proposal
	return nil
}

func (r *Repository) FindProposalByID(_ context.Context, id string) (application.ImportProposal, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	proposal, ok := r.proposals[id]
	if !ok {
		return application.ImportProposal{}, fmt.Errorf("proposal not found: %s", id)
	}
	return proposal, nil
}

func (r *Repository) SaveApprovedPurchase(_ context.Context, proposalID string, purchase domain.ApprovedPurchase) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.approved[proposalID] = purchase
	return nil
}

func (r *Repository) FindApprovedPurchaseByProposalID(_ context.Context, proposalID string) (domain.ApprovedPurchase, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	purchase, ok := r.approved[proposalID]
	if !ok {
		return domain.ApprovedPurchase{}, fmt.Errorf("approved purchase not found for proposal: %s", proposalID)
	}
	return purchase, nil
}

func (r *Repository) SaveExportResult(_ context.Context, proposalID string, result application.ExportResult) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.exports[proposalID] = result
	return nil
}
