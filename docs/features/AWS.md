# AWS

AWS resource ingestion coverage in the bronze layer.

## 🔑 IAM (`iam`)

| Resource | SDK Client | Method | Scope | Status |
|----------|-----------|--------|-------|:------:|
| Users | `iam.Client` | `ListUsers()` | Global | |
| User Policies | `iam.Client` | `ListAttachedUserPolicies()` | Global | |
| User Groups | `iam.Client` | `ListGroupsForUser()` | Global | |
| User MFA Devices | `iam.Client` | `ListMFADevices()` | Global | |
| Access Keys | `iam.Client` | `ListAccessKeys()` | Global | |
| Roles | `iam.Client` | `ListRoles()` | Global | |
| Role Policies | `iam.Client` | `ListAttachedRolePolicies()` | Global | |
| Groups | `iam.Client` | `ListGroups()` | Global | |
| Group Policies | `iam.Client` | `ListAttachedGroupPolicies()` | Global | |
| Policies | `iam.Client` | `ListPolicies()` | Global | |
| Policy Versions | `iam.Client` | `GetPolicyVersion()` | Global | |
| Instance Profiles | `iam.Client` | `ListInstanceProfiles()` | Global | |
| SAML Providers | `iam.Client` | `ListSAMLProviders()` | Global | |
| OIDC Providers | `iam.Client` | `ListOpenIDConnectProviders()` | Global | |
| Password Policy | `iam.Client` | `GetAccountPasswordPolicy()` | Global | |
| Credential Report | `iam.Client` | `GetCredentialReport()` | Global | |

## 🏢 Organizations (`organizations`)

| Resource | SDK Client | Method | Scope | Status |
|----------|-----------|--------|-------|:------:|
| Organization | `organizations.Client` | `DescribeOrganization()` | Global | |
| Accounts | `organizations.Client` | `ListAccounts()` | Global | |
| Organizational Units | `organizations.Client` | `ListOrganizationalUnitsForParent()` | Global | |
| SCPs | `organizations.Client` | `ListPolicies()` | Global | |

## 🪪 STS (`sts`)

| Resource | SDK Client | Method | Scope | Status |
|----------|-----------|--------|-------|:------:|
| Caller Identity | `sts.Client` | `GetCallerIdentity()` | Global | |

## 🖥️ EC2 (`ec2`)

### Compute

| Resource | SDK Client | Method | Scope | Status |
|----------|-----------|--------|-------|:------:|
| Instances | `ec2.Client` | `DescribeInstances()` | Regional | ✅ |
| Volumes | `ec2.Client` | `DescribeVolumes()` | Regional | |
| AMIs | `ec2.Client` | `DescribeImages()` | Regional | |
| Snapshots | `ec2.Client` | `DescribeSnapshots()` | Regional | |
| Key Pairs | `ec2.Client` | `DescribeKeyPairs()` | Regional | |
| ENIs | `ec2.Client` | `DescribeNetworkInterfaces()` | Regional | |
| EIPs | `ec2.Client` | `DescribeAddresses()` | Regional | |

### Networking (VPC)

| Resource | SDK Client | Method | Scope | Status |
|----------|-----------|--------|-------|:------:|
| VPCs | `ec2.Client` | `DescribeVpcs()` | Regional | |
| Subnets | `ec2.Client` | `DescribeSubnets()` | Regional | |
| Route Tables | `ec2.Client` | `DescribeRouteTables()` | Regional | |
| Internet Gateways | `ec2.Client` | `DescribeInternetGateways()` | Regional | |
| NAT Gateways | `ec2.Client` | `DescribeNatGateways()` | Regional | |
| NACLs | `ec2.Client` | `DescribeNetworkAcls()` | Regional | |
| Security Groups | `ec2.Client` | `DescribeSecurityGroups()` | Regional | |
| VPC Endpoints | `ec2.Client` | `DescribeVpcEndpoints()` | Regional | |
| VPC Peering | `ec2.Client` | `DescribeVpcPeeringConnections()` | Regional | |
| Flow Logs | `ec2.Client` | `DescribeFlowLogs()` | Regional | |
| Transit Gateways | `ec2.Client` | `DescribeTransitGateways()` | Regional | |
| Transit Gateway Attachments | `ec2.Client` | `DescribeTransitGatewayAttachments()` | Regional | |

