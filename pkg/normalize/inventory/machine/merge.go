package machine

import (
	"time"

	"danny.vn/hotpot/pkg/normalize/inventory/mergeutil"
)

// MergedMachine is the result of merging normalized rows from multiple providers.
type MergedMachine struct {
	Hostname         string
	OSType           string
	OSName           string
	Status           string
	InternalIP       string
	ExternalIP       string
	Environment      string
	CloudProject     string
	CloudZone        string
	CloudMachineType string
	Created          *time.Time
	CollectedAt      time.Time
	FirstCollectedAt time.Time
	BronzeLinks      []BronzeLink
}

// BronzeLink tracks which bronze record contributed to a merged machine.
type BronzeLink struct {
	Provider         string
	BronzeTable      string
	BronzeResourceID string
}

// mergedEntry tracks a machine plus its merge keys during the merge process.
type mergedEntry struct {
	machine MergedMachine
	keys    map[string]bool
}

// mergePool accumulates merged machines keyed by named merge keys.
type mergePool struct {
	machines []*mergedEntry
	keyIndex map[string]int // "mac:AA:BB:..." → index into machines
}

func newMergePool() *mergePool {
	return &mergePool{
		keyIndex: make(map[string]int),
	}
}

// add appends a new machine and indexes its merge keys.
func (p *mergePool) add(m MergedMachine, mergeKeys map[string][]string) {
	idx := len(p.machines)
	nsKeys := mergeutil.NamespacedKeys(mergeKeys)
	keySet := make(map[string]bool, len(nsKeys))
	for _, k := range nsKeys {
		keySet[k] = true
	}
	p.machines = append(p.machines, &mergedEntry{
		machine: m,
		keys:    keySet,
	})
	for _, key := range nsKeys {
		p.keyIndex[key] = idx
	}
}

// find returns the index of an existing machine sharing any merge key, or -1.
func (p *mergePool) find(mergeKeys map[string][]string) int {
	for _, key := range mergeutil.NamespacedKeys(mergeKeys) {
		if idx, ok := p.keyIndex[key]; ok {
			return idx
		}
	}
	return -1
}

// absorbKeys adds new merge keys to an existing machine and updates the index.
func (p *mergePool) absorbKeys(idx int, mergeKeys map[string][]string) {
	mm := p.machines[idx]
	for _, key := range mergeutil.NamespacedKeys(mergeKeys) {
		mm.keys[key] = true
		p.keyIndex[key] = idx
	}
}

// MergeMachines takes normalized rows grouped by provider order and runs dedup.
// Provider order determines field priority (first non-empty wins).
func MergeMachines(rows []NormalizedMachine, providerOrder []string) []MergedMachine {
	// Group rows by provider key, preserving order within each provider.
	byProvider := make(map[string][]NormalizedMachine)
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
				// Matched existing machine — enrich it.
				m := &pool.machines[idx].machine
				mergeutil.SetIfEmpty(&m.Hostname, row.Hostname)
				mergeutil.SetIfEmpty(&m.OSType, row.OSType)
				mergeutil.SetIfEmpty(&m.OSName, row.OSName)
				mergeutil.SetIfEmpty(&m.Status, row.Status)
				mergeutil.SetIfEmpty(&m.InternalIP, row.InternalIP)
				mergeutil.SetIfEmpty(&m.ExternalIP, row.ExternalIP)
				mergeutil.SetIfEmpty(&m.Environment, row.Environment)
				mergeutil.SetIfEmpty(&m.CloudProject, row.CloudProject)
				mergeutil.SetIfEmpty(&m.CloudZone, row.CloudZone)
				mergeutil.SetIfEmpty(&m.CloudMachineType, row.CloudMachineType)
				mergeCreated(&m.Created, row.Created)
				mergeTimestamps(m, row.CollectedAt, row.FirstCollectedAt)
				m.BronzeLinks = append(m.BronzeLinks, link)
				pool.absorbKeys(idx, row.MergeKeys)
			} else if row.IsBase {
				// Base provider — create new machine.
				m := MergedMachine{
					Hostname:         row.Hostname,
					OSType:           row.OSType,
					OSName:           row.OSName,
					Status:           row.Status,
					InternalIP:       row.InternalIP,
					ExternalIP:       row.ExternalIP,
					Environment:      row.Environment,
					CloudProject:     row.CloudProject,
					CloudZone:        row.CloudZone,
					CloudMachineType: row.CloudMachineType,
					Created:          row.Created,
					CollectedAt:      row.CollectedAt,
					FirstCollectedAt: row.FirstCollectedAt,
					BronzeLinks:      []BronzeLink{link},
				}
				pool.add(m, row.MergeKeys)
			}
			// Merge-only provider with no match — record is dropped.
		}
	}

	result := make([]MergedMachine, 0, len(pool.machines))
	for _, entry := range pool.machines {
		result = append(result, entry.machine)
	}
	return result
}

func mergeCreated(dst **time.Time, val *time.Time) {
	if val == nil {
		return
	}
	if *dst == nil || val.Before(**dst) {
		*dst = val
	}
}

func mergeTimestamps(m *MergedMachine, collected, firstCollected time.Time) {
	if m.CollectedAt.IsZero() || collected.After(m.CollectedAt) {
		m.CollectedAt = collected
	}
	if m.FirstCollectedAt.IsZero() || firstCollected.Before(m.FirstCollectedAt) {
		m.FirstCollectedAt = firstCollected
	}
}
