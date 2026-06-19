package jsonexport

import (
	"context"
	"encoding/json"
	"os"

	"projeto_pos/backend/internal/application"
)

type Exporter struct{}

func NewExporter() Exporter {
	return Exporter{}
}

func (e Exporter) Export(_ context.Context, proposal application.ImportProposal, path string) error {
	data, err := json.MarshalIndent(proposal, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
