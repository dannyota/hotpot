# Contributing

Guidelines for contributing to Hotpot.

## Getting Started

1. Fork the repository
2. Create a feature branch
3. Make changes following code style
4. Submit a pull request

## Code Style

See [CODE_STYLE.md](guides/CODE_STYLE.md) for coding conventions.

## Commit Messages

Follow [Go commit message style](https://go.dev/wiki/CommitMessage):

```
pkg/area: short summary in lowercase

Optional body explaining why, not what.
Wrap at ~72 characters.
```

**Rules:**
- First line: `package: summary` (lowercase, no period)
- Keep subject under 72 characters
- Use imperative mood: "add", "fix", "update" (not "added", "fixes")
- Body explains *why* the change was made

**Examples:**

```
ingest/gcp: add compute disk ingestion

Adds support for GCP Compute Disk resources with SCD Type 4
history tracking.
```

```
docs: update architecture overview
```

```
fix: resolve nil pointer in config reload
```

For initial commits or multi-area changes, omit the package prefix.

## Pull Requests

- Keep PRs focused on a single change
- Update documentation if needed
- Ensure tests pass
