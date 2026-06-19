package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"projeto_pos/backend/internal/adapters/documentreader/manualjson"
	"projeto_pos/backend/internal/adapters/export/csvexport"
	"projeto_pos/backend/internal/adapters/export/jsonexport"
	"projeto_pos/backend/internal/application"
)

func main() {
	input := flag.String("input", "testdata/sample-import-proposal.json", "arquivo JSON de entrada controlada")
	jsonOut := flag.String("json-out", "out/import-proposal-result.json", "arquivo de saída JSON")
	csvOut := flag.String("csv-out", "out/import-proposal-items.csv", "arquivo de saída CSV")
	requiresReview := flag.Bool("requires-review", true, "marca origem como dependente de revisão humana")
	flag.Parse()

	if err := os.MkdirAll("out", 0755); err != nil {
		panic(err)
	}

	service := application.NewImportPurchaseService(manualjson.NewReader())
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

	fmt.Printf("proposal_id=%s\n", proposal.ID)
	fmt.Printf("status=%s\n", proposal.Status)
	fmt.Printf("validation_results=%d\n", len(proposal.ValidationResults))
	fmt.Printf("json_out=%s\n", *jsonOut)
	fmt.Printf("csv_out=%s\n", *csvOut)
}
