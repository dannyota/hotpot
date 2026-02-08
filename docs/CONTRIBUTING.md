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

Follow **Conventional Commits** format:

```
<type>: <subject>

<optional body>
```

**Types:**
- `feat:` - New feature
- `fix:` - Bug fix
- `docs:` - Documentation changes
- `refactor:` - Code refactoring (no functional change)
- `test:` - Adding or updating tests
- `chore:` - Build, dependencies, tooling

**Rules:**
- **Subject:** < 72 characters, imperative mood ("add" not "added")
- **Body:** Optional, explain *why* not *what* (code shows what)
- **No Co-Authored-By** lines (project policy)

**Examples:**

```
feat: add GCP firewall ingestion

- Create bronze schema for firewall rules
- Implement converter and diff logic
- Add workflow integration
```

```
fix: prevent dev database = production in migrate tool

Add safety check that compares database URLs before running
atlas migrate diff. Prevents accidental data loss.
```

```
docs: update architecture overview
```

```
refactor: simplify config validation logic
```

## Pull Requests

- Keep PRs focused on a single change
- Update documentation if needed
- Ensure tests pass
