package imageocr

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"projeto_pos/backend/internal/adapters/documentreader/internal/structuredjson"
	"projeto_pos/backend/internal/application"
	"projeto_pos/backend/internal/domain"
)

type OCRClient interface {
	Extract(ctx context.Context, imagePath string) (application.RawDocumentExtraction, error)
}

type Reader struct {
	Client OCRClient
}

func NewReader(client OCRClient) Reader {
	return Reader{Client: client}
}

func (r Reader) Read(ctx context.Context, path string) (application.RawDocumentExtraction, error) {
	if !isImage(path) {
		return application.RawDocumentExtraction{}, fmt.Errorf("image OCR reader expects image file: %s", path)
	}
	if r.Client == nil {
		return application.RawDocumentExtraction{}, fmt.Errorf("image OCR reader requires OCR client")
	}
	return r.Client.Extract(ctx, path)
}

type SidecarJSONClient struct {
	Suffix        string
	FallbackPaths []string
	OpenAI        OpenAIClient
}

func NewSidecarJSONClient() SidecarJSONClient {
	fallbacks := []string{
		os.Getenv("IMAGE_OCR_FALLBACK_JSON"),
		"frontend/demo-import-proposal.json",
		"../frontend/demo-import-proposal.json",
		"backend/testdata/sample-import-proposal.json",
		"../backend/testdata/sample-import-proposal.json",
		"testdata/sample-import-proposal.json",
	}
	return SidecarJSONClient{
		Suffix:        ".ocr.json",
		FallbackPaths: fallbacks,
		OpenAI:        NewOpenAIClientFromEnv(),
	}
}

func (c SidecarJSONClient) Extract(ctx context.Context, imagePath string) (application.RawDocumentExtraction, error) {
	sidecar := imagePath + c.Suffix
	if _, err := os.Stat(sidecar); err == nil {
		return structuredjson.ReadRawExtraction(ctx, sidecar)
	}
	if c.OpenAI.Enabled() {
		return c.OpenAI.Extract(ctx, imagePath)
	}
	for _, fallback := range c.FallbackPaths {
		if strings.TrimSpace(fallback) == "" {
			continue
		}
		if _, err := os.Stat(fallback); err == nil {
			raw, err := structuredjson.ReadRawExtraction(ctx, fallback)
			if err != nil {
				return application.RawDocumentExtraction{}, err
			}
			raw.SourceDocument.FileName = filepath.Base(imagePath)
			raw.SourceDocument.FileType = "image/" + strings.TrimPrefix(strings.ToLower(filepath.Ext(imagePath)), ".")
			raw.SourceDocument.StorageLocation = imagePath
			raw.SourceDocument.Observations = append(raw.SourceDocument.Observations, "extracao demonstrativa por fallback; OCR real ainda nao implementado")
			return raw, nil
		}
	}
	return application.RawDocumentExtraction{}, fmt.Errorf("image OCR sidecar not found for %s and no fallback JSON is available", imagePath)
}

type OpenAIClient struct {
	APIKey     string
	Model      string
	Endpoint   string
	HTTPClient *http.Client
}

func NewOpenAIClientFromEnv() OpenAIClient {
	return OpenAIClient{
		APIKey:   strings.TrimSpace(os.Getenv("OPENAI_API_KEY")),
		Model:    envOrDefault("OPENAI_VISION_MODEL", "gpt-4o"),
		Endpoint: envOrDefault("OPENAI_RESPONSES_ENDPOINT", "https://api.openai.com/v1/responses"),
		HTTPClient: &http.Client{
			Timeout: 90 * time.Second,
		},
	}
}

func (c OpenAIClient) Enabled() bool {
	return strings.TrimSpace(c.APIKey) != ""
}