## ⚡ Lambda (`lambda`)

| Resource | SDK Client | Method | Scope | Status |
|----------|-----------|--------|-------|:------:|
| Functions | `lambda.Client` | `ListFunctions()` | Regional | |
| Function URL Configs | `lambda.Client` | `GetFunctionUrlConfig()` | Regional | |
| Event Source Mappings | `lambda.Client` | `ListEventSourceMappings()` | Regional | |
| Layers | `lambda.Client` | `ListLayers()` | Regional | |

## 📦 ECS (`ecs`)

| Resource | SDK Client | Method | Scope | Status |
|----------|-----------|--------|-------|:------:|
| Clusters | `ecs.Client` | `DescribeClusters()` | Regional | |
| Services | `ecs.Client` | `DescribeServices()` | Regional | |
| Task Definitions | `ecs.Client` | `ListTaskDefinitions()` | Regional | |

## ☸️ EKS (`eks`)

| Resource | SDK Client | Method | Scope | Status |
|----------|-----------|--------|-------|:------:|
| Clusters | `eks.Client` | `DescribeCluster()` | Regional | |
| Node Groups | `eks.Client` | `ListNodegroups()` | Regional | |
| Fargate Profiles | `eks.Client` | `ListFargateProfiles()` | Regional | |

## 📈 Auto Scaling (`autoscaling`)

| Resource | SDK Client | Method | Scope | Status |
|----------|-----------|--------|-------|:------:|
| Auto Scaling Groups | `autoscaling.Client` | `DescribeAutoScalingGroups()` | Regional | |
| Launch Configurations | `autoscaling.Client` | `DescribeLaunchConfigurations()` | Regional | |
| Launch Templates | `ec2.Client` | `DescribeLaunchTemplates()` | Regional | |

## 🪣 S3 (`s3`)

| Resource | SDK Client | Method | Scope | Status |
|----------|-----------|--------|-------|:------:|
| Buckets | `s3.Client` | `ListBuckets()` | Global | |
| Bucket Policies | `s3.Client` | `GetBucketPolicy()` | Regional | |
| Bucket ACLs | `s3.Client` | `GetBucketAcl()` | Regional | |
| Bucket Encryption | `s3.Client` | `GetBucketEncryption()` | Regional | |
| Public Access Block | `s3.Client` | `GetPublicAccessBlock()` | Regional | |
| Bucket Versioning | `s3.Client` | `GetBucketVersioning()` | Regional | |
| Bucket Logging | `s3.Client` | `GetBucketLogging()` | Regional | |

## 📁 EFS (`efs`)

| Resource | SDK Client | Method | Scope | Status |
|----------|-----------|--------|-------|:------:|
| File Systems | `efs.Client` | `DescribeFileSystems()` | Regional | |

## 🔗 ELBv2 (`elasticloadbalancingv2`)

| Resource | SDK Client | Method | Scope | Status |
|----------|-----------|--------|-------|:------:|
| Load Balancers | `elasticloadbalancingv2.Client` | `DescribeLoadBalancers()` | Regional | |
| Target Groups | `elasticloadbalancingv2.Client` | `DescribeTargetGroups()` | Regional | |
| Listeners | `elasticloadbalancingv2.Client` | `DescribeListeners()` | Regional | |

## 🌐 CloudFront (`cloudfront`)

| Resource | SDK Client | Method | Scope | Status |
|----------|-----------|--------|-------|:------:|
| Distributions | `cloudfront.Client` | `ListDistributions()` | Global | |

## 🗺️ Route 53 (`route53`)

| Resource | SDK Client | Method | Scope | Status |
|----------|-----------|--------|-------|:------:|
| Hosted Zones | `route53.Client` | `ListHostedZones()` | Global | |
| Resource Record Sets | `route53.Client` | `ListResourceRecordSets()` | Global | |

