package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"projeto_pos/backend/internal/adapters/documentreader/imageocr"
	"projeto_pos/backend/internal/adapters/documentreader/manualjson"
	pdfreader "projeto_pos/backend/internal/adapters/documentreader/pdf"
	xlsxreader "projeto_pos/backend/internal/adapters/documentreader/xlsx"
	"projeto_pos/backend/internal/adapters/export/csvexport"
	"projeto_pos/backend/internal/adapters/persistence/memory"
	"projeto_pos/backend/internal/application"
)

type Server struct {
	repository *memory.Repository
	uploadDir  string
	mux        *http.ServeMux
}

func NewServer(repository *memory.Repository, uploadDir string, frontendDir string) *Server {
	if repository == nil {
		repository = memory.NewRepository()
	}
	server := &Server{
		repository: repository,
		uploadDir:  uploadDir,
		mux:        http.NewServeMux(),
	}
	server.routes(frontendDir)
	return server
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *Server) routes(frontendDir string) {
	s.mux.HandleFunc("POST /api/imports", s.handleCreateImport)
	s.mux.HandleFunc("GET /api/imports/{id}", s.handleGetImport)
	s.mux.HandleFunc("POST /api/imports/{id}/review", s.handleReviewImport)
	s.mux.HandleFunc("POST /api/imports/{id}/approve", s.handleApproveImport)
	s.mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	if frontendDir != "" {
		fileServer := http.FileServer(http.Dir(frontendDir))
		s.mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Cache-Control", "no-store")
			fileServer.ServeHTTP(w, r)
		}))
	}
}

func (s *Server) handleCreateImport(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("file is required: %w", err))
		return
	}
	defer file.Close()

	if err := os.MkdirAll(s.uploadDir, 0755); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	targetPath := filepath.Join(s.uploadDir, fmt.Sprintf("%d-%s", time.Now().UnixNano(), sanitizeFileName(header.Filename)))
	target, err := os.Create(targetPath)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if _, err := io.Copy(target, file); err != nil {
		target.Close()
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if err := target.Close(); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	readerName := formValue(r, "reader", "auto")
	reader, err := SelectReader(readerName, targetPath)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	requiresReview, err := parseBool(formValue(r, "requiresHumanReview", "true"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	service := application.NewImportPurchaseServiceWithDependencies(reader, s.repository, nil)
	proposal, err := service.Import(r.Context(), application.ImportPurchaseInput{
		Path:                targetPath,
		RequiresHumanReview: requiresReview,
	})
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err)
		return
	}
	writeJSON(w, http.StatusCreated, proposal)
}

func (s *Server) handleGetImport(w http.ResponseWriter, r *http.Request) {
	proposal, err := s.repository.FindProposalByID(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeJSON(w, http.StatusOK, proposal)
}

func (s *Server) handleReviewImport(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Edits []application.ReviewEdit `json:"edits"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	for i := range request.Edits {
		if request.Edits[i].EditedAt.IsZero() {
			request.Edits[i].EditedAt = time.Now().UTC()
		}
		if request.Edits[i].EditedBy == "" {
			request.Edits[i].EditedBy = "operadora"
		}
	}
	proposal, err := application.NewReviewImportProposalService(s.repository).Review(r.Context(), r.PathValue("id"), request.Edits)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err)
		return
	}
	writeJSON(w, http.StatusOK, proposal)
}

func (s *Server) handleApproveImport(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Reviewer string `json:"reviewer"`
	}
	_ = json.NewDecoder(r.Body).Decode(&request)
	if request.Reviewer == "" {
		request.Reviewer = "operadora"
	}
	approved, err := application.NewApprovePurchaseService(s.repository).Approve(r.Context(), r.PathValue("id"), request.Reviewer)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err)
		return
	}
	exportPath := filepath.Join(filepath.Dir(s.uploadDir), "exports", fmt.Sprintf("nex-import-%s.csv", sanitizeFileName(r.PathValue("id"))))
	if err := os.MkdirAll(filepath.Dir(exportPath), 0755); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	exporter := csvexport.NewManagementExporter(exportPath)
	exportResult, err := application.NewExportApprovedPurchaseService(s.repository, exporter).Export(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"approvedPurchase": approved,
		"exportResult":     exportResult,
	})
}

func SelectReader(name string, input string) (application.DocumentReader, error) {
	if name == "" || name == "auto" {
		switch strings.ToLower(filepath.Ext(input)) {
		case ".json":
			name = "manualjson"
		case ".xlsx":
			name = "xlsx"
		case ".pdf":
			name = "pdf"
		case ".jpg", ".jpeg", ".png", ".webp", ".tif", ".tiff":
			name = "imageocr"
		default:
			return nil, fmt.Errorf("could not choose reader for %s", input)
		}
	}

	switch name {
	case "manualjson":
		return manualjson.NewReader(), nil
	case "xlsx":
		return xlsxreader.NewReader(), nil
	case "pdf":
		return pdfreader.NewReader(), nil
	case "imageocr":
		return imageocr.NewReader(imageocr.NewSidecarJSONClient()), nil
	default:
		return nil, fmt.Errorf("unknown reader: %s", name)
	}
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]string{"error": err.Error()})
}

func formValue(r *http.Request, key string, fallback string) string {
	value := strings.TrimSpace(r.FormValue(key))
	if value == "" {
		return fallback
	}
	return value
}

func parseBool(value string) (bool, error) {
	if value == "" {
		return false, nil
	}
	return strconv.ParseBool(value)
}

func sanitizeFileName(name string) string {
	name = filepath.Base(name)
	name = strings.ReplaceAll(name, " ", "-")
	name = strings.Map(func(r rune) rune {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '-' || r == '_' || r == '.' {
			return r
		}
		return '-'
	}, name)
	if name == "" || name == "." {
		return "upload"
	}
	return name
}

func Run(ctx context.Context, addr string, uploadDir string, frontendDir string) error {
	server := &http.Server{
		Addr:              addr,
		Handler:           NewServer(memory.NewRepository(), uploadDir, frontendDir),
		ReadHeaderTimeout: 10 * time.Second,
	}
	errCh := make(chan error, 1)
	go func() {
		errCh <- server.ListenAndServe()
	}()
	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return server.Shutdown(shutdownCtx)
	case err := <-errCh:
		if err == http.ErrServerClosed {
			return nil
		}
		return err
	}
}
