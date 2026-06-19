package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"projeto_pos/backend/internal/adapters/documentreader/imageocr"
	"projeto_pos/backend/internal/adapters/documentreader/manualjson"
	pdfreader "projeto_pos/backend/internal/adapters/documentreader/pdf"
	xlsxreader "projeto_pos/backend/internal/adapters/documentreader/xlsx"
	"projeto_pos/backend/internal/adapters/export/csvexport"
	"projeto_pos/backend/internal/adapters/export/jsonexport"
	"projeto_pos/backend/internal/adapters/export/xlsxexport"
	"projeto_pos/backend/internal/adapters/persistence/memory"
	"projeto_pos/backend/internal/application"
	"projeto_pos/backend/internal/domain"
)

func main() {
	input := flag.String("input", "testdata/sample-import-proposal.json", "controlled input file")
	readerName := flag.String("reader", "auto", "input reader: auto, manualjson, xlsx, pdf, imageocr")
	jsonOut := flag.String("json-out", "out/import-proposal-result.json", "proposal JSON output")
	csvOut := flag.String("csv-out", "out/import-proposal-items.csv", "proposal CSV output")
	approvedCSVOut := flag.String("approved-csv-out", "out/approved-purchase-items.csv", "approved purchase CSV output")
	approvedXLSXOut := flag.String("approved-xlsx-out", "out/approved-purchase-items.xlsx", "approved purchase XLSX output")
	requiresReview := flag.Bool("requires-review", true, "mark source as requiring human review")
	approve := flag.Bool("approve", false, "approve and export purchase if there are no blocking validations")
	flag.Parse()

	if err := os.MkdirAll("out", 0755); err != nil {
		panic(err)
	}

	reader, err := selectReader(*readerName, *input)
	if err != nil {
		panic(err)
	}

	repository := memory.NewRepository()
	service := application.NewImportPurchaseServiceWithDependencies(reader, repository, nil)
	proposal, err := service.Import(context.Background(), application.ImportPurchaseInput{
		Path:                *input,
		RequiresHumanReview: *requiresReview,
	})
	if err != nil {
		panic(err)
	}

	if err := jsonexport.NewExporter().Export(context.Background(), proposal, *jsonOut); err != nil {
		panic(err)
	}
	if err := csvexport.NewExporter().Export(context.Background(), proposal, *csvOut); err != nil {
		panic(err)
	}

	if *approve {
		if domain.HasBlocking(proposal.ValidationResults) {
			panic(errors.New("proposal has blocking validations and cannot be approved"))
		}
		approved, err := application.NewApprovePurchaseService(repository).Approve(context.Background(), proposal.ID, "cli")
		if err != nil {
			panic(err)
		}
		if _, err := csvexport.NewManagementExporter(*approvedCSVOut).ExportApprovedPurchase(context.Background(), approved); err != nil {
			panic(err)
		}
		if _, err := xlsxexport.NewExporter(*approvedXLSXOut).ExportApprovedPurchase(context.Background(), approved); err != nil {
			panic(err)
		}
		fmt.Printf("approved_csv_out=%s\n", *approvedCSVOut)
		fmt.Printf("approved_xlsx_out=%s\n", *approvedXLSXOut)
	}

	fmt.Printf("proposal_id=%s\n", proposal.ID)
	fmt.Printf("status=%s\n", proposal.Status)
	fmt.Printf("validation_results=%d\n", len(proposal.ValidationResults))
	fmt.Printf("json_out=%s\n", *jsonOut)
	fmt.Printf("csv_out=%s\n", *csvOut)
}

func selectReader(name string, input string) (application.DocumentReader, error) {
	if name == "auto" {
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
