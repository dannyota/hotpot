package server

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzegreennodecomputeserver"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzegreennodecomputeserversecgroup"
)

// Service handles GreenNode server ingestion.
type Service struct {
	client    *Client
	entClient *ent.Client
	history   *HistoryService
}

// NewService creates a new server ingestion service.
func NewService(client *Client, entClient *ent.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestResult contains the result of server ingestion.
type IngestResult struct {
	ServerCount    int
	CollectedAt    time.Time
	DurationMillis int64
}

// Ingest fetches servers from GreenNode and stores them in the bronze layer.
func (s *Service) Ingest(ctx context.Context, projectID string) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	servers, err := s.client.ListServers(ctx)
	if err != nil {
		return nil, fmt.Errorf("list servers: %w", err)
	}

	serverDataList := make([]*ServerData, 0, len(servers))
	for _, srv := range servers {
		data, err := ConvertServer(srv, projectID, collectedAt)
		if err != nil {
			return nil, fmt.Errorf("convert server: %w", err)
		}
		serverDataList = append(serverDataList, data)
	}

	if err := s.saveServers(ctx, serverDataList); err != nil {
		return nil, fmt.Errorf("save servers: %w", err)
	}

	return &IngestResult{
		ServerCount:    len(serverDataList),
		CollectedAt:    collectedAt,
		DurationMillis: time.Since(startTime).Milliseconds(),
	}, nil
}

