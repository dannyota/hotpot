package k8snode

import (
	"time"

	"danny.vn/hotpot/pkg/normalize/inventory/mergeutil"
)

// MergedK8sNode is the result of merging normalized rows from multiple providers.
type MergedK8sNode struct {
	NodeName         string
	ClusterName      string
	NodePool         string
	Status           string
	Provisioning     string
	CloudProject     string
	CloudZone        string
	CloudMachineType string
	InternalIP       string
	ExternalIP       string
	CollectedAt      time.Time
	FirstCollectedAt time.Time
	BronzeLinks      []BronzeLink
}

// BronzeLink tracks which bronze record contributed to a merged k8s node.
type BronzeLink struct {
	Provider         string
	BronzeTable      string
	BronzeResourceID string
}

// mergedEntry tracks a k8s node plus its merge keys during the merge process.
type mergedEntry struct {
	node MergedK8sNode
	keys map[string]bool
}

// mergePool accumulates merged k8s nodes keyed by named merge keys.
type mergePool struct {
	nodes    []*mergedEntry
	keyIndex map[string]int
}

func newMergePool() *mergePool {
	return &mergePool{
		keyIndex: make(map[string]int),
	}
}

// add appends a new node and indexes its merge keys.
func (p *mergePool) add(n MergedK8sNode, mergeKeys map[string][]string) {
	idx := len(p.nodes)
	nsKeys := mergeutil.NamespacedKeys(mergeKeys)
	keySet := make(map[string]bool, len(nsKeys))
	for _, k := range nsKeys {
		keySet[k] = true
	}
	p.nodes = append(p.nodes, &mergedEntry{
		node: n,
		keys: keySet,
	})
	for _, key := range nsKeys {
		p.keyIndex[key] = idx
	}
}

// find returns the index of an existing node sharing any merge key, or -1.
func (p *mergePool) find(mergeKeys map[string][]string) int {
	for _, key := range mergeutil.NamespacedKeys(mergeKeys) {
		if idx, ok := p.keyIndex[key]; ok {
			return idx
		}
	}
	return -1
}

// absorbKeys adds new merge keys to an existing node and updates the index.
func (p *mergePool) absorbKeys(idx int, mergeKeys map[string][]string) {
	entry := p.nodes[idx]
	for _, key := range mergeutil.NamespacedKeys(mergeKeys) {
		entry.keys[key] = true
		p.keyIndex[key] = idx
	}
}

// MergeK8sNodes takes normalized rows grouped by provider order and runs dedup.
// Provider order determines field priority (first non-empty wins).
func MergeK8sNodes(rows []NormalizedK8sNode, providerOrder []string) []MergedK8sNode {
	// Group rows by provider key, preserving order within each provider.
	byProvider := make(map[string][]NormalizedK8sNode)
	for i := range rows {
		byProvider[rows[i].Provider] = append(byProvider[rows[i].Provider], rows[i])
	}

	pool := newMergePool()

	// Process providers in registered order.
	for _, pkey := range providerOrder {
		providerRows := byProvider[pkey]
		for i := range providerRows {
			row := &providerRows[i]
			link := BronzeLink{
				Provider:         row.Provider,
				BronzeTable:      row.BronzeTable,
				BronzeResourceID: row.BronzeResourceID,
			}

			if idx := pool.find(row.MergeKeys); idx >= 0 {
				// Matched existing node — enrich it.
				n := &pool.nodes[idx].node
				mergeutil.SetIfEmpty(&n.NodeName, row.NodeName)
				mergeutil.SetIfEmpty(&n.ClusterName, row.ClusterName)
				mergeutil.SetIfEmpty(&n.NodePool, row.NodePool)
				mergeutil.SetIfEmpty(&n.Status, row.Status)
				mergeutil.SetIfEmpty(&n.Provisioning, row.Provisioning)
				mergeutil.SetIfEmpty(&n.CloudProject, row.CloudProject)
				mergeutil.SetIfEmpty(&n.CloudZone, row.CloudZone)
				mergeutil.SetIfEmpty(&n.CloudMachineType, row.CloudMachineType)
				mergeutil.SetIfEmpty(&n.InternalIP, row.InternalIP)
				mergeutil.SetIfEmpty(&n.ExternalIP, row.ExternalIP)
				mergeTimestamps(n, row.CollectedAt, row.FirstCollectedAt)
				n.BronzeLinks = append(n.BronzeLinks, link)
				pool.absorbKeys(idx, row.MergeKeys)
			} else if row.IsBase {
				// Base provider — create new node.
				n := MergedK8sNode{
					NodeName:         row.NodeName,
					ClusterName:      row.ClusterName,
					NodePool:         row.NodePool,
					Status:           row.Status,
					Provisioning:     row.Provisioning,
					CloudProject:     row.CloudProject,
					CloudZone:        row.CloudZone,
					CloudMachineType: row.CloudMachineType,
					InternalIP:       row.InternalIP,
					ExternalIP:       row.ExternalIP,
					CollectedAt:      row.CollectedAt,
					FirstCollectedAt: row.FirstCollectedAt,
					BronzeLinks:      []BronzeLink{link},
				}
				pool.add(n, row.MergeKeys)
			}
			// Merge-only provider with no match — record is dropped.
		}
	}

	result := make([]MergedK8sNode, 0, len(pool.nodes))
	for _, entry := range pool.nodes {
		result = append(result, entry.node)
	}
	return result
}

func mergeTimestamps(n *MergedK8sNode, collected, firstCollected time.Time) {
	if n.CollectedAt.IsZero() || collected.After(n.CollectedAt) {
		n.CollectedAt = collected
	}
	if n.FirstCollectedAt.IsZero() || firstCollected.Before(n.FirstCollectedAt) {
		n.FirstCollectedAt = firstCollected
	}
}
