# ADR 002: Ent Issue #2330 Workaround - History Table Naming

## Status
Accepted - Temporary workaround

## Context
During GORM to Ent migration, we encountered [ent Issue #2330](https://github.com/ent/ent/issues/2330) which causes code generation failure when multiple types use the same table name across different PostgreSQL schemas.

## Problem
- Bronze schema: `bronze.gcp_compute_instances`
- Bronze history schema: `bronze_history.gcp_compute_instances`
- Ent generates constants based on TABLE NAME only (ignores schema annotation)
- Both generate `GcpComputeInstancesColumns` → duplicate declaration → `entc.Generate()` fails

## Decision
Use `_history` suffix for all history table names:
- Bronze: `bronze.gcp_compute_instances` (unchanged)
- Bronze history: `bronze_history.gcp_compute_instances_history` (added `_history` suffix)

## Consequences

### Positive
- ✅ Code generation succeeds (no duplicate constants)
- ✅ Runtime queries work correctly with `AlternateSchema()`
- ✅ No Atlas Pro required
- ✅ Simple, predictable naming pattern

### Negative
- ❌ Table names don't perfectly mirror between bronze and bronze_history
- ❌ Requires renaming tables from current GORM schema
- ❌ History table names are longer

## Future Cleanup
**When ent fixes Issue #2330** (currently open since Feb 2022):

1. Create migration to rename history tables:
   ```sql
   -- Per resource type
   ALTER TABLE bronze_history.gcp_compute_instances_history
   RENAME TO gcp_compute_instances;
   ```

2. Update ent schema annotations to remove `_history` suffix:
   ```go
   entsql.Annotation{Table: "gcp_compute_instances"}  // remove _history
   ```

3. Regenerate ent code
4. Verify all queries still work
5. Update this document to "Superseded"

## References
- [Ent Issue #2330](https://github.com/ent/ent/issues/2330)
- Schema definitions: `pkg/schema/bronzehistory/`
- Ent documentation: `docs/guides/ENT_SCHEMAS.md`

## Implementation Status
- **Decided:** 2026-02-08 (Plan phase)
- **Implemented:** 2026-02-08 (Migration complete)
- **Status:** ✅ Active - All 111 ent schemas use `_history` suffix pattern