func (c OpenAIClient) Extract(ctx context.Context, imagePath string) (application.RawDocumentExtraction, error) {
	if !c.Enabled() {
		return application.RawDocumentExtraction{}, fmt.Errorf("openai client requires OPENAI_API_KEY")
	}
	image, err := os.ReadFile(imagePath)
	if err != nil {
		return application.RawDocumentExtraction{}, err
	}
	mimeType := mimeTypeForImage(imagePath)
	requestBody := openAIResponsesRequest{
		Model:           c.Model,
		MaxOutputTokens: 12000,
		Input: []openAIInput{
			{
				Role: "system",
				Content: []openAIContent{
					{Type: "input_text", Text: extractionSystemPrompt()},
				},
			},
			{
				Role: "user",
				Content: []openAIContent{
					{Type: "input_text", Text: "Extraia os dados deste pedido de compra/venda fotografado. A tabela de itens costuma ter muitas linhas; percorra do primeiro item ate a linha anterior a Total Produtos e responda apenas no JSON do schema."},
					{Type: "input_image", ImageURL: fmt.Sprintf("data:%s;base64,%s", mimeType, base64.StdEncoding.EncodeToString(image))},
				},
			},
		},
		Text: openAIText{
			Format: extractionSchema(),
		},
	}
	payload, err := json.Marshal(requestBody)
	if err != nil {
		return application.RawDocumentExtraction{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.Endpoint, bytes.NewReader(payload))
	if err != nil {
		return application.RawDocumentExtraction{}, err
	}
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("Content-Type", "application/json")

	httpClient := c.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return application.RawDocumentExtraction{}, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return application.RawDocumentExtraction{}, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return application.RawDocumentExtraction{}, fmt.Errorf("openai extraction failed with status %d: %s", resp.StatusCode, string(body))
	}
	var response openAIResponsesResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return application.RawDocumentExtraction{}, err
	}
	content := response.OutputText()
	if strings.TrimSpace(content) == "" {
		return application.RawDocumentExtraction{}, fmt.Errorf("openai extraction returned empty output")
	}
	var extracted aiPurchaseExtraction
	if err := json.Unmarshal([]byte(content), &extracted); err != nil {
		return application.RawDocumentExtraction{}, fmt.Errorf("parse openai extraction JSON: %w: %s", err, content)
	}
	raw := extracted.toRawDocumentExtraction(imagePath)
	raw.SourceDocument.Observations[0] = fmt.Sprintf("extracao real por OpenAI Vision (%s)", c.Model)
	return raw, nil
}

func isImage(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".jpg", ".jpeg", ".png", ".webp", ".tif", ".tiff":
		return true
	default:
		return false
	}
}

func envOrDefault(key string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func mimeTypeForImage(path string) string {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".webp":
		return "image/webp"
	case ".tif", ".tiff":
		return "image/tiff"
	default:
		return "image/jpeg"
	}
}

func extractionSystemPrompt() string {
	return strings.TrimSpace(`Voce extrai dados de pedidos de compra/venda fotografados para importacao em sistema de estoque.
Responda exclusivamente com JSON valido no schema solicitado.
Use strings vazias quando nao conseguir ler um campo.
Use valores numericos em reais, nao em centavos.
Preserve uma linha por item visivel na tabela.
Leia a tabela linha a linha, da coluna Codigo ate Vl. Total.
Nao pare depois das primeiras linhas: continue ate a linha imediatamente anterior a "Total Produtos".
Se existirem linhas destacadas em amarelo, trate cada faixa amarela como uma linha candidata de item.
O documento de exemplo costuma ter cerca de 30 a 40 linhas de itens; se voce retornar muito menos, inclua um warning explicando que a extracao foi parcial.
Nao substitua produtos ilegíveis por produtos parecidos.
Nao use conhecimento geral para completar descricao, codigo ou referencia.
Se a tabela estiver ilegivel, retorne menos itens com warnings em vez de inventar linhas.
Nao invente produto interno NEX; esta sprint nao exige vinculo item a item com cadastro interno.`)
}

type openAIResponsesRequest struct {
	Model           string        `json:"model"`
	MaxOutputTokens int           `json:"max_output_tokens,omitempty"`
	Input           []openAIInput `json:"input"`
	Text            openAIText    `json:"text"`
}

type openAIInput struct {
	Role    string          `json:"role"`
	Content []openAIContent `json:"content"`
}

type openAIContent struct {
	Type     string `json:"type"`
	Text     string `json:"text,omitempty"`
	ImageURL string `json:"image_url,omitempty"`
}

type openAIText struct {
	Format map[string]any `json:"format"`
}

type openAIResponsesResponse struct {
	Output []struct {
		Type    string `json:"type"`
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	} `json:"output"`
}

func (r openAIResponsesResponse) OutputText() string {
	for _, output := range r.Output {
		for _, content := range output.Content {
			if strings.TrimSpace(content.Text) != "" {
				return content.Text
			}
		}
	}
	return ""
}

