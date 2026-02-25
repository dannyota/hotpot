# Glossary

## 🏗️ Architecture Terms

| Term | Definition |
|------|------------|
| **Medallion Architecture** | Data pipeline pattern with Bronze→Silver→Gold layers |
| **Bronze** | Raw data layer, stores API responses as-is |
| **Silver** | Normalized layer, unified data models |
| **Gold** | Analytics layer, alerts and compliance results |
| **Layer** | One stage in the medallion pipeline (ingest/normalize/detect) |
| **Provider** | External data source (GCP, GreenNode, SentinelOne, etc.) |

## 📦 Package Names

| Package | Layer | Purpose |
|---------|-------|---------|
| `ingest` | Bronze | Collect data from external APIs |
| `normalize` | Silver | Transform and unify data models |
| `detect` | Gold | Run detection rules, generate alerts |
| `admin` | UI | Web interface, dashboards |
| `base` | Shared | Utilities and models used by all layers |
| Agent | External | Text-to-SQL (WrenAI + Ollama / Vertex AI) |

## 🌐 External Sources

| Abbreviation | Full Name | Type |
|--------------|-----------|------|
| **GCP** | Google Cloud Platform | Cloud inventory |
| **GreenNode** | GreenNode (formerly VNG Cloud) | Cloud inventory |
| **S1** | SentinelOne | EDR/Endpoint security |
| **DO** | DigitalOcean | Cloud inventory |
| **Vault** | HashiCorp Vault | Secrets/PKI management |
| **AWS** | Amazon Web Services | Cloud inventory |
| **SCC** | Security Command Center (GCP) | Vulnerability scanner |

## ⚙️ Tech Stack

| Term | Definition |
|------|------------|
| **Temporal** | Workflow orchestration engine |
| **Activity** | Single unit of work in Temporal |
| **Workflow** | Orchestrates multiple activities |
| **Ent** | Type-safe Go entity framework |
| **dialect.Driver** | Shared database driver, used to create per-service ent clients |

## 🗄️ Database

| Term | Definition |
|------|------------|
| **Schema** | PostgreSQL schema (bronze/silver/gold) |
| **Annotations()** | Ent method to specify table name and indexes |

## 🔧 Code Patterns

| Term | Definition |
|------|------------|
| **client.go** | External API client wrapper |
| **activities.go** | Temporal activity definitions |
| **workflows.go** | Temporal workflow definitions |
| **provider.go** | Provider self-registration via init() |
| **run.go** | Module entry point (`Run()` function) |

## 🤖 Agent Terms

| Term | Definition |
|------|------------|
| **WrenAI** | Open-source text-to-SQL engine with semantic layer |
| **Ollama** | Local LLM runtime (OpenAI-compatible API) |
| **Vertex AI** | Google Cloud managed LLM (Gemini models) |
| **Text-to-SQL** | Natural language → SQL query |
| **LLM** | Large Language Model (Qwen, Llama, Gemini) |
