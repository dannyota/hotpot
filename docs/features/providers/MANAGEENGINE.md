# ManageEngine Endpoint Central

ManageEngine Endpoint Central API resource ingestion coverage in the bronze layer.

## 🖥️ API v1.4 (`/api/1.4/`)

### SoM (Scope of Management)

| Resource | Endpoint | Status |
|----------|----------|:------:|
| Summary | `/som/summary` | |
| Computers | `/som/computers` | |
| Remote Offices | `/som/remoteoffice` | |

### Inventory

| Resource | Endpoint | Status |
|----------|----------|:------:|
| Summary | `/inventory/allsummary` | |
| Scan Computers | `/inventory/scancomputers` | |
| Software List | `/inventory/software` | |
| Hardware List | `/inventory/hardware` | |
| Computer Detail Summary | `/inventory/compdetailssummary` | |
| Installed Software (per Computer) | `/inventory/installedsoftware` | |
| Software Metering Rules | `/inventory/swmeteringsummary` | |
| Licensed Software | `/inventory/licensesoftware` | |
| Licenses (per Software) | `/inventory/licenses` | |
| Filter Parameters | `/inventory/filterparams` | |

### Patch Management

| Resource | Endpoint | Status |
|----------|----------|:------:|
| All Patches | `/patch/allpatches` | |
| Patch Summary | `/patch/patchsummary` | |
| Systems with Patch Status | `/patch/allsystems` | |
| Patch Status Across Computers | `/patch/patchstatus` | |
| Patch Report for System | `/patch/systemreport` | |
| Systems and Patch Details | `/patch/allsystemspatches` | |
| Downloaded Patches | `/patch/downloadedpatches` | |
| Supported Patches | `/patch/supportedpatches` | |
| System Health Policy | `/patch/systemhealthpolicy` | |
| Deployment Policies | `/patch/deploypolicy` | |
| Deployment Configurations | `/patch/deployconfigs` | |
| Patch Approval Settings | `/patch/approvalsettings` | |
| Patch Database Update Status | `/patch/dbupdatestatus` | |
| Patch Scan System List | `/patch/patchscansystems` | |

### Vulnerability Management

| Resource | Endpoint | Status |
|----------|----------|:------:|
| Vulnerabilities | `/dcapi/threats/vulnerabilities` | |
| Vulnerable Computers | `/dcapi/threats/vulnerablecomputers` | |
| Server Misconfigurations | `/dcapi/threats/servermisconfigs` | |
| System Misconfigurations | `/dcapi/threats/systemmisconfigs` | |
| Computers with Server Misconfigs | `/dcapi/threats/servermisconfigcomputers` | |
| Threat and Patch Summary | `/dcapi/threats/threatsummary` | |
| Vulnerability-Computer Details | `/dcapi/threats/vulncomputers` | |

### Device Control Reports

| Resource | Endpoint | Status |
|----------|----------|:------:|
| Device Summary | `/dcapi/devicecontrol/summary` | |
| Device Activity Audit | `/dcapi/devicecontrol/activityaudit` | |
| Device Status per Computer | `/dcapi/devicecontrol/statuspercomputer` | |
| Device Status per Mac | `/dcapi/devicecontrol/statuspermac` | |
| Blocked Devices Audit | `/dcapi/devicecontrol/blockeddevices` | |
| File Activity Trace | `/dcapi/devicecontrol/fileactivity` | |
| File Shadow Operations | `/dcapi/devicecontrol/fileshadow` | |
| Temporary Device Exemptions | `/dcapi/devicecontrol/tempexemptions` | |
| Device Type Exemptions | `/dcapi/devicecontrol/typeexemptions` | |

### Data Loss Prevention Reports

| Resource | Endpoint | Status |
|----------|----------|:------:|
| Deployed Rules | `/dcapi/dlp/rules` | |
| Deployed Applications | `/dcapi/dlp/applications` | |
| Deployed Devices | `/dcapi/dlp/devices` | |
| Deployed Web Domains | `/dcapi/dlp/webdomains` | |
| Deployed Email Domains | `/dcapi/dlp/emaildomains` | |
| Deployed Network Printers | `/dcapi/dlp/networkprinters` | |
| Deployed USB Printers | `/dcapi/dlp/usbprinters` | |
| Endpoint Activity Report | `/dcapi/dlp/endpointactivity` | |
| Justification Summary | `/dcapi/dlp/justifications` | |
| Enterprise False Positives | `/dcapi/dlp/falsepositivesreport` | |

### BitLocker

| Resource | Endpoint | Status |
|----------|----------|:------:|
| BitLocker Drive Report | `/dcapi/bitlocker/drivereport` | |
| TPM Report | `/dcapi/bitlocker/tpmreport` | |
| Recovery Key Details | `/dcapi/bitlocker/recoverykeys` | |

### Custom Fields

| Resource | Endpoint | Status |
|----------|----------|:------:|
| Custom Fields | `/dcapi/customfield` | |
| Custom Data Types | `/dcapi/customfield/datatypes` | |
| Custom Data Values | `/dcapi/customfield/values` | |

### Reports

| Resource | Endpoint | Status |
|----------|----------|:------:|
| Query Reports | `/dcapi/reports/query` | |
| Custom Reports | `/dcapi/reports/custom` | |

### Common

| Resource | Endpoint | Status |
|----------|----------|:------:|
| Custom Groups | `/dcapi/common/customgroups` | |
| Server Properties | `/dcapi/common/serverproperties` | |

## 📊 Summary

**Total: 0/63 (0%)**

| API | Implemented | Total |
|-----|:-----------:|:-----:|
| SoM | 0 | 3 |
| Inventory | 0 | 10 |
| Patch Management | 0 | 14 |
| Vulnerability Management | 0 | 7 |
| Device Control Reports | 0 | 9 |
| Data Loss Prevention Reports | 0 | 10 |
| BitLocker | 0 | 3 |
| Custom Fields | 0 | 3 |
| Reports | 0 | 2 |
| Common | 0 | 2 |

See [EXTERNAL_RESOURCES.md](../reference/EXTERNAL_RESOURCES.md) for compliance benchmarks, open source tools, and cloud provider documentation.
