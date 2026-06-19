package nexapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"projeto_pos/backend/internal/application"
	"projeto_pos/backend/internal/domain"
)

type Exporter struct {
	Endpoint   string
	Token      string
	HTTPClient *http.Client
}

func NewExporter(endpoint string, token string) Exporter {
	return Exporter{Endpoint: endpoint, Token: token, HTTPClient: http.DefaultClient}
}

func (e Exporter) ExportApprovedPurchase(ctx context.Context, purchase domain.ApprovedPurchase) (application.ExportResult, error) {
	if e.Endpoint == "" {
		return application.ExportResult{}, fmt.Errorf("nex api endpoint is required")
	}
	body, err := json.Marshal(purchase)
	if err != nil {
		return application.ExportResult{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, e.Endpoint, bytes.NewReader(body))
	if err != nil {
		return application.ExportResult{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	if e.Token != "" {
		req.Header.Set("Authorization", "Bearer "+e.Token)
	}

	client := e.HTTPClient
	if client == nil {
		client = http.DefaultClient
	}
	resp, err := client.Do(req)
	if err != nil {
		return application.ExportResult{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return application.ExportResult{}, fmt.Errorf("nex api returned status %d", resp.StatusCode)
	}
	return application.ExportResult{
		Status:      application.ExportStatusSuccess,
		Destination: "nex_api",
		Reference:   e.Endpoint,
		CreatedAt:   time.Now().UTC(),
	}, nil
}