## 🗄️ RDS (`rds`)

| Resource | SDK Client | Method | Scope | Status |
|----------|-----------|--------|-------|:------:|
| DB Instances | `rds.Client` | `DescribeDBInstances()` | Regional | |
| DB Clusters | `rds.Client` | `DescribeDBClusters()` | Regional | |
| DB Snapshots | `rds.Client` | `DescribeDBSnapshots()` | Regional | |
| DB Subnet Groups | `rds.Client` | `DescribeDBSubnetGroups()` | Regional | |

## 📊 DynamoDB (`dynamodb`)

| Resource | SDK Client | Method | Scope | Status |
|----------|-----------|--------|-------|:------:|
| Tables | `dynamodb.Client` | `DescribeTable()` | Regional | |

## 🧠 ElastiCache (`elasticache`)

| Resource | SDK Client | Method | Scope | Status |
|----------|-----------|--------|-------|:------:|
| Clusters | `elasticache.Client` | `DescribeCacheClusters()` | Regional | |
| Replication Groups | `elasticache.Client` | `DescribeReplicationGroups()` | Regional | |

## 📦 Redshift (`redshift`)

| Resource | SDK Client | Method | Scope | Status |
|----------|-----------|--------|-------|:------:|
| Clusters | `redshift.Client` | `DescribeClusters()` | Regional | |

## 🛡️ GuardDuty (`guardduty`)

| Resource | SDK Client | Method | Scope | Status |
|----------|-----------|--------|-------|:------:|
| Detectors | `guardduty.Client` | `ListDetectors()` | Regional | |
| Findings | `guardduty.Client` | `ListFindings()` | Regional | |

## 🔒 Security Hub (`securityhub`)

| Resource | SDK Client | Method | Scope | Status |
|----------|-----------|--------|-------|:------:|
| Hub | `securityhub.Client` | `DescribeHub()` | Regional | |
| Findings | `securityhub.Client` | `GetFindings()` | Regional | |
| Standards | `securityhub.Client` | `GetEnabledStandards()` | Regional | |

## 📋 CloudTrail (`cloudtrail`)

| Resource | SDK Client | Method | Scope | Status |
|----------|-----------|--------|-------|:------:|
| Trails | `cloudtrail.Client` | `DescribeTrails()` | Regional | |
| Trail Status | `cloudtrail.Client` | `GetTrailStatus()` | Regional | |
| Event Data Stores | `cloudtrail.Client` | `ListEventDataStores()` | Regional | |

## 🔑 KMS (`kms`)

| Resource | SDK Client | Method | Scope | Status |
|----------|-----------|--------|-------|:------:|
| Keys | `kms.Client` | `ListKeys()` | Regional | |
| Key Metadata | `kms.Client` | `DescribeKey()` | Regional | |
| Aliases | `kms.Client` | `ListAliases()` | Regional | |
| Key Policies | `kms.Client` | `GetKeyPolicy()` | Regional | |

## 🤫 Secrets Manager (`secretsmanager`)

| Resource | SDK Client | Method | Scope | Status |
|----------|-----------|--------|-------|:------:|
| Secrets | `secretsmanager.Client` | `ListSecrets()` | Regional | |

## 📜 ACM (`acm`)

| Resource | SDK Client | Method | Scope | Status |
|----------|-----------|--------|-------|:------:|
| Certificates | `acm.Client` | `ListCertificates()` | Regional | |
| Certificate Details | `acm.Client` | `DescribeCertificate()` | Regional | |

## 🛡️ WAFv2 (`wafv2`)

| Resource | SDK Client | Method | Scope | Status |
|----------|-----------|--------|-------|:------:|
| Web ACLs | `wafv2.Client` | `ListWebACLs()` | Regional | |
| IP Sets | `wafv2.Client` | `ListIPSets()` | Regional | |
| Rule Groups | `wafv2.Client` | `ListRuleGroups()` | Regional | |

## ⚙️ Config (`configservice`)

