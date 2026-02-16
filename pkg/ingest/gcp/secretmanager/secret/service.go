package secret

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzegcpsecretmanagersecret"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzegcpsecretmanagersecretlabel"
)

// Service handles GCP Secret Manager secret ingestion.
type Service struct {
	client    *Client
	entClient *ent.Client
	history   *HistoryService
}

// NewService creates a new secret ingestion service.
func NewService(client *Client, entClient *ent.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestParams contains parameters for secret ingestion.
type IngestParams struct {
	ProjectID string
}

// IngestResult contains the result of secret ingestion.
type IngestResult struct {
	ProjectID      string
	SecretCount    int
	CollectedAt    time.Time
	DurationMillis int64
}

// Ingest fetches secrets from GCP and stores them in the bronze layer.
func (s *Service) Ingest(ctx context.Context, params IngestParams) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	secrets, err := s.client.ListSecrets(ctx, params.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list secrets: %w", err)
	}

	secretDataList := make([]*SecretData, 0, len(secrets))
	for _, sec := range secrets {
		data, err := ConvertSecret(sec, params.ProjectID, collectedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to convert secret: %w", err)
		}
		secretDataList = append(secretDataList, data)
	}

	if err := s.saveSecrets(ctx, secretDataList); err != nil {
		return nil, fmt.Errorf("failed to save secrets: %w", err)
	}

	return &IngestResult{
		ProjectID:      params.ProjectID,
		SecretCount:    len(secretDataList),
		CollectedAt:    collectedAt,
		DurationMillis: time.Since(startTime).Milliseconds(),
	}, nil
}

func (s *Service) saveSecrets(ctx context.Context, secrets []*SecretData) error {
	if len(secrets) == 0 {
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

	for _, secretData := range secrets {
		existing, err := tx.BronzeGCPSecretManagerSecret.Query().
			Where(bronzegcpsecretmanagersecret.ID(secretData.ID)).
			WithLabels().
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("failed to load existing secret %s: %w", secretData.Name, err)
		}

		diff := DiffSecretData(existing, secretData)

		if !diff.HasAnyChange() && existing != nil {
			if err := tx.BronzeGCPSecretManagerSecret.UpdateOneID(secretData.ID).
				SetCollectedAt(secretData.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update collected_at for secret %s: %w", secretData.Name, err)
			}
			continue
		}

		// Delete old children if updating
		if existing != nil {
			if err := deleteSecretChildren(ctx, tx, secretData.ID); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to delete old children for secret %s: %w", secretData.Name, err)
			}
		}

		var savedSecret *ent.BronzeGCPSecretManagerSecret
		if existing == nil {
			create := tx.BronzeGCPSecretManagerSecret.Create().
				SetID(secretData.ID).
				SetName(secretData.Name).
				SetCreateTime(secretData.CreateTime).
				SetEtag(secretData.Etag).
				SetProjectID(secretData.ProjectID).
				SetCollectedAt(secretData.CollectedAt).
				SetFirstCollectedAt(secretData.CollectedAt)

			if secretData.ReplicationJSON != nil {
				create.SetReplicationJSON(secretData.ReplicationJSON)
			}
			if secretData.RotationJSON != nil {
				create.SetRotationJSON(secretData.RotationJSON)
			}
			if secretData.TopicsJSON != nil {
				create.SetTopicsJSON(secretData.TopicsJSON)
			}
			if secretData.VersionAliasesJSON != nil {
				create.SetVersionAliasesJSON(secretData.VersionAliasesJSON)
			}
			if secretData.AnnotationsJSON != nil {
				create.SetAnnotationsJSON(secretData.AnnotationsJSON)
			}

			savedSecret, err = create.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create secret %s: %w", secretData.Name, err)
			}
		} else {
			update := tx.BronzeGCPSecretManagerSecret.UpdateOneID(secretData.ID).
				SetName(secretData.Name).
				SetCreateTime(secretData.CreateTime).
				SetEtag(secretData.Etag).
				SetProjectID(secretData.ProjectID).
				SetCollectedAt(secretData.CollectedAt)

			if secretData.ReplicationJSON != nil {
				update.SetReplicationJSON(secretData.ReplicationJSON)
			}
			if secretData.RotationJSON != nil {
				update.SetRotationJSON(secretData.RotationJSON)
			}
			if secretData.TopicsJSON != nil {
				update.SetTopicsJSON(secretData.TopicsJSON)
			}
			if secretData.VersionAliasesJSON != nil {
				update.SetVersionAliasesJSON(secretData.VersionAliasesJSON)
			}
			if secretData.AnnotationsJSON != nil {
				update.SetAnnotationsJSON(secretData.AnnotationsJSON)
			}

			savedSecret, err = update.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update secret %s: %w", secretData.Name, err)
			}
		}

		if err := createSecretChildren(ctx, tx, savedSecret, secretData); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to create children for secret %s: %w", secretData.Name, err)
		}

		if diff.IsNew {
			if err := s.history.CreateHistory(ctx, tx, secretData, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create history for secret %s: %w", secretData.Name, err)
			}
		} else {
			if err := s.history.UpdateHistory(ctx, tx, existing, secretData, diff, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update history for secret %s: %w", secretData.Name, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func deleteSecretChildren(ctx context.Context, tx *ent.Tx, secretID string) error {
	_, err := tx.BronzeGCPSecretManagerSecretLabel.Delete().
		Where(bronzegcpsecretmanagersecretlabel.HasSecretWith(bronzegcpsecretmanagersecret.ID(secretID))).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete labels: %w", err)
	}
	return nil
}

func createSecretChildren(ctx context.Context, tx *ent.Tx, savedSecret *ent.BronzeGCPSecretManagerSecret, secretData *SecretData) error {
	for _, label := range secretData.Labels {
		_, err := tx.BronzeGCPSecretManagerSecretLabel.Create().
			SetKey(label.Key).
			SetValue(label.Value).
			SetSecret(savedSecret).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create label: %w", err)
		}
	}
	return nil
}

// DeleteStaleSecrets removes secrets that were not collected in the latest run.
func (s *Service) DeleteStaleSecrets(ctx context.Context, projectID string, collectedAt time.Time) error {
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

	staleSecrets, err := tx.BronzeGCPSecretManagerSecret.Query().
		Where(
			bronzegcpsecretmanagersecret.ProjectID(projectID),
			bronzegcpsecretmanagersecret.CollectedAtLT(collectedAt),
		).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, sec := range staleSecrets {
		if err := s.history.CloseHistory(ctx, tx, sec.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to close history for secret %s: %w", sec.ID, err)
		}

		if err := tx.BronzeGCPSecretManagerSecret.DeleteOne(sec).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete secret %s: %w", sec.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
