package httpmonitor

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
)

// Rules holds detection rules loaded from config.httpmonitor_rules.
type Rules struct {
	byKey map[string]*Rule
}

// Rule represents a single detection rule configuration.
type Rule struct {
	ID         int
	IsActive   bool
	Status     string
	Thresholds map[string]float64
}

// loadRules reads all rules from the config table.
func loadRules(ctx context.Context, db *sql.DB) (*Rules, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT rule_id, rule_key, status, is_active, thresholds_json
		FROM config.httpmonitor_rules
		WHERE is_active = true`)
	if err != nil {
		return nil, fmt.Errorf("query httpmonitor rules: %w", err)
	}
	defer rows.Close()

	r := &Rules{byKey: make(map[string]*Rule)}
	for rows.Next() {
		var (
			id         int
			key        string
			status     string
			isActive   bool
			threshJSON []byte
		)
		if err := rows.Scan(&id, &key, &status, &isActive, &threshJSON); err != nil {
			return nil, fmt.Errorf("scan rule: %w", err)
		}
		rule := &Rule{
			ID:       id,
			IsActive: isActive,
			Status:   status,
		}
		if len(threshJSON) > 0 {
			if err := json.Unmarshal(threshJSON, &rule.Thresholds); err != nil {
				return nil, fmt.Errorf("unmarshal thresholds for %s: %w", key, err)
			}
		}
		r.byKey[key] = rule
	}
	return r, rows.Err()
}

// Threshold returns the threshold value for a given rule key and threshold key.
// Returns defaultVal if the rule doesn't exist or the threshold key isn't set.
func (r *Rules) Threshold(ruleKey, key string, defaultVal float64) float64 {
	rule, ok := r.byKey[ruleKey]
	if !ok || rule.Thresholds == nil {
		return defaultVal
	}
	if v, ok := rule.Thresholds[key]; ok {
		return v
	}
	return defaultVal
}

// ThresholdInt returns the threshold as int64.
func (r *Rules) ThresholdInt(ruleKey, key string, defaultVal int64) int64 {
	rule, ok := r.byKey[ruleKey]
	if !ok || rule.Thresholds == nil {
		return defaultVal
	}
	if v, ok := rule.Thresholds[key]; ok {
		return int64(v)
	}
	return defaultVal
}