| Resource | SDK Client | Method | Scope | Status |
|----------|-----------|--------|-------|:------:|
| Configuration Recorders | `configservice.Client` | `DescribeConfigurationRecorders()` | Regional | |
| Config Rules | `configservice.Client` | `DescribeConfigRules()` | Regional | |
| Compliance | `configservice.Client` | `DescribeComplianceByConfigRule()` | Regional | |

## 🔍 IAM Access Analyzer (`accessanalyzer`)

| Resource | SDK Client | Method | Scope | Status |
|----------|-----------|--------|-------|:------:|
| Analyzers | `accessanalyzer.Client` | `ListAnalyzers()` | Regional | |
| Findings | `accessanalyzer.Client` | `ListFindings()` | Regional | |

## 📊 CloudWatch (`cloudwatch`)

| Resource | SDK Client | Method | Scope | Status |
|----------|-----------|--------|-------|:------:|
| Alarms | `cloudwatch.Client` | `DescribeAlarms()` | Regional | |
| Log Groups | `cloudwatchlogs.Client` | `DescribeLogGroups()` | Regional | |
| Metric Filters | `cloudwatchlogs.Client` | `DescribeMetricFilters()` | Regional | |

## 📬 SNS (`sns`)

| Resource | SDK Client | Method | Scope | Status |
|----------|-----------|--------|-------|:------:|
| Topics | `sns.Client` | `ListTopics()` | Regional | |
| Subscriptions | `sns.Client` | `ListSubscriptions()` | Regional | |
| Topic Attributes | `sns.Client` | `GetTopicAttributes()` | Regional | |

## 📥 SQS (`sqs`)

| Resource | SDK Client | Method | Scope | Status |
|----------|-----------|--------|-------|:------:|
| Queues | `sqs.Client` | `ListQueues()` | Regional | |
| Queue Attributes | `sqs.Client` | `GetQueueAttributes()` | Regional | |

## 🔧 SSM (`ssm`)

| Resource | SDK Client | Method | Scope | Status |
|----------|-----------|--------|-------|:------:|
| Parameters | `ssm.Client` | `DescribeParameters()` | Regional | |
| Managed Instances | `ssm.Client` | `DescribeInstanceInformation()` | Regional | |

## 🚀 API Gateway (`apigateway` / `apigatewayv2`)

| Resource | SDK Client | Method | Scope | Status |
|----------|-----------|--------|-------|:------:|
| REST APIs | `apigateway.Client` | `GetRestApis()` | Regional | |
| HTTP APIs | `apigatewayv2.Client` | `GetApis()` | Regional | |
| Stages | `apigateway.Client` | `GetStages()` | Regional | |

## 📊 Summary

**Total: 1/138 (1%)**

| Service | Implemented | Total |
|---------|:-----------:|:-----:|
| IAM | 0 | 16 |
| Organizations | 0 | 4 |
| STS | 0 | 1 |
| EC2 (Compute) | 1 | 7 |
| EC2 (Networking) | 0 | 12 |
| Lambda | 0 | 4 |
| ECS | 0 | 3 |
| EKS | 0 | 3 |
| Auto Scaling | 0 | 3 |
| S3 | 0 | 7 |
| EFS | 0 | 1 |
| ELBv2 | 0 | 3 |
| CloudFront | 0 | 1 |
| Route 53 | 0 | 2 |
| RDS | 0 | 4 |
| DynamoDB | 0 | 1 |
| ElastiCache | 0 | 2 |
| Redshift | 0 | 1 |
| GuardDuty | 0 | 2 |
| Security Hub | 0 | 3 |
| CloudTrail | 0 | 3 |
| KMS | 0 | 4 |
| Secrets Manager | 0 | 1 |
| ACM | 0 | 2 |
| WAFv2 | 0 | 3 |
| Config | 0 | 3 |
| Access Analyzer | 0 | 2 |
| CloudWatch | 0 | 3 |
| SNS | 0 | 3 |
| SQS | 0 | 2 |
| SSM | 0 | 2 |
| API Gateway | 0 | 3 |

See [EXTERNAL_RESOURCES.md](../reference/EXTERNAL_RESOURCES.md) for compliance benchmarks, open source tools, and cloud provider documentation.
