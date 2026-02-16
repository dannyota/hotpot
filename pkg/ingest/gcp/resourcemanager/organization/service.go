package organization

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzegcporganization"
)

// Service handles GCP Organization ingestion.
type Service struct {
	client    *Client
	entClient *ent.Client
	history   *HistoryService
}

// NewService creates a new organization ingestion service.
func NewService(client *Client, entClient *ent.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestResult contains the result of organization ingestion.
type IngestResult struct {
	OrganizationCount int
	CollectedAt       time.Time
	DurationMillis    int64
}

// Ingest fetches all accessible organizations from GCP and stores them in the bronze layer.
func (s *Service) Ingest(ctx context.Context) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	// Fetch organizations from GCP
	organizations, err := s.client.SearchOrganizations(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to search organizations: %w", err)
	}

	// Convert to organization data
	orgDataList := make([]*OrganizationData, 0, len(organizations))
	for _, org := range organizations {
		data := ConvertOrganization(org, collectedAt)
		orgDataList = append(orgDataList, data)
	}

	// Save to database
	if err := s.saveOrganizations(ctx, orgDataList); err != nil {
		return nil, fmt.Errorf("failed to save organizations: %w", err)
	}

	return &IngestResult{
		OrganizationCount: len(orgDataList),
		CollectedAt:       collectedAt,
		DurationMillis:    time.Since(startTime).Milliseconds(),
	}, nil
}

// saveOrganizations saves organizations to the database with history tracking.
func (s *Service) saveOrganizations(ctx context.Context, organizations []*OrganizationData) error {
	if len(organizations) == 0 {
		return nil
	}

	now := time.Now()

	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	defer func() {
		if v := recover(); v != nil {
			tx.Rollback()
			panic(v)
		}
	}()

	for _, orgData := range organizations {
		// Load existing organization
		existing, err := tx.BronzeGCPOrganization.Query().
			Where(bronzegcporganization.ID(orgData.ID)).
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("failed to load existing organization %s: %w", orgData.ID, err)
		}

		// Compute diff
		diff := DiffOrganizationData(existing, orgData)

		// Skip if no changes
		if !diff.HasAnyChange() && existing != nil {
			// Update collected_at only
			if err := tx.BronzeGCPOrganization.UpdateOneID(orgData.ID).
				SetCollectedAt(orgData.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update collected_at for organization %s: %w", orgData.ID, err)
			}
			continue
		}

		// Create or update organization
		if existing == nil {
			// Create new organization
			create := tx.BronzeGCPOrganization.Create().
				SetID(orgData.ID).
				SetName(orgData.Name).
				SetCollectedAt(orgData.CollectedAt).
				SetFirstCollectedAt(orgData.CollectedAt)

			if orgData.DisplayName != "" {
				create.SetDisplayName(orgData.DisplayName)
			}
			if orgData.State != "" {
				create.SetState(orgData.State)
			}
			if orgData.DirectoryCustomerID != "" {
				create.SetDirectoryCustomerID(orgData.DirectoryCustomerID)
			}
			if orgData.Etag != "" {
				create.SetEtag(orgData.Etag)
			}
			if orgData.CreateTime != "" {
				create.SetCreateTime(orgData.CreateTime)
			}
			if orgData.UpdateTime != "" {
				create.SetUpdateTime(orgData.UpdateTime)
			}
			if orgData.DeleteTime != "" {
				create.SetDeleteTime(orgData.DeleteTime)
			}

			_, err = create.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create organization %s: %w", orgData.ID, err)
			}
		} else {
			// Update existing organization
			update := tx.BronzeGCPOrganization.UpdateOneID(orgData.ID).
				SetName(orgData.Name).
				SetCollectedAt(orgData.CollectedAt)

			if orgData.DisplayName != "" {
				update.SetDisplayName(orgData.DisplayName)
			}
			if orgData.State != "" {
				update.SetState(orgData.State)
			}
			if orgData.DirectoryCustomerID != "" {
				update.SetDirectoryCustomerID(orgData.DirectoryCustomerID)
			}
			if orgData.Etag != "" {
				update.SetEtag(orgData.Etag)
			}
			if orgData.CreateTime != "" {
				update.SetCreateTime(orgData.CreateTime)
			}
			if orgData.UpdateTime != "" {
				update.SetUpdateTime(orgData.UpdateTime)
			}
			if orgData.DeleteTime != "" {
				update.SetDeleteTime(orgData.DeleteTime)
			}

			_, err = update.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update organization %s: %w", orgData.ID, err)
			}
		}

		// Track history
		if diff.IsNew {
			if err := s.history.CreateHistory(ctx, tx, orgData, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create history for organization %s: %w", orgData.ID, err)
			}
		} else {
			if err := s.history.UpdateHistory(ctx, tx, existing, orgData, diff, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update history for organization %s: %w", orgData.ID, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// DeleteStaleOrganizations removes organizations that were not collected in the latest run.
// Also closes history records for deleted organizations.
func (s *Service) DeleteStaleOrganizations(ctx context.Context, collectedAt time.Time) error {
	now := time.Now()

	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	defer func() {
		if v := recover(); v != nil {
			tx.Rollback()
			panic(v)
		}
	}()

	// Find stale organizations
	staleOrgs, err := tx.BronzeGCPOrganization.Query().
		Where(bronzegcporganization.CollectedAtLT(collectedAt)).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Close history and delete each stale organization
	for _, org := range staleOrgs {
		// Close history
		if err := s.history.CloseHistory(ctx, tx, org.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to close history for organization %s: %w", org.ID, err)
		}

		// Delete organization
		if err := tx.BronzeGCPOrganization.DeleteOne(org).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete organization %s: %w", org.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
