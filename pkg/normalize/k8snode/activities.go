package k8snode

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"go.temporal.io/sdk/activity"

	"github.com/dannyota/hotpot/pkg/base/config"
	entk8snode "github.com/dannyota/hotpot/pkg/storage/ent/k8snode"
	"github.com/dannyota/hotpot/pkg/storage/ent/k8snode/silverk8snode"
	"github.com/dannyota/hotpot/pkg/storage/ent/k8snode/silverk8snodebronzelink"
	"github.com/dannyota/hotpot/pkg/storage/ent/k8snode/silverk8snodenormalized"
)

func generateID() string {
	return uuid.New().String()
}

// Activities holds dependencies for normalize/merge Temporal activities.
type Activities struct {
	configService *config.Service
	entClient     *entk8snode.Client
	db            *sql.DB
	providers     map[string]Provider
	providerOrder []string
}

// NewActivities creates an Activities instance.
func NewActivities(configService *config.Service, entClient *entk8snode.Client, db *sql.DB, providers []Provider) *Activities {
	pmap := make(map[string]Provider, len(providers))
	order := make([]string, 0, len(providers))
	for _, p := range providers {
		pmap[p.Key()] = p
		order = append(order, p.Key())
	}
	return &Activities{
		configService: configService,
		entClient:     entClient,
		db:            db,
		providers:     pmap,
		providerOrder: order,
	}
}

// NormalizeK8sNodeProviderActivity function reference for Temporal registration.
var NormalizeK8sNodeProviderActivity = (*Activities).NormalizeK8sNodeProvider

// MergeK8sNodesActivity function reference for Temporal registration.
var MergeK8sNodesActivity = (*Activities).MergeK8sNodes

// NormalizeProviderParams identifies which provider to normalize.
type NormalizeProviderParams struct {
	ProviderKey string
}

// NormalizeProviderResult holds normalization statistics.
type NormalizeProviderResult struct {
	ProviderKey string
	Upserted    int
	Deleted     int
}

// NormalizeProvider loads bronze data for one provider, creates/updates k8s_node_normalized,
// and deletes stale rows.
func (a *Activities) NormalizeK8sNodeProvider(ctx context.Context, params NormalizeProviderParams) (*NormalizeProviderResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Normalizing provider", "provider", params.ProviderKey)

	provider, ok := a.providers[params.ProviderKey]
	if !ok {
		return nil, fmt.Errorf("unknown provider: %s", params.ProviderKey)
	}

	// Load and normalize bronze data.
	records, err := provider.Load(ctx, a.db)
	if err != nil {
		return nil, fmt.Errorf("load %s: %w", params.ProviderKey, err)
	}

	now := time.Now()
	activeIDs := make(map[string]bool, len(records))

	// Load existing normalized rows for this provider to distinguish create vs update.
	existingIDs := make(map[string]bool)
	existing, err := a.entClient.SilverK8sNodeNormalized.Query().
		Where(silverk8snodenormalized.ProviderEQ(params.ProviderKey)).
		Select(silverk8snodenormalized.FieldID).
		Strings(ctx)
	if err != nil {
		return nil, fmt.Errorf("query existing normalized for %s: %w", params.ProviderKey, err)
	}
	for _, id := range existing {
		existingIDs[id] = true
	}

	// Create or update normalized rows.
	for _, rec := range records {
		rid := rec.ResourceID()
		activeIDs[rid] = true

		if existingIDs[rid] {
			// Update existing row.
			_, err := a.entClient.SilverK8sNodeNormalized.UpdateOneID(rid).
				SetProvider(rec.Provider).
				SetIsBase(rec.IsBase).
				SetBronzeTable(rec.BronzeTable).
				SetBronzeResourceID(rec.BronzeResourceID).
				SetNodeName(rec.NodeName).
				SetClusterName(rec.ClusterName).
				SetNodePool(rec.NodePool).
				SetStatus(rec.Status).
				SetProvisioning(rec.Provisioning).
				SetCloudProject(rec.CloudProject).
				SetCloudZone(rec.CloudZone).
				SetCloudMachineType(rec.CloudMachineType).
				SetInternalIP(rec.InternalIP).
				SetExternalIP(rec.ExternalIP).
				SetMergeKeysJSON(rec.MergeKeys).
				SetCollectedAt(rec.CollectedAt).
				SetFirstCollectedAt(rec.FirstCollectedAt).
				SetNormalizedAt(now).
				Save(ctx)
			if err != nil {
				return nil, fmt.Errorf("update normalized %s: %w", rid, err)
			}
		} else {
			// Create new row.
			err := a.entClient.SilverK8sNodeNormalized.Create().
				SetID(rid).
				SetProvider(rec.Provider).
				SetIsBase(rec.IsBase).
				SetBronzeTable(rec.BronzeTable).
				SetBronzeResourceID(rec.BronzeResourceID).
				SetNodeName(rec.NodeName).
				SetClusterName(rec.ClusterName).
				SetNodePool(rec.NodePool).
				SetStatus(rec.Status).
				SetProvisioning(rec.Provisioning).
				SetCloudProject(rec.CloudProject).
				SetCloudZone(rec.CloudZone).
				SetCloudMachineType(rec.CloudMachineType).
				SetInternalIP(rec.InternalIP).
				SetExternalIP(rec.ExternalIP).
				SetMergeKeysJSON(rec.MergeKeys).
				SetCollectedAt(rec.CollectedAt).
				SetFirstCollectedAt(rec.FirstCollectedAt).
				SetNormalizedAt(now).
				Exec(ctx)
			if err != nil {
				return nil, fmt.Errorf("create normalized %s: %w", rid, err)
			}
		}
	}

	// Delete stale normalized rows for this provider.
	var staleIDs []string
	for _, id := range existing {
		if !activeIDs[id] {
			staleIDs = append(staleIDs, id)
		}
	}

	deleted := 0
	if len(staleIDs) > 0 {
		deleted, err = a.entClient.SilverK8sNodeNormalized.Delete().
			Where(silverk8snodenormalized.IDIn(staleIDs...)).
			Exec(ctx)
		if err != nil {
			slog.Warn("Failed to delete stale normalized rows",
				"provider", params.ProviderKey, "count", len(staleIDs), "error", err)
		}
	}

	logger.Info("Normalized provider",
		"provider", params.ProviderKey,
		"upserted", len(records),
		"deleted", deleted)

	return &NormalizeProviderResult{
		ProviderKey: params.ProviderKey,
		Upserted:    len(records),
		Deleted:     deleted,
	}, nil
}