type aiPurchaseExtraction struct {
	Supplier struct {
		LegalName         string `json:"legalName"`
		TradeName         string `json:"tradeName"`
		DocumentNumber    string `json:"documentNumber"`
		StateRegistration string `json:"stateRegistration"`
	} `json:"supplier"`
	DocumentNumber string `json:"documentNumber"`
	IssueDate      string `json:"issueDate"`
	PriceTable     string `json:"priceTable"`
	FreightMode    string `json:"freightMode"`
	PageCount      int    `json:"pageCount"`
	CurrentPage    int    `json:"currentPage"`
	Items          []struct {
		LineNumber          int     `json:"lineNumber"`
		SupplierProductCode string  `json:"supplierProductCode"`
		Barcode             string  `json:"barcode"`
		Reference           string  `json:"reference"`
		Description         string  `json:"description"`
		Unit                string  `json:"unit"`
		PackageQuantity     int     `json:"packageQuantity"`
		Quantity            float64 `json:"quantity"`
		UnitCost            float64 `json:"unitCost"`
		TotalCost           float64 `json:"totalCost"`
	} `json:"items"`
	Totals struct {
		ProductsTotal   float64 `json:"productsTotal"`
		Discount        float64 `json:"discount"`
		Addition        float64 `json:"addition"`
		IPI             float64 `json:"ipi"`
		TaxSubstitution float64 `json:"taxSubstitution"`
		FCPST           float64 `json:"fcpSt"`
		GrandTotal      float64 `json:"grandTotal"`
	} `json:"totals"`
	Warnings []string `json:"warnings"`
}

func (e aiPurchaseExtraction) toRawDocumentExtraction(imagePath string) application.RawDocumentExtraction {
	issueDate, _ := time.Parse("2006-01-02", e.IssueDate)
	pageCount := e.PageCount
	if pageCount == 0 {
		pageCount = 1
	}
	currentPage := e.CurrentPage
	if currentPage == 0 {
		currentPage = pageCount
	}
	items := make([]domain.PurchaseItem, 0, len(e.Items))
	for i, item := range e.Items {
		lineNumber := item.LineNumber
		if lineNumber == 0 {
			lineNumber = i + 1
		}
		packageQuantity := item.PackageQuantity
		if packageQuantity == 0 {
			packageQuantity = 1
		}
		items = append(items, domain.PurchaseItem{
			LineNumber:          lineNumber,
			SupplierProductCode: strings.TrimSpace(item.SupplierProductCode),
			Barcode:             strings.TrimSpace(item.Barcode),
			Reference:           strings.TrimSpace(item.Reference),
			Description:         strings.TrimSpace(item.Description),
			Unit:                strings.TrimSpace(item.Unit),
			PackageQuantity:     packageQuantity,
			Quantity:            item.Quantity,
			UnitCost:            domain.NewMoneyFromFloat(item.UnitCost),
			TotalCost:           domain.NewMoneyFromFloat(item.TotalCost),
		})
	}
	observations := append([]string{"extracao real por OpenAI Vision"}, e.Warnings...)
	return application.RawDocumentExtraction{
		SourceDocument: domain.SourceDocument{
			ID:                   filepath.Base(imagePath),
			FileName:             filepath.Base(imagePath),
			FileType:             mimeTypeForImage(imagePath),
			PageCount:            pageCount,
			CurrentPage:          currentPage,
			StorageLocation:      imagePath,
			QualityScore:         0.8,
			DetectedDocumentType: domain.SourceDocumentTypePhoto,
			Observations:         observations,
		},
		Purchase: domain.Purchase{
			Supplier: domain.Supplier{
				LegalName:         strings.TrimSpace(e.Supplier.LegalName),
				TradeName:         strings.TrimSpace(e.Supplier.TradeName),
				DocumentNumber:    strings.TrimSpace(e.Supplier.DocumentNumber),
				StateRegistration: strings.TrimSpace(e.Supplier.StateRegistration),
			},
			DocumentNumber:     strings.TrimSpace(e.DocumentNumber),
			IssueDate:          issueDate,
			SourceDocumentType: domain.SourceDocumentTypePhoto,
			PriceTable:         strings.TrimSpace(e.PriceTable),
			FreightMode:        strings.TrimSpace(e.FreightMode),
			Items:              items,
			Totals: domain.PurchaseTotals{
				ProductsTotal:   domain.NewMoneyFromFloat(e.Totals.ProductsTotal),
				Discount:        domain.NewMoneyFromFloat(e.Totals.Discount),
				Addition:        domain.NewMoneyFromFloat(e.Totals.Addition),
				IPI:             domain.NewMoneyFromFloat(e.Totals.IPI),
				TaxSubstitution: domain.NewMoneyFromFloat(e.Totals.TaxSubstitution),
				FCPST:           domain.NewMoneyFromFloat(e.Totals.FCPST),
				GrandTotal:      domain.NewMoneyFromFloat(e.Totals.GrandTotal),
			},
		},
		ExtractedFields: []application.ExtractedField{
			{FieldPath: "purchase.documentNumber", RawText: e.DocumentNumber, NormalizedValue: e.DocumentNumber, Confidence: 0.75, Page: currentPage, Status: application.FieldExtracted},
			{FieldPath: "purchase.supplier.legalName", RawText: e.Supplier.LegalName, NormalizedValue: e.Supplier.LegalName, Confidence: 0.75, Page: currentPage, Status: application.FieldExtracted},
		},
		ExtractorName:    "openai-vision-reader",
		ExtractorVersion: "0.1.0",
		ExtractedAt:      time.Now().UTC(),
	}
}

