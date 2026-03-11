package config

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// SeedURIAttackPatterns inserts predefined URI attack detection patterns.
// Patterns are derived from OWASP ModSecurity CRS for traceability.
func SeedURIAttackPatterns(ctx context.Context, db *sql.DB) error {
	if len(uriAttackPatterns) == 0 {
		return nil
	}

	now := time.Now()
	var b strings.Builder
	b.WriteString(`INSERT INTO config.uri_attack_patterns
		(pattern_type, match_mode, pattern, match_target, crs_rule_id, description, is_active, created_at, updated_at)
		VALUES `)

	args := make([]any, 0, len(uriAttackPatterns)*9)
	for i, p := range uriAttackPatterns {
		if i > 0 {
			b.WriteString(", ")
		}
		base := i * 9
		fmt.Fprintf(&b, "($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)",
			base+1, base+2, base+3, base+4, base+5, base+6, base+7, base+8, base+9)

		var matchTarget, crsRuleID *string
		if p.matchTarget != "" {
			matchTarget = &p.matchTarget
		}
		if p.crsRuleID != "" {
			crsRuleID = &p.crsRuleID
		}

		args = append(args, p.patternType, p.matchMode, p.pattern,
			matchTarget, crsRuleID, p.description, true, now, now)
	}

	b.WriteString(` ON CONFLICT (pattern_type, pattern) DO NOTHING`)

	_, err := db.ExecContext(ctx, b.String(), args...)
	if err != nil {
		return fmt.Errorf("upsert uri attack patterns (%d entries): %w", len(uriAttackPatterns), err)
	}
	return nil
}

type uriAttackEntry struct {
	patternType string // "lfi", "sqli", "rce", "xss", "ssrf"
	matchMode   string // "substring", "regex"
	pattern     string
	matchTarget string // "" = full URI
	crsRuleID   string // OWASP CRS rule ID, "" = custom
	description string
}

