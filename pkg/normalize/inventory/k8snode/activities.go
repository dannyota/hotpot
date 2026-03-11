package k8snode

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.temporal.io/sdk/activity"

	"danny.vn/hotpot/pkg/base/config"
	entk8snode "danny.vn/hotpot/pkg/storage/ent/inventory/k8snode"
	"danny.vn/hotpot/pkg/storage/ent/inventory/k8snode/inventoryk8snode"
	"danny.vn/hotpot/pkg/storage/ent/inventory/k8snode/inventoryk8snodebronzelink"
	"danny.vn/hotpot/pkg/storage/ent/inventory/k8snode/inventoryk8snodenormalized"
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

	// Load existing normalized IDs for this provider to detect stale rows.
	existing, err := a.entClient.InventoryK8sNodeNormalized.Query().
		Where(inventoryk8snodenormalized.ProviderEQ(params.ProviderKey)).
		Select(inventoryk8snodenormalized.FieldID).
		Strings(ctx)
	if err != nil {
		return nil, fmt.Errorf("query existing normalized for %s: %w", params.ProviderKey, err)
	}

	// Batch upsert all normalized rows.
	for _, rec := range records {
		activeIDs[rec.ResourceID()] = true
	}
	if err := a.upsertNormalizedBatch(ctx, records, now); err != nil {
		return nil, err
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
		deleted, err = a.entClient.InventoryK8sNodeNormalized.Delete().
			Where(inventoryk8snodenormalized.IDIn(staleIDs...)).
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

// upsertNormalizedBatch bulk-upserts a batch of records via raw SQL INSERT...ON CONFLICT.
func (a *Activities) upsertNormalizedBatch(ctx context.Context, batch []NormalizedK8sNode, now time.Time) error {
	if len(batch) == 0 {
		return nil
	}

	var b strings.Builder
	b.WriteString(`INSERT INTO silver.inventory_k8s_node_normalized
		(resource_id, provider, is_base, bronze_table, bronze_resource_id,
		 node_name, cluster_name, node_pool, status, provisioning,
		 cloud_project, cloud_zone, cloud_machine_type,
		 internal_ip, external_ip, merge_keys_json,
		 collected_at, first_collected_at, normalized_at)
		VALUES `)

	const cols = 19
	args := make([]any, 0, len(batch)*cols)
	for i, rec := range batch {
		if i > 0 {
			b.WriteByte(',')
		}
		base := i * cols
		b.WriteByte('(')
		for j := range cols {
			if j > 0 {
				b.WriteByte(',')
			}
			b.WriteByte('$')
			b.WriteString(strconv.Itoa(base + j + 1))
		}
		b.WriteByte(')')
		mergeKeysJSON, _ := json.Marshal(rec.MergeKeys)
		args = append(args,
			rec.ResourceID(), rec.Provider, rec.IsBase, rec.BronzeTable, rec.BronzeResourceID,
			rec.NodeName, rec.ClusterName, rec.NodePool, rec.Status, rec.Provisioning,
			rec.CloudProject, rec.CloudZone, rec.CloudMachineType,
			rec.InternalIP, rec.ExternalIP, string(mergeKeysJSON),
			rec.CollectedAt, rec.FirstCollectedAt, now,
		)
	}
	b.WriteString(` ON CONFLICT (resource_id) DO UPDATE SET
		provider = EXCLUDED.provider,
		is_base = EXCLUDED.is_base,
		bronze_table = EXCLUDED.bronze_table,
		bronze_resource_id = EXCLUDED.bronze_resource_id,
		node_name = EXCLUDED.node_name,
		cluster_name = EXCLUDED.cluster_name,
		node_pool = EXCLUDED.node_pool,
		status = EXCLUDED.status,
		provisioning = EXCLUDED.provisioning,
		cloud_project = EXCLUDED.cloud_project,
		cloud_zone = EXCLUDED.cloud_zone,
		cloud_machine_type = EXCLUDED.cloud_machine_type,
		internal_ip = EXCLUDED.internal_ip,
		external_ip = EXCLUDED.external_ip,
		merge_keys_json = EXCLUDED.merge_keys_json,
		collected_at = EXCLUDED.collected_at,
		first_collected_at = EXCLUDED.first_collected_at,
		normalized_at = EXCLUDED.normalized_at`)

	_, err := a.db.ExecContext(ctx, b.String(), args...)
	if err != nil {
		return fmt.Errorf("upsert normalized batch: %w", err)
	}
	return nil
}

// MergeK8sNodesResult holds merge statistics.
type MergeK8sNodesResult struct {
	Created int
	Updated int
	Deleted int
}

// MergeK8sNodes reads all normalized rows, runs dedup, and writes to silver.inventory_k8s_nodes.
func (a *Activities) MergeK8sNodes(ctx context.Context) (*MergeK8sNodesResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting k8s node merge")

	// Read all normalized rows.
	normalizedRows, err := a.entClient.InventoryK8sNodeNormalized.Query().All(ctx)
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
	existingNodes, err := a.entClient.InventoryK8sNode.Query().
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

		// Wrap create/update + bronze link insert in a transaction.
		tx, err := a.entClient.Tx(ctx)
		if err != nil {
			return nil, fmt.Errorf("begin transaction for %s: %w", nodeID, err)
		}

		if isNew {
			// Create new k8s node.
			err = tx.InventoryK8sNode.Create().
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
				tx.Rollback()
				return nil, fmt.Errorf("create k8s node %s: %w", nodeID, err)
			}
		} else {
			// Update existing k8s node.
			_, err = tx.InventoryK8sNode.UpdateOneID(nodeID).
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
				tx.Rollback()
				return nil, fmt.Errorf("update k8s node %s: %w", nodeID, err)
			}

			// Delete old bronze links before inserting new ones.
			_, err = tx.InventoryK8sNodeBronzeLink.Delete().
				Where(inventoryk8snodebronzelink.HasK8sNodeWith(inventoryk8snode.IDEQ(nodeID))).
				Exec(ctx)
			if err != nil {
				tx.Rollback()
				return nil, fmt.Errorf("delete old bronze links for %s: %w", nodeID, err)
			}
		}

		// Bulk insert bronze links.
		builders := make([]*entk8snode.InventoryK8sNodeBronzeLinkCreate, 0, len(m.BronzeLinks))
		for _, link := range m.BronzeLinks {
			builders = append(builders, tx.InventoryK8sNodeBronzeLink.Create().
				SetProvider(link.Provider).
				SetBronzeTable(link.BronzeTable).
				SetBronzeResourceID(link.BronzeResourceID).
				SetK8sNodeID(nodeID))
		}
		if err = tx.InventoryK8sNodeBronzeLink.CreateBulk(builders...).Exec(ctx); err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("create bronze links for %s: %w", nodeID, err)
		}

		if err = tx.Commit(); err != nil {
			return nil, fmt.Errorf("commit merge for %s: %w", nodeID, err)
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
		_, err = a.entClient.InventoryK8sNodeBronzeLink.Delete().
			Where(inventoryk8snodebronzelink.HasK8sNodeWith(inventoryk8snode.IDIn(staleIDs...))).
			Exec(ctx)
		if err != nil {
			slog.Warn("Failed to delete stale bronze links", "count", len(staleIDs), "error", err)
		}

		deleted, err = a.entClient.InventoryK8sNode.Delete().
			Where(inventoryk8snode.IDIn(staleIDs...)).
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