func extractionSchema() map[string]any {
	return map[string]any{
		"type": "json_schema",
		"name": "purchase_order_extraction",
		"schema": map[string]any{
			"type":                 "object",
			"additionalProperties": false,
			"required":             []string{"supplier", "documentNumber", "issueDate", "priceTable", "freightMode", "pageCount", "currentPage", "items", "totals", "warnings"},
			"properties": map[string]any{
				"supplier": map[string]any{
					"type":                 "object",
					"additionalProperties": false,
					"required":             []string{"legalName", "tradeName", "documentNumber", "stateRegistration"},
					"properties": map[string]any{
						"legalName":         map[string]any{"type": "string"},
						"tradeName":         map[string]any{"type": "string"},
						"documentNumber":    map[string]any{"type": "string"},
						"stateRegistration": map[string]any{"type": "string"},
					},
				},
				"documentNumber": map[string]any{"type": "string"},
				"issueDate":      map[string]any{"type": "string", "description": "Data em formato YYYY-MM-DD ou string vazia."},
				"priceTable":     map[string]any{"type": "string"},
				"freightMode":    map[string]any{"type": "string"},
				"pageCount":      map[string]any{"type": "integer"},
				"currentPage":    map[string]any{"type": "integer"},
				"items": map[string]any{
					"type": "array",
					"items": map[string]any{
						"type":                 "object",
						"additionalProperties": false,
						"required":             []string{"lineNumber", "supplierProductCode", "barcode", "reference", "description", "unit", "packageQuantity", "quantity", "unitCost", "totalCost"},
						"properties": map[string]any{
							"lineNumber":          map[string]any{"type": "integer"},
							"supplierProductCode": map[string]any{"type": "string"},
							"barcode":             map[string]any{"type": "string"},
							"reference":           map[string]any{"type": "string"},
							"description":         map[string]any{"type": "string"},
							"unit":                map[string]any{"type": "string"},
							"packageQuantity":     map[string]any{"type": "integer"},
							"quantity":            map[string]any{"type": "number"},
							"unitCost":            map[string]any{"type": "number"},
							"totalCost":           map[string]any{"type": "number"},
						},
					},
				},
				"totals": map[string]any{
					"type":                 "object",
					"additionalProperties": false,
					"required":             []string{"productsTotal", "discount", "addition", "ipi", "taxSubstitution", "fcpSt", "grandTotal"},
					"properties": map[string]any{
						"productsTotal":   map[string]any{"type": "number"},
						"discount":        map[string]any{"type": "number"},
						"addition":        map[string]any{"type": "number"},
						"ipi":             map[string]any{"type": "number"},
						"taxSubstitution": map[string]any{"type": "number"},
						"fcpSt":           map[string]any{"type": "number"},
						"grandTotal":      map[string]any{"type": "number"},
					},
				},
				"warnings": map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
			},
		},
		"strict": true,
	}
}
