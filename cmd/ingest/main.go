package main

import (
	"context"
	"log"

	"hotpot/pkg/base/app"
	"hotpot/pkg/ingest"
)

func main() {
	ctx := context.Background()

	application, err := app.New(app.Options{})
	if err != nil {
		log.Fatalf("Failed to create app: %v", err)
	}

	if err := application.Start(ctx); err != nil {
		log.Fatalf("Failed to start: %v", err)
	}
	defer application.Stop()

	application.Run(ingest.Run)

	if err := application.Wait(); err != nil {
		log.Fatalf("Service failed: %v", err)
	}
}