func (s *Service) saveServers(ctx context.Context, servers []*ServerData) error {
	if len(servers) == 0 {
		return nil
	}

	now := time.Now()

	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return fmt.Errorf("start transaction: %w", err)
	}
	defer func() {
		if v := recover(); v != nil {
			tx.Rollback()
			panic(v)
		}
	}()

	for _, data := range servers {
		existing, err := tx.BronzeGreenNodeComputeServer.Query().
			Where(bronzegreennodecomputeserver.ID(data.ID)).
			WithSecGroups().
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("load existing server %s: %w", data.Name, err)
		}

		diff := DiffServerData(existing, data)

		if !diff.HasAnyChange() && existing != nil {
			if err := tx.BronzeGreenNodeComputeServer.UpdateOneID(data.ID).
				SetCollectedAt(data.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("update collected_at for server %s: %w", data.Name, err)
			}
			continue
		}

		if existing != nil {
			if err := s.deleteServerChildren(ctx, tx, data.ID); err != nil {
				tx.Rollback()
				return fmt.Errorf("delete children for server %s: %w", data.Name, err)
			}
		}

		var savedServer *ent.BronzeGreenNodeComputeServer
		if existing == nil {
			create := tx.BronzeGreenNodeComputeServer.Create().
				SetID(data.ID).
				SetName(data.Name).
				SetStatus(data.Status).
				SetLocation(data.Location).
				SetZoneID(data.ZoneID).
				SetCreatedAtAPI(data.CreatedAtAPI).
				SetBootVolumeID(data.BootVolumeID).
				SetEncryptionVolume(data.EncryptionVolume).
				SetLicence(data.Licence).
				SetMetadata(data.Metadata).
				SetMigrateState(data.MigrateState).
				SetProduct(data.Product).
				SetServerGroupID(data.ServerGroupID).
				SetServerGroupName(data.ServerGroupName).
				SetSSHKeyName(data.SSHKeyName).
				SetStopBeforeMigrate(data.StopBeforeMigrate).
				SetUser(data.User).
				SetImageID(data.ImageID).
				SetImageType(data.ImageType).
				SetImageVersion(data.ImageVersion).
				SetFlavorID(data.FlavorID).
				SetFlavorName(data.FlavorName).
				SetFlavorCPU(data.FlavorCPU).
				SetFlavorMemory(data.FlavorMemory).
				SetFlavorGpu(data.FlavorGPU).
				SetFlavorBandwidth(data.FlavorBandwidth).
				SetProjectID(data.ProjectID).
				SetCollectedAt(data.CollectedAt).
				SetFirstCollectedAt(data.CollectedAt)

			if data.InterfacesJSON != nil {
				create.SetInterfacesJSON(data.InterfacesJSON)
			}

			savedServer, err = create.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("create server %s: %w", data.Name, err)
			}
		} else {
			update := tx.BronzeGreenNodeComputeServer.UpdateOneID(data.ID).
				SetName(data.Name).
				SetStatus(data.Status).
				SetLocation(data.Location).
				SetZoneID(data.ZoneID).
				SetCreatedAtAPI(data.CreatedAtAPI).
				SetBootVolumeID(data.BootVolumeID).
				SetEncryptionVolume(data.EncryptionVolume).
				SetLicence(data.Licence).
				SetMetadata(data.Metadata).
				SetMigrateState(data.MigrateState).
				SetProduct(data.Product).
				SetServerGroupID(data.ServerGroupID).
				SetServerGroupName(data.ServerGroupName).
				SetSSHKeyName(data.SSHKeyName).
				SetStopBeforeMigrate(data.StopBeforeMigrate).
				SetUser(data.User).
				SetImageID(data.ImageID).
				SetImageType(data.ImageType).
				SetImageVersion(data.ImageVersion).
				SetFlavorID(data.FlavorID).
				SetFlavorName(data.FlavorName).
				SetFlavorCPU(data.FlavorCPU).
				SetFlavorMemory(data.FlavorMemory).
				SetFlavorGpu(data.FlavorGPU).
				SetFlavorBandwidth(data.FlavorBandwidth).
				SetProjectID(data.ProjectID).
				SetCollectedAt(data.CollectedAt)

			if data.InterfacesJSON != nil {
				update.SetInterfacesJSON(data.InterfacesJSON)
			}

			savedServer, err = update.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("update server %s: %w", data.Name, err)
			}
		}

		if err := s.createServerChildren(ctx, tx, savedServer, data); err != nil {
			tx.Rollback()
			return fmt.Errorf("create children for server %s: %w", data.Name, err)
		}

		if diff.IsNew {
			if err := s.history.CreateHistory(ctx, tx, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("create history for server %s: %w", data.Name, err)
			}
		} else {
			if err := s.history.UpdateHistory(ctx, tx, existing, data, diff, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("update history for server %s: %w", data.Name, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

func (s *Service) deleteServerChildren(ctx context.Context, tx *ent.Tx, serverID string) error {
	_, err := tx.BronzeGreenNodeComputeServerSecGroup.Delete().
		Where(bronzegreennodecomputeserversecgroup.HasServerWith(bronzegreennodecomputeserver.ID(serverID))).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("delete sec groups: %w", err)
	}
	return nil
}

func (s *Service) createServerChildren(ctx context.Context, tx *ent.Tx, server *ent.BronzeGreenNodeComputeServer, data *ServerData) error {
	for _, sg := range data.SecGroups {
		_, err := tx.BronzeGreenNodeComputeServerSecGroup.Create().
			SetServer(server).
			SetUUID(sg.UUID).
			SetName(sg.Name).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("create sec group %s: %w", sg.Name, err)
		}
	}
	return nil
}

// DeleteStaleServers removes servers not collected in the latest run.
func (s *Service) DeleteStaleServers(ctx context.Context, projectID string, collectedAt time.Time) error {
	now := time.Now()

	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return fmt.Errorf("start transaction: %w", err)
	}
	defer func() {
		if v := recover(); v != nil {
			tx.Rollback()
			panic(v)
		}
	}()

	stale, err := tx.BronzeGreenNodeComputeServer.Query().
		Where(
			bronzegreennodecomputeserver.ProjectID(projectID),
			bronzegreennodecomputeserver.CollectedAtLT(collectedAt),
		).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("query stale servers: %w", err)
	}

	for _, srv := range stale {
		if err := s.history.CloseHistory(ctx, tx, srv.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("close history for server %s: %w", srv.ID, err)
		}
		if err := s.deleteServerChildren(ctx, tx, srv.ID); err != nil {
			tx.Rollback()
			return fmt.Errorf("delete children for server %s: %w", srv.ID, err)
		}
		if err := tx.BronzeGreenNodeComputeServer.DeleteOneID(srv.ID).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("delete server %s: %w", srv.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}
