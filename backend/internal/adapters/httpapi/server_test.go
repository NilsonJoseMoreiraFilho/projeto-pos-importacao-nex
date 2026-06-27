package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"projeto_pos/backend/internal/adapters/persistence/memory"
	"projeto_pos/backend/internal/application"
)

func TestCreateImportAndGetProposal(t *testing.T) {
	server := httptest.NewServer(NewServer(memory.NewRepository(), t.TempDir(), ""))
	defer server.Close()

	body, contentType := multipartBody(t, "file", "sample.json", readFile(t, "../../..//testdata/sample-import-proposal.json"), map[string]string{
		"reader":              "manualjson",
		"requiresHumanReview": "true",
	})
	resp, err := http.Post(server.URL+"/api/imports", contentType, body)
	if err != nil {
		t.Fatalf("post import: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		data, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 201, got %d: %s", resp.StatusCode, string(data))
	}
	var proposal application.ImportProposal
	if err := json.NewDecoder(resp.Body).Decode(&proposal); err != nil {
		t.Fatalf("decode proposal: %v", err)
	}
	if proposal.ID == "" {
		t.Fatal("expected proposal id")
	}

	getResp, err := http.Get(server.URL + "/api/imports/" + proposal.ID)
	if err != nil {
		t.Fatalf("get import: %v", err)
	}
	defer getResp.Body.Close()
	if getResp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", getResp.StatusCode)
	}
}

func TestReviewEditsAndApproveExportsPurchase(t *testing.T) {
	server := httptest.NewServer(NewServer(memory.NewRepository(), t.TempDir(), ""))
	defer server.Close()

	content := strings.Replace(string(readFile(t, "../../..//testdata/sample-import-proposal.json")), `"currentPage": 1`, `"currentPage": 2`, 1)
	body, contentType := multipartBody(t, "file", "sample.json", []byte(content), map[string]string{
		"reader":              "manualjson",
		"requiresHumanReview": "false",
	})
	resp, err := http.Post(server.URL+"/api/imports", contentType, body)
	if err != nil {
		t.Fatalf("post import: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		data, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 201, got %d: %s", resp.StatusCode, string(data))
	}
	var proposal application.ImportProposal
	if err := json.NewDecoder(resp.Body).Decode(&proposal); err != nil {
		t.Fatalf("decode proposal: %v", err)
	}

	reviewBody := bytes.NewBufferString(`{"edits":[{"fieldPath":"purchase.documentNumber","value":"11441701","editedBy":"tester"}]}`)
	reviewResp, err := http.Post(server.URL+"/api/imports/"+proposal.ID+"/review", "application/json", reviewBody)
	if err != nil {
		t.Fatalf("post review: %v", err)
	}
	defer reviewResp.Body.Close()
	if reviewResp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(reviewResp.Body)
		t.Fatalf("expected 200, got %d: %s", reviewResp.StatusCode, string(data))
	}

	approveResp, err := http.Post(server.URL+"/api/imports/"+proposal.ID+"/approve", "application/json", bytes.NewBufferString(`{"reviewer":"tester"}`))
	if err != nil {
		t.Fatalf("post approve: %v", err)
	}
	defer approveResp.Body.Close()
	if approveResp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(approveResp.Body)
		t.Fatalf("expected 200, got %d: %s", approveResp.StatusCode, string(data))
	}
	var output struct {
		ExportResult application.ExportResult `json:"exportResult"`
	}
	if err := json.NewDecoder(approveResp.Body).Decode(&output); err != nil {
		t.Fatalf("decode approve response: %v", err)
	}
	if output.ExportResult.Destination != "csv" {
		t.Fatalf("unexpected export destination: %s", output.ExportResult.Destination)
	}
	if _, err := os.Stat(output.ExportResult.Reference); err != nil {
		t.Fatalf("expected export file to exist: %v", err)
	}
}

func TestSelectReaderRejectsUnknownExtension(t *testing.T) {
	_, err := SelectReader("auto", "pedido.txt")
	if err == nil {
		t.Fatal("expected error for unknown extension")
	}
}

func multipartBody(t *testing.T, fieldName string, fileName string, fileContent []byte, fields map[string]string) (*bytes.Buffer, string) {
	t.Helper()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	for key, value := range fields {
		if err := writer.WriteField(key, value); err != nil {
			t.Fatalf("write field: %v", err)
		}
	}
	part, err := writer.CreateFormFile(fieldName, fileName)
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := part.Write(fileContent); err != nil {
		t.Fatalf("write file: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart: %v", err)
	}
	return body, writer.FormDataContentType()
}

func readFile(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file %s: %v", path, err)
	}
	return data
}

func TestRunCanStartAndStop(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := Run(ctx, "127.0.0.1:0", t.TempDir(), ""); err != nil {
		t.Fatalf("run stopped with error: %v", err)
	}
}
