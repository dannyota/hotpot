package eol

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	entreference "github.com/dannyota/hotpot/pkg/storage/ent/reference"
)

const insertBatchSize = 1000

// Service handles EOL data persistence.
type Service struct {
	client    *Client
	entClient *entreference.Client
}

// NewService creates a new EOL service.
func NewService(client *Client, entClient *entreference.Client) *Service {
	return &Service{client: client, entClient: entClient}
}

// IngestResult contains the result of an EOL ingestion.
type IngestResult struct {
	ProductCount    int
	CycleCount      int
	IdentifierCount int
	DurationMillis  int64
}

// Ingest downloads and replaces all EOL data.
func (s *Service) Ingest(ctx context.Context, heartbeat func(string)) (*IngestResult, error) {
	start := time.Now()

	products, err := s.client.Download(heartbeat)
	if err != nil {
		return nil, fmt.Errorf("download EOL data: %w", err)
	}

	// Count total cycles
	totalCycles := 0
	for _, p := range products {
		totalCycles += len(p.Cycles)
	}

	slog.Info("Downloaded EOL data", "products", len(products), "cycles", totalCycles)
	heartbeat(fmt.Sprintf("saving %d products, %d cycles", len(products), totalCycles))

	now := time.Now()

	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	// Delete existing cycles first (references products)
	deletedCycles, err := tx.BronzeReferenceEOLCycle.Delete().Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("delete existing EOL cycles: %w", err)
	}
	slog.Info("Deleted existing EOL cycles", "count", deletedCycles)

	// Delete existing identifiers
	deletedIdentifiers, err := tx.BronzeReferenceEOLIdentifier.Delete().Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("delete existing EOL identifiers: %w", err)
	}
	slog.Info("Deleted existing EOL identifiers", "count", deletedIdentifiers)

	// Delete existing products
	deletedProducts, err := tx.BronzeReferenceEOLProduct.Delete().Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("delete existing EOL products: %w", err)
	}
	slog.Info("Deleted existing EOL products", "count", deletedProducts)

	// Bulk insert products
	for i := 0; i < len(products); i += insertBatchSize {
		end := min(i+insertBatchSize, len(products))
		batch := products[i:end]

		builders := make([]*entreference.BronzeReferenceEOLProductCreate, len(batch))
		for j, p := range batch {
			b := tx.BronzeReferenceEOLProduct.Create().
				SetID(p.Slug).
				SetName(p.Name).
				SetCategory(p.Category).
				SetCollectedAt(now).
				SetFirstCollectedAt(now)
			if len(p.Tags) > 0 {
				b.SetTags(p.Tags)
			}
			builders[j] = b
		}

		if err := tx.BronzeReferenceEOLProduct.CreateBulk(builders...).Exec(ctx); err != nil {
			return nil, fmt.Errorf("bulk insert EOL products batch %d: %w", i/insertBatchSize, err)
		}

		heartbeat(fmt.Sprintf("saved %d/%d products", end, len(products)))
	}

	// Bulk insert cycles
	var allCycles []cycleWithProduct
	for _, p := range products {
		for _, c := range p.Cycles {
			allCycles = append(allCycles, cycleWithProduct{product: p.Slug, cycle: c})
		}
	}

	for i := 0; i < len(allCycles); i += insertBatchSize {
		end := min(i+insertBatchSize, len(allCycles))
		batch := allCycles[i:end]

		builders := make([]*entreference.BronzeReferenceEOLCycleCreate, len(batch))
		for j, item := range batch {
			id := item.product + ":" + item.cycle.Cycle
			b := tx.BronzeReferenceEOLCycle.Create().
				SetID(id).
				SetProduct(item.product).
				SetCycle(item.cycle.Cycle).
				SetCollectedAt(now).
				SetFirstCollectedAt(now)

			if item.cycle.ReleaseDate != nil {
				b.SetReleaseDate(*item.cycle.ReleaseDate)
			}
			if item.cycle.EOAS != nil {
				b.SetEoas(*item.cycle.EOAS)
			}
			if item.cycle.EOL != nil {
				b.SetEol(*item.cycle.EOL)
			}
			if item.cycle.EOES != nil {
				b.SetEoes(*item.cycle.EOES)
			}
			if item.cycle.Latest != "" {
				b.SetLatest(item.cycle.Latest)
			}
			if item.cycle.LatestReleaseDate != nil {
				b.SetLatestReleaseDate(*item.cycle.LatestReleaseDate)
			}
			if item.cycle.LTS != nil {
				b.SetLts(*item.cycle.LTS)
			}

			builders[j] = b
		}

		if err := tx.BronzeReferenceEOLCycle.CreateBulk(builders...).Exec(ctx); err != nil {
			return nil, fmt.Errorf("bulk insert EOL cycles batch %d: %w", i/insertBatchSize, err)
		}

		heartbeat(fmt.Sprintf("saved %d/%d cycles", end, len(allCycles)))
	}

	// Bulk insert identifiers
	var allIdentifiers []identifierWithProduct
	for _, p := range products {
		for _, id := range p.Identifiers {
			allIdentifiers = append(allIdentifiers, identifierWithProduct{product: p.Slug, identifier: id})
		}
	}

	for i := 0; i < len(allIdentifiers); i += insertBatchSize {
		end := min(i+insertBatchSize, len(allIdentifiers))
		batch := allIdentifiers[i:end]

		builders := make([]*entreference.BronzeReferenceEOLIdentifierCreate, len(batch))
		for j, item := range batch {
			id := item.product + ":" + item.identifier.Type + ":" + item.identifier.Value
			builders[j] = tx.BronzeReferenceEOLIdentifier.Create().
				SetID(id).
				SetProduct(item.product).
				SetIdentifierType(item.identifier.Type).
				SetValue(item.identifier.Value).
				SetCollectedAt(now).
				SetFirstCollectedAt(now)
		}

		if err := tx.BronzeReferenceEOLIdentifier.CreateBulk(builders...).Exec(ctx); err != nil {
			return nil, fmt.Errorf("bulk insert EOL identifiers batch %d: %w", i/insertBatchSize, err)
		}

		heartbeat(fmt.Sprintf("saved %d/%d identifiers", end, len(allIdentifiers)))
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	return &IngestResult{
		ProductCount:    len(products),
		CycleCount:      len(allCycles),
		IdentifierCount: len(allIdentifiers),
		DurationMillis:  time.Since(start).Milliseconds(),
	}, nil
}

type cycleWithProduct struct {
	product string
	cycle   CycleData
}

type identifierWithProduct struct {
	product    string
	identifier IdentifierData
}
