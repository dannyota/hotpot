# Hotpot Documentation

Hotpot is a multi-cloud security CMDB using Medallion Data Architecture (Bronze → Silver → Gold) with AI-powered querying.

## 📚 Quick Start

New to Hotpot? Read these in order:

1. **[Architecture Overview](./architecture/OVERVIEW.md)** - System design and project structure
2. **[Architecture Principles](./architecture/PRINCIPLES.md)** - Core architecture rules and patterns
3. **[Code Style Guide](./guides/CODE_STYLE.md)** - Naming conventions and testing standards
4. **[Glossary](./reference/GLOSSARY.md)** - Terms and abbreviations

## 🏗️ Architecture

| Document | Description |
|----------|-------------|
| [OVERVIEW](./architecture/OVERVIEW.md) | System design, layer model, project structure |
| [PRINCIPLES](./architecture/PRINCIPLES.md) | 13 architecture principles (rate limiting, no cross-layer imports, etc.) |
| [HISTORY](./architecture/HISTORY.md) | SCD Type 4 temporal change tracking |

## 📖 Development Guides

### Data Layer (Ent ORM)

| Document | Description |
|----------|-------------|
| [ENT_SCHEMAS](./guides/ENT_SCHEMAS.md) | **Essential**: Ent schema patterns, naming conventions, common mistakes |
| [ACTIVITIES](./guides/ACTIVITIES.md) | Activity implementation template and checklist |
| [WORKFLOWS](./guides/WORKFLOWS.md) | Temporal workflow patterns |

### Code Quality

| Document | Description |
|----------|-------------|
| [CODE_STYLE](./guides/CODE_STYLE.md) | Naming conventions, error handling, testing standards |
| [DOC_STYLE](./guides/DOC_STYLE.md) | Documentation writing guide, templates, formatting |
| [CONTRIBUTING](./CONTRIBUTING.md) | How to contribute to the project |

## 🚀 Features

### Providers

| Document | Description |
|----------|-------------|
| [GCP](./features/providers/GCP.md) | Google Cloud Platform integration |
| [GCP_ROADMAP](./features/providers/GCP_ROADMAP.md) | GCP implementation roadmap |
| [AWS](./features/providers/AWS.md) | Amazon Web Services integration |
| [GreenNode](./features/providers/GREENNODE.md) | GreenNode (formerly VNG Cloud) integration |
| [DIGITALOCEAN](./features/providers/DIGITALOCEAN.md) | DigitalOcean integration |
| [MANAGEENGINE](./features/providers/MANAGEENGINE.md) | ManageEngine Endpoint Central integration |
| [SENTINELONE](./features/providers/SENTINELONE.md) | SentinelOne integration |
| [REFERENCE](./features/providers/REFERENCE.md) | Reference data (NVD CPE) |

### Pipelines

| Document | Description |
|----------|-------------|
| [HTTPMONITOR](./features/pipelines/HTTPMONITOR.md) | HTTP traffic anomaly detection |
| [SENSITIVE_DATA_REVIEW](./features/pipelines/SENSITIVE_DATA_REVIEW.md) | Sensitive data detection and masking |

### UI

| Document | Description |
|----------|-------------|
| [ADMIN](./features/ui/ADMIN.md) | Built-in web interface for viewing data |
| [AGENT](./features/ui/AGENT.md) | AI-powered text-to-SQL query interface (WrenAI) |

## 📋 Reference

| Document | Description |
|----------|-------------|
| [GLOSSARY](./reference/GLOSSARY.md) | Terms, abbreviations, and definitions |
| [EXTERNAL_RESOURCES](./reference/EXTERNAL_RESOURCES.md) | Compliance benchmarks, CSPM tools, cloud provider docs |
| [Architectural Decisions](./decisions/) | ADRs - documents explaining important decisions (why we chose X over Y) |

## ⚙️ Setup & Operations

| Document | Description |
|----------|-------------|
| [CONFIGURATION](./setup/CONFIGURATION.md) | Config via YAML or Vault, hot-reload, validation |
| [MIGRATIONS](./setup/MIGRATIONS.md) | Database schema migrations with Atlas |
| [METABASE](./setup/METABASE.md) | Web UI setup and admin interface configuration |

## 🆘 Common Tasks

**Adding a new GCP resource type:**
1. Read [ENT_SCHEMAS](./guides/ENT_SCHEMAS.md) - Schema patterns (CRITICAL)
2. Read [ACTIVITIES](./guides/ACTIVITIES.md) - Activity template
3. Follow the checklist in ACTIVITIES.md

**Understanding the data flow:**
1. Read [OVERVIEW](./architecture/OVERVIEW.md) - Layer model
2. Read [HISTORY](./architecture/HISTORY.md) - How changes are tracked

**Architecture questions:**
- See [PRINCIPLES](./architecture/PRINCIPLES.md) - 13 core principles
- Check [decisions/](./decisions/) - Past architectural decisions
