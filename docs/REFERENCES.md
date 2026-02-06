# References

Tools, benchmarks, and documentation for building cloud security data pipelines. Review these to understand what others do and how they do it.

## Compliance Benchmarks

| Benchmark | Clouds | What to study |
|-----------|--------|---------------|
| [CIS Google Cloud Platform Foundation v4.0.0](https://www.cisecurity.org/benchmark/google_cloud_computing_platform) | GCP | IAM, logging, networking, VMs, storage, SQL, BigQuery, Dataproc |
| [CIS Amazon Web Services Foundations](https://www.cisecurity.org/benchmark/amazon_web_services) | AWS | IAM, logging, monitoring, networking, storage |
| [CIS Microsoft Azure Foundations](https://www.cisecurity.org/benchmark/microsoft_azure) | Azure | IAM, Security Center, storage, SQL, logging, networking |
| [CIS Google Kubernetes Engine (GKE)](https://www.cisecurity.org/benchmark/kubernetes) | GCP | GKE-specific security controls |
| [CIS Amazon Elastic Kubernetes Service (EKS)](https://www.cisecurity.org/benchmark/kubernetes) | AWS | EKS-specific security controls |
| [CIS Kubernetes](https://www.cisecurity.org/benchmark/kubernetes) | Any | General Kubernetes hardening |

## Open Source CSPM / Auditing Tools

### Multi-Cloud

| Tool | Languages | Clouds | What to study |
|------|-----------|--------|---------------|
| [prowler-cloud/prowler](https://github.com/prowler-cloud/prowler) | Python | AWS, Azure, GCP, K8s | Detection logic for CIS, NIST 800-53, PCI-DSS, HIPAA, GDPR; check structure and framework mapping |
| [cloudquery/cloudquery](https://github.com/cloudquery/cloudquery) | Go | 70+ sources | Data pipeline architecture, resource normalization, SQL table schemas per cloud provider |
| [turbot/steampipe](https://github.com/turbot/steampipe) | Go | AWS, Azure, GCP, 140+ | SQL-over-APIs pattern, plugin architecture for cloud resource querying |
| [nccgroup/ScoutSuite](https://github.com/nccgroup/ScoutSuite) | Python | AWS, Azure, GCP | API-based config collection, rule engine, report generation |
| [aquasecurity/cloudsploit](https://github.com/aquasecurity/cloudsploit) | JavaScript | AWS, Azure, GCP, OCI | Lightweight check implementations, plugin pattern |

### GCP-Specific

| Tool | What to study |
|------|---------------|
| [GoogleCloudPlatform/inspec-gcp-cis-benchmark](https://github.com/GoogleCloudPlatform/inspec-gcp-cis-benchmark) | Official Google CIS 4.0 InSpec profile — maps each control to exact API calls and fields |
| [turbot/steampipe-plugin-gcp](https://github.com/turbot/steampipe-plugin-gcp) | GCP API client patterns, data schemas, resource relationships |
| [turbot/steampipe-mod-gcp-compliance](https://github.com/turbot/steampipe-mod-gcp-compliance) | SQL compliance queries for CIS v1.2–v4.0, HIPAA, PCI DSS, NIST 800-53, NIST CSF |

### AWS-Specific

| Tool | What to study |
|------|---------------|
| [turbot/steampipe-plugin-aws](https://github.com/turbot/steampipe-plugin-aws) | AWS API client patterns and data schemas |
| [turbot/steampipe-mod-aws-compliance](https://github.com/turbot/steampipe-mod-aws-compliance) | SQL compliance queries for CIS, PCI DSS, HIPAA, NIST, FedRAMP, SOC 2 |
| [toniblyx/prowler (AWS checks)](https://github.com/prowler-cloud/prowler/tree/master/prowler/providers/aws) | AWS-specific detection logic |

### Azure-Specific

| Tool | What to study |
|------|---------------|
| [turbot/steampipe-plugin-azure](https://github.com/turbot/steampipe-plugin-azure) | Azure API client patterns and data schemas |
| [turbot/steampipe-mod-azure-compliance](https://github.com/turbot/steampipe-mod-azure-compliance) | SQL compliance queries for CIS, PCI DSS, HIPAA, NIST |

## Data Pipeline References

| Project | What to study |
|---------|---------------|
| [cloudquery/cloudquery](https://github.com/cloudquery/cloudquery) | Bronze-layer pattern: extract cloud resources into SQL tables, handles pagination, rate limiting, multi-account |
| [cloudquery/plugin-sdk](https://github.com/cloudquery/plugin-sdk) | Plugin SDK for writing source/destination plugins, resource table definitions |
| [turbot/steampipe-plugin-sdk](https://github.com/turbot/steampipe-plugin-sdk) | Go SDK for building cloud API plugins with caching, hydrate functions, column definitions |

## Cloud Provider Documentation

### GCP

- [Cloud Asset Inventory](https://cloud.google.com/asset-inventory/docs) — authoritative resource enumeration across projects/folders/orgs
- [Cloud Audit Logs](https://cloud.google.com/logging/docs/audit) — admin activity, data access, system event, policy denied logs
- [Security Command Center](https://cloud.google.com/security-command-center/docs) — aggregated findings, vulnerability scanning, compliance posture
- [VPC Service Controls](https://cloud.google.com/vpc-service-controls/docs) — data exfiltration prevention, service perimeters
- [Organization Policy](https://cloud.google.com/resource-manager/docs/organization-policy/overview) — org-level guardrails and constraints
- [Security Best Practices](https://cloud.google.com/security/best-practices) — IAM, encryption, network security, compliance blueprints
- [GCP API Discovery](https://cloud.google.com/apis/docs/overview) — full list of GCP APIs and their resource types

### AWS

- [AWS Config](https://docs.aws.amazon.com/config/) — resource inventory, configuration history, compliance rules
- [AWS Security Hub](https://docs.aws.amazon.com/securityhub/) — aggregated findings, CIS/PCI/NIST compliance checks
- [AWS CloudTrail](https://docs.aws.amazon.com/cloudtrail/) — API activity logs
- [AWS IAM Access Analyzer](https://docs.aws.amazon.com/IAM/latest/UserGuide/what-is-access-analyzer.html) — external access findings

### Azure

- [Microsoft Defender for Cloud](https://learn.microsoft.com/en-us/azure/defender-for-cloud/) — CSPM, compliance assessments, security recommendations
- [Azure Resource Graph](https://learn.microsoft.com/en-us/azure/governance/resource-graph/) — resource querying at scale
- [Azure Activity Log](https://learn.microsoft.com/en-us/azure/azure-monitor/essentials/activity-log) — subscription-level events
- [Azure Policy](https://learn.microsoft.com/en-us/azure/governance/policy/) — policy-as-code enforcement

## Compliance Frameworks

| Framework | Focus | Key resource |
|-----------|-------|-------------|
| SOC 2 | Trust Service Criteria (security, availability, processing integrity, confidentiality, privacy) | [AICPA SOC 2](https://www.aicpa.org/soc2) |
| PCI DSS v4.0 | Payment card data security | [PCI SSC](https://www.pcisecuritystandards.org/) |
| HIPAA | Protected health information | [HHS HIPAA](https://www.hhs.gov/hipaa/) |
| NIST 800-53 Rev 5 | Federal information systems | [NIST SP 800-53](https://csrc.nist.gov/publications/detail/sp/800-53/rev-5/final) |
| NIST CSF 2.0 | Cybersecurity risk management | [NIST CSF](https://www.nist.gov/cyberframework) |
| ISO 27001/27017/27018 | Information security / cloud / PII | [ISO 27001](https://www.iso.org/standard/27001) |
| GDPR | EU data protection | [GDPR text](https://gdpr-info.eu/) |
| FedRAMP | Federal cloud authorization | [FedRAMP](https://www.fedramp.gov/) |
| CSA CCM v4 | Cloud Controls Matrix | [CSA CCM](https://cloudsecurityalliance.org/research/cloud-controls-matrix/) |

## Archived (read for patterns only)

- [forseti-security/forseti-security](https://github.com/forseti-security/forseti-security) — archived Jan 2025, was Google's standard GCP security tool; useful for understanding inventory snapshot and policy scanning patterns
