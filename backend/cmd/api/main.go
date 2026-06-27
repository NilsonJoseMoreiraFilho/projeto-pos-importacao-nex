package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"

	"projeto_pos/backend/internal/adapters/httpapi"
)

func main() {
	addr := flag.String("addr", ":8080", "HTTP listen address")
	uploadDir := flag.String("upload-dir", "out/uploads", "directory for uploaded documents")
	frontendDir := flag.String("frontend-dir", "../frontend", "directory with static frontend files")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	fmt.Printf("listening on http://localhost%s\n", *addr)
	if err := httpapi.Run(ctx, *addr, *uploadDir, *frontendDir); err != nil {
		panic(err)
	}
}