// Patterns derived from OWASP ModSecurity CRS v4 (Apache 2.0 license).
// Only URI-applicable patterns are included — body/header patterns are skipped.
var uriAttackPatterns = []uriAttackEntry{
	// =========================================================================
	// LFI / Path Traversal (CRS 930)
	// =========================================================================

	// --- Directory traversal: literal and encoded variants ---
	{"lfi", "substring", "/../", "", "930110", "Directory traversal (literal)"},
	{"lfi", "substring", "\\..\\", "", "930110", "Directory traversal (backslash)"},
	{"lfi", "substring", "/..;/", "", "930110", "Directory traversal (semicolon bypass — Tomcat/Jetty)"},
	{"lfi", "regex", `(?:%2e|\.){2}[/\\%]`, "", "930100", "Encoded dot-dot with separator"},
	{"lfi", "regex", `%2e%2e[%2f%5c/\\]`, "", "930100", "URL-encoded ../ or ..\\"},
	{"lfi", "regex", `%252e%252e%252f`, "", "930100", "Double-encoded ../"},
	{"lfi", "regex", `%c0%ae%c0%ae[%c0%af/\\]`, "", "930100", "Overlong UTF-8 encoded ../"},
	{"lfi", "substring", "%00", "", "930100", "Null byte injection in path"},

	// --- Sensitive Unix files ---
	{"lfi", "substring", "/etc/passwd", "", "930120", "Unix password file"},
	{"lfi", "substring", "/etc/shadow", "", "930120", "Unix shadow password file"},
	{"lfi", "substring", "/proc/self/", "", "930120", "Linux proc self (environ, cmdline, fd)"},

	// --- Sensitive web/app files ---
	{"lfi", "substring", "/.env", "", "930130", "Dotenv config file"},
	{"lfi", "substring", "/.git/", "", "930130", "Git repository metadata"},
	{"lfi", "substring", "/.htaccess", "", "930130", "Apache access control file"},
	{"lfi", "substring", "/.htpasswd", "", "930130", "Apache password file"},
	{"lfi", "substring", "/wp-config.php", "", "930130", "WordPress database credentials"},
	{"lfi", "substring", "/web.config", "", "930130", "ASP.NET web configuration"},
	{"lfi", "substring", "/.aws/", "", "930130", "AWS credentials directory"},
	{"lfi", "substring", "/.ssh/", "", "930130", "SSH keys and known_hosts"},
	{"lfi", "substring", "/.kube/", "", "930130", "Kubernetes config directory"},
	{"lfi", "substring", "/.docker/", "", "930130", "Docker config directory"},

	// --- Windows sensitive files ---
	{"lfi", "substring", "boot.ini", "", "930120", "Windows boot configuration"},
	{"lfi", "substring", "win.ini", "", "930120", "Windows system configuration"},

	// --- Protocol wrappers ---
	{"lfi", "regex", `(?i)(?:php|phar|zip|data|expect|glob)://`, "", "930120", "PHP stream wrapper / protocol handler"},
	{"lfi", "regex", `(?i)file://`, "", "930120", "file:// protocol access"},

	// =========================================================================
	// SQL Injection (CRS 942)
	// =========================================================================

	// --- Boolean-based blind ---
	{"sqli", "regex", `(?i)['"]\\s*(?:OR|AND)\s+['"]?\d+['"]?\s*[=<>]\s*['"]?\d+`, "", "942180", "Boolean-based SQLi: OR/AND digit=digit with quotes"},
	{"sqli", "regex", `(?i)['"]\\s*(?:OR|AND)\s+(?:true|false|null)\b`, "", "942180", "Boolean-based SQLi: OR/AND true/false"},
	{"sqli", "regex", `(?i)(?:OR|AND)\s+\d+\s*=\s*\d+\s*(?:--|#|/\*)`, "", "942210", "Boolean tautology with comment terminator"},

	// --- UNION-based ---
	{"sqli", "regex", `(?i)UNION\s+(?:ALL\s+)?SELECT\b`, "", "942270", "UNION SELECT injection"},

	// --- Stacked queries ---
	{"sqli", "regex", `(?i);\s*(?:DROP|DELETE|INSERT|UPDATE|ALTER|CREATE|TRUNCATE)\s`, "", "942350", "Stacked query: DDL/DML after semicolon"},
	{"sqli", "regex", `(?i);\s*(?:EXEC|EXECUTE)\s`, "", "942190", "Stacked query: EXEC after semicolon"},

	// --- Time-based blind ---
	{"sqli", "regex", `(?i)(?:SLEEP|BENCHMARK|PG_SLEEP|WAITFOR\s+DELAY)\s*\(`, "", "942160", "Time-based blind: SLEEP/BENCHMARK/PG_SLEEP/WAITFOR"},

	// --- Error-based ---
	{"sqli", "regex", `(?i)(?:EXTRACTVALUE|UPDATEXML|XMLTYPE)\s*\(`, "", "942151", "Error-based: XML extraction functions"},

	// --- Common SQL functions in attack context ---
	{"sqli", "regex", `(?i)(?:CONCAT|GROUP_CONCAT|CHAR|CHR|SUBSTR|SUBSTRING|ASCII|ORD|HEX|UNHEX)\s*\(`, "", "942151", "SQL string/encoding functions in URI"},

	// --- Comment injection ---
	{"sqli", "substring", "/**/", "", "942440", "SQL inline comment (obfuscation)"},
	{"sqli", "regex", `(?i)['"]\s*(?:--|#)`, "", "942440", "SQL comment after quote (line terminator)"},

	// --- Encoded quotes with SQL keywords ---
	{"sqli", "regex", `(?i)%27\s*(?:OR|AND|UNION|SELECT)`, "", "942180", "URL-encoded single quote with SQL keyword"},
	{"sqli", "regex", `(?i)%22\s*(?:OR|AND|UNION|SELECT)`, "", "942180", "URL-encoded double quote with SQL keyword"},

	// --- Information schema / system tables ---
	{"sqli", "substring", "information_schema", "", "942140", "Access to information_schema"},
	{"sqli", "substring", "pg_catalog", "", "942140", "Access to PostgreSQL system catalog"},

	// --- MySQL comment execution ---
	{"sqli", "regex", `(?i)/\*!\d*\s`, "", "942500", "MySQL versioned comment execution"},

	// --- xp_cmdshell ---
	{"sqli", "regex", `(?i)xp_cmdshell\s*\(`, "", "942190", "MSSQL xp_cmdshell command execution"},

	// =========================================================================
	// Command Injection / RCE (CRS 932)
	// =========================================================================

	// --- Shell metacharacters with command context ---
	{"rce", "regex", `(?i)[;|` + "`" + `]\s*(?:cat|ls|id|whoami|uname|pwd|curl|wget|nc|ncat|bash|sh|python|perl|ruby|php|chmod|chown|chgrp|kill|rm|mv|cp|find|grep|awk|sed|env|printenv|hostname|ifconfig|netstat|ss|ping|traceroute|dig|nslookup)\b`, "", "932235", "Shell operator followed by Unix command"},
	{"rce", "regex", `(?i)(?:&&|\|\|)\s*(?:cat|ls|id|whoami|uname|curl|wget|nc|bash|sh|python|perl|ruby|php|chmod|chown|env)\b`, "", "932235", "Chained shell operator with Unix command"},

	// --- Shell command substitution ---
	{"rce", "regex", `\$\(\s*(?:cat|ls|id|whoami|uname|curl|wget|bash|sh|python|perl)\b`, "", "932130", "Command substitution $() with Unix command"},
	{"rce", "substring", "${IFS}", "", "932130", "Shell variable IFS (field separator bypass)"},

	// --- Direct Unix shell paths ---
	{"rce", "substring", "/bin/sh", "", "932160", "Direct path to sh shell"},
	{"rce", "substring", "/bin/bash", "", "932160", "Direct path to bash shell"},
	{"rce", "substring", "/dev/tcp/", "", "932160", "Bash /dev/tcp reverse shell"},
	{"rce", "substring", "/dev/udp/", "", "932160", "Bash /dev/udp reverse shell"},

	// --- Windows command injection ---
	{"rce", "regex", `(?i)cmd(?:\.exe)?\s*/c\b`, "", "932370", "Windows cmd.exe /c execution"},
	{"rce", "regex", `(?i)powershell(?:\.exe)?\s`, "", "932120", "Windows PowerShell execution"},

	// --- Encoded shell metacharacters ---
	{"rce", "regex", `(?i)%7c\s*(?:cat|ls|id|whoami|curl|wget|bash|sh)\b`, "", "932235", "URL-encoded pipe with Unix command"},
	{"rce", "regex", `(?i)%3b\s*(?:cat|ls|id|whoami|curl|wget|bash|sh|rm|chmod)\b`, "", "932235", "URL-encoded semicolon with Unix command"},

	// --- Shellshock ---
	{"rce", "substring", "() {", "", "932170", "Shellshock (CVE-2014-6271) function definition"},

	// =========================================================================
	// XSS (CRS 941)
	// =========================================================================

	// --- Script tags ---
	{"xss", "regex", `(?i)<\s*/?\s*script\b`, "", "941110", "Script tag injection"},
	{"xss", "regex", `(?i)%3c\s*/?\s*script\b`, "", "941110", "URL-encoded script tag"},

	// --- Event handlers ---
	{"xss", "regex", `(?i)\bon(?:error|load|click|mouseover|focus|blur|submit|change|input|keydown|keyup)\s*=`, "", "941120", "HTML event handler attribute injection"},

	// --- JavaScript URIs ---
	{"xss", "regex", `(?i)(?:java|vb|live)script\s*:`, "", "941170", "JavaScript/VBScript URI scheme"},

	// --- HTML tag injection with dangerous attributes ---
	{"xss", "regex", `(?i)<\s*(?:img|svg|iframe|object|embed|video|audio|body|input|details|marquee|form|math|meta|link|base)\b[^>]*\bon\w+\s*=`, "", "941160", "HTML tag with event handler"},
	{"xss", "regex", `(?i)%3c\s*(?:img|svg|iframe|object|embed)\b`, "", "941160", "URL-encoded HTML tag injection"},

	// --- Encoded angle brackets ---
	{"xss", "regex", `(?i)(?:%3c|&#0*60;?|&#x0*3c;?)\s*(?:script|img|svg|iframe|object|embed)\b`, "", "941160", "HTML-entity/URL-encoded tag injection"},

	// --- Expression injection ---
	{"xss", "regex", `(?i)expression\s*\(`, "", "941140", "CSS expression() injection"},
	{"xss", "regex", `(?i)url\s*\(\s*['"]?\s*javascript:`, "", "941140", "CSS url(javascript:) injection"},

	// --- Dangerous JS functions ---
	{"xss", "regex", `(?i)(?:eval|alert|prompt|confirm)\s*\(`, "", "941390", "Dangerous JavaScript function call in URI"},

	// =========================================================================
	// SSRF (CRS 934 + custom)
	// =========================================================================

	// --- Cloud metadata endpoints ---
	{"ssrf", "substring", "169.254.169.254", "", "934110", "AWS/Azure/GCP metadata IP"},
	{"ssrf", "substring", "metadata.google.internal", "", "934110", "GCP metadata hostname"},
	{"ssrf", "substring", "100.100.100.200", "", "934110", "Alibaba Cloud metadata IP"},

	// --- Cloud metadata paths ---
	{"ssrf", "substring", "/latest/meta-data/", "", "934110", "AWS EC2 metadata path"},
	{"ssrf", "substring", "/computeMetadata/v1/", "", "934110", "GCP metadata path"},
	{"ssrf", "substring", "/metadata/v1/", "", "934110", "Azure/DO metadata path"},
	{"ssrf", "substring", "/metadata/instance", "", "934110", "Azure instance metadata path"},

	// --- Loopback and localhost ---
	{"ssrf", "regex", `(?i)(?:https?://|//)(?:127\.0\.0\.1|localhost|\[::1\]|0\.0\.0\.0)(?:[:/]|$)`, "", "934120", "Loopback address in URL parameter"},

	// --- Loopback encoding evasion ---
	{"ssrf", "substring", "2130706433", "", "934120", "Decimal-encoded 127.0.0.1"},
	{"ssrf", "substring", "0x7f000001", "", "934120", "Hex-encoded 127.0.0.1"},
	{"ssrf", "substring", "0177.0.0.1", "", "934120", "Octal-encoded 127.0.0.1"},

	// --- Metadata IP encoding evasion ---
	{"ssrf", "substring", "2852039166", "", "934120", "Decimal-encoded 169.254.169.254"},
	{"ssrf", "substring", "0xA9FEA9FE", "", "934120", "Hex-encoded 169.254.169.254"},

	// --- Private IP ranges in URL parameters ---
	{"ssrf", "regex", `(?i)(?:https?://|//)(?:10\.\d{1,3}\.\d{1,3}\.\d{1,3}|172\.(?:1[6-9]|2\d|3[01])\.\d{1,3}\.\d{1,3}|192\.168\.\d{1,3}\.\d{1,3})(?:[:/]|$)`, "", "934120", "Private RFC1918 IP in URL parameter"},

	// --- Internal hostnames ---
	{"ssrf", "substring", "host.docker.internal", "", "934190", "Docker host-access hostname"},
	{"ssrf", "substring", "kubernetes.default.svc", "", "934190", "Kubernetes API server internal hostname"},

	// --- Dangerous protocol schemes ---
	{"ssrf", "regex", `(?i)gopher://`, "", "934110", "Gopher protocol (SSRF relay)"},
	{"ssrf", "regex", `(?i)dict://`, "", "custom", "Dict protocol (SSRF relay)"},

	// --- DNS rebinding ---
	{"ssrf", "substring", "169.254.169.254.nip.io", "", "934110", "DNS rebinding via nip.io for metadata"},
	{"ssrf", "substring", "127.0.0.1.nip.io", "", "934120", "DNS rebinding via nip.io for loopback"},
}
