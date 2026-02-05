# CLAUDE.md

Hotpot throws cloud security data into one pot—raw data simmers through bronze, silver, and gold layers into insights, compliance reports, and AI-powered answers.

## Read First

1. `docs/architecture/OVERVIEW.md` — system design, project structure
2. `docs/architecture/PRINCIPLES.md` — architecture rules, patterns
3. `docs/guides/CODE_STYLE.md` — code style, testing conventions
4. `docs/GLOSSARY.md` — terms and abbreviations

## Key Rules

| Rule | Description |
|------|-------------|
| Models location | All models in `src/pkg/base/models/{bronze,silver,gold}/` |
| No cross-layer imports | Layers import `base/` only, not each other |
| Data flow | ingest→bronze, normalize→silver, detect→gold |
| Entry points | `cmd/` is thin, all logic in `src/pkg/` |

## Quick Reference

| Layer | Package | Schema | Purpose |
|-------|---------|--------|---------|
| Bronze | `src/pkg/ingest/` | `bronze.*` | Collect raw data |
| Silver | `src/pkg/normalize/` | `silver.*` | Unify models |
| Gold | `src/pkg/detect/` | `gold.*` | Alerts, compliance |
| Agent | External (WrenAI) | reads all | Text-to-SQL |
| UI | `src/pkg/admin/` | reads all | Web interface |

## Workflow

For complex tasks (new features, multi-file changes):
1. Propose a plan first — list files to create/modify
2. Wait for approval before implementing
3. Implement step by step

## Don't

- Never put models in layer packages — use `base/models/`
- Never import one layer from another — use database
