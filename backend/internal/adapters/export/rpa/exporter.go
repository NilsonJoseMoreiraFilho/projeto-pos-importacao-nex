package rpa

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"time"

	"projeto_pos/backend/internal/application"
	"projeto_pos/backend/internal/domain"
)

type Exporter struct {
	Command string
	Args    []string
}

func NewExporter(command string, args ...string) Exporter {
	return Exporter{Command: command, Args: args}
}

func (e Exporter) ExportApprovedPurchase(ctx context.Context, purchase domain.ApprovedPurchase) (application.ExportResult, error) {
	if e.Command == "" {
		return application.ExportResult{}, fmt.Errorf("rpa command is required")
	}
	payload, err := os.CreateTemp("", "approved-purchase-*.json")
	if err != nil {
		return application.ExportResult{}, err
	}
	defer os.Remove(payload.Name())
	if err := json.NewEncoder(payload).Encode(purchase); err != nil {
		payload.Close()
		return application.ExportResult{}, err
	}
	if err := payload.Close(); err != nil {
		return application.ExportResult{}, err
	}

	args := append([]string{}, e.Args...)
	args = append(args, payload.Name())
	cmd := exec.CommandContext(ctx, e.Command, args...)
	if output, err := cmd.CombinedOutput(); err != nil {
		return application.ExportResult{}, fmt.Errorf("rpa command failed: %w: %s", err, string(output))
	}
	return application.ExportResult{
		Status:      application.ExportStatusSuccess,
		Destination: "rpa",
		Reference:   e.Command,
		CreatedAt:   time.Now().UTC(),
	}, nil
}