// MergeK8sNodesResult holds merge statistics.
type MergeK8sNodesResult struct {
	Created int
	Updated int
	Deleted int
}

// MergeK8sNodes reads all normalized rows, runs dedup, and writes to silver.k8s_nodes.
func (a *Activities) MergeK8sNodes(ctx context.Context) (*MergeK8sNodesResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting k8s node merge")

	// Read all normalized rows.
	normalizedRows, err := a.entClient.SilverK8sNodeNormalized.Query().All(ctx)
	if err != nil {
		return nil, fmt.Errorf("query normalized rows: %w", err)
	}

	// Convert ent models to domain types.
	rows := make([]NormalizedK8sNode, 0, len(normalizedRows))
	for _, r := range normalizedRows {
		rows = append(rows, NormalizedK8sNode{
			Provider:         r.Provider,
			IsBase:           r.IsBase,
			BronzeTable:      r.BronzeTable,
			BronzeResourceID: r.BronzeResourceID,
			NodeName:         r.NodeName,
			ClusterName:      r.ClusterName,
			NodePool:         r.NodePool,
			Status:           r.Status,
			Provisioning:     r.Provisioning,
			CloudProject:     r.CloudProject,
			CloudZone:        r.CloudZone,
			CloudMachineType: r.CloudMachineType,
			InternalIP:       r.InternalIP,
			ExternalIP:       r.ExternalIP,
			CollectedAt:      r.CollectedAt,
			FirstCollectedAt: r.FirstCollectedAt,
			MergeKeys:        r.MergeKeysJSON,
		})
	}

	// Run merge engine.
	merged := MergeK8sNodes(rows, a.providerOrder)

	// Load existing k8s nodes with bronze links for stable ID matching.
	existingNodes, err := a.entClient.SilverK8sNode.Query().
		WithBronzeLinks().
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("query existing k8s nodes: %w", err)
	}

	// Build map: bronze_resource_id → existing node resource_id.
	bronzeToNodeID := make(map[string]string)
	existingNodeIDs := make(map[string]bool, len(existingNodes))
	for _, n := range existingNodes {
		existingNodeIDs[n.ID] = true
		for _, link := range n.Edges.BronzeLinks {
			bronzeToNodeID[link.BronzeResourceID] = n.ID
		}
	}

	now := time.Now()
	var created, updated int

	// Track which existing node IDs are still active.
	activeNodeIDs := make(map[string]bool)

	for _, m := range merged {
		// Find stable ID: check if any bronze_resource_id matches an existing node.
		var nodeID string
		for _, link := range m.BronzeLinks {
			if id, ok := bronzeToNodeID[link.BronzeResourceID]; ok {
				nodeID = id
				break
			}
		}

		isNew := nodeID == ""
		if isNew {
			nodeID = generateID()
			created++
		} else {
			updated++
		}
		activeNodeIDs[nodeID] = true

		if isNew {
			// Create new k8s node.
			err := a.entClient.SilverK8sNode.Create().
				SetID(nodeID).
				SetNodeName(m.NodeName).
				SetClusterName(m.ClusterName).
				SetNodePool(m.NodePool).
				SetStatus(m.Status).
				SetProvisioning(m.Provisioning).
				SetCloudProject(m.CloudProject).
				SetCloudZone(m.CloudZone).
				SetCloudMachineType(m.CloudMachineType).
				SetInternalIP(m.InternalIP).
				SetExternalIP(m.ExternalIP).
				SetCollectedAt(m.CollectedAt).
				SetFirstCollectedAt(m.FirstCollectedAt).
				SetNormalizedAt(now).
				Exec(ctx)
			if err != nil {
				return nil, fmt.Errorf("create k8s node %s: %w", nodeID, err)
			}
		} else {
			// Update existing k8s node.
			_, err := a.entClient.SilverK8sNode.UpdateOneID(nodeID).
				SetNodeName(m.NodeName).
				SetClusterName(m.ClusterName).
				SetNodePool(m.NodePool).
				SetStatus(m.Status).
				SetProvisioning(m.Provisioning).
				SetCloudProject(m.CloudProject).
				SetCloudZone(m.CloudZone).
				SetCloudMachineType(m.CloudMachineType).
				SetInternalIP(m.InternalIP).
				SetExternalIP(m.ExternalIP).
				SetCollectedAt(m.CollectedAt).
				SetNormalizedAt(now).
				Save(ctx)
			if err != nil {
				return nil, fmt.Errorf("update k8s node %s: %w", nodeID, err)
			}

			// Delete old bronze links before inserting new ones.
			_, err = a.entClient.SilverK8sNodeBronzeLink.Delete().
				Where(silverk8snodebronzelink.HasK8sNodeWith(silverk8snode.IDEQ(nodeID))).
				Exec(ctx)
			if err != nil {
				return nil, fmt.Errorf("delete old bronze links for %s: %w", nodeID, err)
			}
		}

		// Insert bronze links.
		for _, link := range m.BronzeLinks {
			err := a.entClient.SilverK8sNodeBronzeLink.Create().
				SetProvider(link.Provider).
				SetBronzeTable(link.BronzeTable).
				SetBronzeResourceID(link.BronzeResourceID).
				SetK8sNodeID(nodeID).
				Exec(ctx)
			if err != nil {
				return nil, fmt.Errorf("create bronze link for %s: %w", nodeID, err)
			}
		}
	}

	// Delete stale k8s nodes.
	var staleIDs []string
	for _, n := range existingNodes {
		if !activeNodeIDs[n.ID] {
			staleIDs = append(staleIDs, n.ID)
		}
	}

	deleted := 0
	if len(staleIDs) > 0 {
		// Delete links first, then nodes.
		_, err = a.entClient.SilverK8sNodeBronzeLink.Delete().
			Where(silverk8snodebronzelink.HasK8sNodeWith(silverk8snode.IDIn(staleIDs...))).
			Exec(ctx)
		if err != nil {
			slog.Warn("Failed to delete stale bronze links", "count", len(staleIDs), "error", err)
		}

		deleted, err = a.entClient.SilverK8sNode.Delete().
			Where(silverk8snode.IDIn(staleIDs...)).
			Exec(ctx)
		if err != nil {
			slog.Warn("Failed to delete stale k8s nodes", "count", len(staleIDs), "error", err)
		}
	}

	logger.Info("K8s node merge complete",
		"created", created,
		"updated", updated,
		"deleted", deleted,
		"total", len(merged))

	return &MergeK8sNodesResult{
		Created: created,
		Updated: updated,
		Deleted: deleted,
	}, nil
}
