package finding

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzegcpsecuritycenterfinding"
)

// Service handles SCC finding ingestion.
type Service struct {
	client    *Client
	entClient *ent.Client
	history   *HistoryService
}

// NewService creates a new SCC finding ingestion service.
func NewService(client *Client, entClient *ent.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestResult contains the result of SCC finding ingestion.
type IngestResult struct {
	FindingCount   int
	CollectedAt    time.Time
	DurationMillis int64
}

// Ingest fetches SCC findings from GCP and stores them in the bronze layer.
func (s *Service) Ingest(ctx context.Context) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	// Fetch findings from GCP
	rawFindings, err := s.client.ListFindings(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list findings: %w", err)
	}

	// Convert to finding data
	findingDataList := make([]*FindingData, 0, len(rawFindings))
	for _, raw := range rawFindings {
		data := ConvertFinding(raw.OrgName, raw.SourceName, raw.Finding, collectedAt)
		if data != nil {
			findingDataList = append(findingDataList, data)
		}
	}

	// Save to database
	if err := s.saveFindings(ctx, findingDataList); err != nil {
		return nil, fmt.Errorf("failed to save findings: %w", err)
	}

	return &IngestResult{
		FindingCount:   len(findingDataList),
		CollectedAt:    collectedAt,
		DurationMillis: time.Since(startTime).Milliseconds(),
	}, nil
}

// saveFindings saves SCC findings to the database with history tracking.
func (s *Service) saveFindings(ctx context.Context, findings []*FindingData) error {
	if len(findings) == 0 {
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

	for _, findingData := range findings {
		// Load existing finding
		existing, err := tx.BronzeGCPSecurityCenterFinding.Query().
			Where(bronzegcpsecuritycenterfinding.ID(findingData.ID)).
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("failed to load existing finding %s: %w", findingData.ID, err)
		}

		// Compute diff
		diff := DiffFindingData(existing, findingData)

		// Skip if no changes
		if !diff.HasAnyChange() && existing != nil {
			if err := tx.BronzeGCPSecurityCenterFinding.UpdateOneID(findingData.ID).
				SetCollectedAt(findingData.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update collected_at for finding %s: %w", findingData.ID, err)
			}
			continue
		}

		// Create or update finding
		if existing == nil {
			create := tx.BronzeGCPSecurityCenterFinding.Create().
				SetID(findingData.ID).
				SetParent(findingData.Parent).
				SetOrganizationID(findingData.OrganizationID).
				SetCollectedAt(findingData.CollectedAt).
				SetFirstCollectedAt(findingData.CollectedAt)

			if findingData.ResourceName != "" {
				create.SetResourceName(findingData.ResourceName)
			}
			if findingData.State != "" {
				create.SetState(findingData.State)
			}
			if findingData.Category != "" {
				create.SetCategory(findingData.Category)
			}
			if findingData.ExternalURI != "" {
				create.SetExternalURI(findingData.ExternalURI)
			}
			if findingData.Severity != "" {
				create.SetSeverity(findingData.Severity)
			}
			if findingData.FindingClass != "" {
				create.SetFindingClass(findingData.FindingClass)
			}
			if findingData.CanonicalName != "" {
				create.SetCanonicalName(findingData.CanonicalName)
			}
			if findingData.Mute != "" {
				create.SetMute(findingData.Mute)
			}
			if findingData.CreateTime != "" {
				create.SetCreateTime(findingData.CreateTime)
			}
			if findingData.EventTime != "" {
				create.SetEventTime(findingData.EventTime)
			}
			if findingData.SourceProperties != nil {
				create.SetSourceProperties(findingData.SourceProperties)
			}
			if findingData.SecurityMarks != nil {
				create.SetSecurityMarks(findingData.SecurityMarks)
			}
			if findingData.Indicator != nil {
				create.SetIndicator(findingData.Indicator)
			}
			if findingData.Vulnerability != nil {
				create.SetVulnerability(findingData.Vulnerability)
			}
			if findingData.Connections != nil {
				create.SetConnections(findingData.Connections)
			}
			if findingData.Compliances != nil {
				create.SetCompliances(findingData.Compliances)
			}
			if findingData.Contacts != nil {
				create.SetContacts(findingData.Contacts)
			}

			_, err = create.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create finding %s: %w", findingData.ID, err)
			}
		} else {
			update := tx.BronzeGCPSecurityCenterFinding.UpdateOneID(findingData.ID).
				SetParent(findingData.Parent).
				SetOrganizationID(findingData.OrganizationID).
				SetCollectedAt(findingData.CollectedAt)

			if findingData.ResourceName != "" {
				update.SetResourceName(findingData.ResourceName)
			}
			if findingData.State != "" {
				update.SetState(findingData.State)
			}
			if findingData.Category != "" {
				update.SetCategory(findingData.Category)
			}
			if findingData.ExternalURI != "" {
				update.SetExternalURI(findingData.ExternalURI)
			}
			if findingData.Severity != "" {
				update.SetSeverity(findingData.Severity)
			}
			if findingData.FindingClass != "" {
				update.SetFindingClass(findingData.FindingClass)
			}
			if findingData.CanonicalName != "" {
				update.SetCanonicalName(findingData.CanonicalName)
			}
			if findingData.Mute != "" {
				update.SetMute(findingData.Mute)
			}
			if findingData.CreateTime != "" {
				update.SetCreateTime(findingData.CreateTime)
			}
			if findingData.EventTime != "" {
				update.SetEventTime(findingData.EventTime)
			}
			if findingData.SourceProperties != nil {
				update.SetSourceProperties(findingData.SourceProperties)
			}
			if findingData.SecurityMarks != nil {
				update.SetSecurityMarks(findingData.SecurityMarks)
			}
			if findingData.Indicator != nil {
				update.SetIndicator(findingData.Indicator)
			}
			if findingData.Vulnerability != nil {
				update.SetVulnerability(findingData.Vulnerability)
			}
			if findingData.Connections != nil {
				update.SetConnections(findingData.Connections)
			}
			if findingData.Compliances != nil {
				update.SetCompliances(findingData.Compliances)
			}
			if findingData.Contacts != nil {
				update.SetContacts(findingData.Contacts)
			}

			_, err = update.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update finding %s: %w", findingData.ID, err)
			}
		}

		// Track history
		if diff.IsNew {
			if err := s.history.CreateHistory(ctx, tx, findingData, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create history for finding %s: %w", findingData.ID, err)
			}
		} else {
			if err := s.history.UpdateHistory(ctx, tx, existing, findingData, diff, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update history for finding %s: %w", findingData.ID, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// DeleteStaleFindings removes findings that were not collected in the latest run.
func (s *Service) DeleteStaleFindings(ctx context.Context, collectedAt time.Time) error {
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

	staleFindings, err := tx.BronzeGCPSecurityCenterFinding.Query().
		Where(bronzegcpsecuritycenterfinding.CollectedAtLT(collectedAt)).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, f := range staleFindings {
		if err := s.history.CloseHistory(ctx, tx, f.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to close history for finding %s: %w", f.ID, err)
		}

		if err := tx.BronzeGCPSecurityCenterFinding.DeleteOne(f).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete finding %s: %w", f.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
