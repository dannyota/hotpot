# Reference Data

Public reference datasets ingested into the bronze layer for software categorization and matching.

## NVD CPE Dictionary (`nvd.nist.gov`)

| Resource | Source | Format | Status |
|----------|--------|--------|:------:|
| CPE Products (App + OS) | `/feeds/json/cpe/2.0/nvdcpe-2.0.tar.gz` | tar.gz of JSON chunks | ✅ |

Fields: part (a=App, o=OS), vendor, product, version, title, deprecated flag.
~1.5M entries (hardware filtered out). Updated daily by NVD.

## Ubuntu Packages (`archive.ubuntu.com`)

| Resource | Source | Format | Status |
|----------|--------|--------|:------:|
| Noble main | `/ubuntu/dists/noble/main/binary-amd64/Packages.gz` | gzip text | ✅ |
| Noble universe | `/ubuntu/dists/noble/universe/binary-amd64/Packages.gz` | gzip text | ✅ |
| Jammy main | `/ubuntu/dists/jammy/main/binary-amd64/Packages.gz` | gzip text | ✅ |
| Jammy universe | `/ubuntu/dists/jammy/universe/binary-amd64/Packages.gz` | gzip text | ✅ |

Fields: package name, release, component, section, description.
Sections: admin, utils, web, net, libs, devel, database, editors, shells, etc.

## RHEL 9 Packages (`mirror.stream.centos.org` / `dl.fedoraproject.org`)

| Resource | Source | Format | Status |
|----------|--------|--------|:------:|
| RHEL 9 BaseOS | CentOS Stream 9 BaseOS `primary.xml.gz` | gzip XML | ✅ |
| RHEL 9 AppStream | CentOS Stream 9 AppStream `primary.xml.gz` | gzip XML | ✅ |
| EPEL 9 | EPEL 9 `primary.xml.xz` | xz XML | ✅ |

Fields: package name, repo, arch, version, rpm_group, summary, url.

## RHEL 7 / CentOS 7 Packages (`vault.centos.org` / `dl.fedoraproject.org`)

| Resource | Source | Format | Status |
|----------|--------|--------|:------:|
| CentOS 7 OS | CentOS 7.9.2009 vault OS `primary.xml.gz` | gzip XML | ✅ |
| CentOS 7 Updates | CentOS 7.9.2009 vault Updates `primary.xml.gz` | gzip XML | ✅ |
| EPEL 7 | EPEL 7 archive `primary.xml.xz` | xz XML | ✅ |

Fields: package name, repo, arch, version, rpm_group, summary, url.
CentOS 7 reached EOL 2024-06-30; repos are archived at `vault.centos.org`.

## Summary

| Source | Resources | Total |
|--------|:---------:|:-----:|
| NVD CPE | 1 | 1 |
| Ubuntu | 4 | 4 |
| RHEL 9 | 3 | 3 |
| RHEL 7 | 3 | 3 |
