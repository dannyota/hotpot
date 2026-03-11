package config

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// SeedAuthEndpointPatterns inserts predefined auth endpoint keywords.
// These keyword patterns detect authentication-related URIs across common
// frameworks (WordPress, Spring, Django, Express, etc.) without configuration.
// Users can add custom patterns via admin UI for application-specific endpoints.
func SeedAuthEndpointPatterns(ctx context.Context, db *sql.DB) error {
	if len(authEndpointPatterns) == 0 {
		return nil
	}

	now := time.Now()
	var b strings.Builder
	b.WriteString(`INSERT INTO config.auth_endpoint_patterns
		(pattern_type, match_mode, pattern, source, description, is_active, created_at, updated_at)
		VALUES `)

	args := make([]any, 0, len(authEndpointPatterns)*8)
	for i, p := range authEndpointPatterns {
		if i > 0 {
			b.WriteString(", ")
		}
		base := i * 8
		fmt.Fprintf(&b, "($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)",
			base+1, base+2, base+3, base+4, base+5, base+6, base+7, base+8)
		args = append(args, p.patternType, "keyword", p.pattern,
			"seed", p.description, true, now, now)
	}

	b.WriteString(` ON CONFLICT (pattern_type, pattern) DO NOTHING`)

	_, err := db.ExecContext(ctx, b.String(), args...)
	if err != nil {
		return fmt.Errorf("upsert auth endpoint patterns (%d entries): %w", len(authEndpointPatterns), err)
	}
	return nil
}

type authEndpointEntry struct {
	patternType string
	pattern     string
	description string
}

var authEndpointPatterns = []authEndpointEntry{
	// Login / Authentication
	{"login", "login", "Login-related URI"},
	{"login", "signin", "Sign-in URI"},
	{"login", "sign-in", "Sign-in URI (hyphenated)"},
	{"login", "authenticate", "Authentication URI"},
	{"login", "oauth/token", "OAuth token endpoint"},
	{"login", "connect/token", "OIDC token endpoint"},

	// OTP / MFA
	{"otp", "otp", "OTP verification"},
	{"otp", "mfa", "MFA verification"},
	{"otp", "2fa", "2FA verification"},
	{"otp", "totp", "TOTP verification"},
	{"otp", "verify-code", "Code verification"},

	// Password Reset
	{"password_reset", "forgot-password", "Forgot password"},
	{"password_reset", "reset-password", "Reset password"},
	{"password_reset", "password/reset", "Password reset path"},
	{"password_reset", "password-recovery", "Password recovery"},

	// Registration
	{"register", "register", "Registration"},
	{"register", "signup", "Sign up"},
	{"register", "sign-up", "Sign up (hyphenated)"},
	{"register", "create-account", "Account creation"},

	// Admin / Management
	{"admin", "admin", "Admin panel"},
	{"admin", "wp-admin", "WordPress admin"},
	{"admin", "actuator", "Spring Boot actuator"},
	{"admin", "phpmyadmin", "phpMyAdmin"},
	{"admin", "grafana", "Grafana dashboard"},
	{"admin", "kibana", "Kibana dashboard"},
	{"admin", "swagger", "Swagger API docs"},
	{"admin", "graphiql", "GraphiQL IDE"},
	{"admin", "console", "Management console"},

	// Token / API Key
	{"token", "refresh-token", "Token refresh"},
	{"token", "apikey", "API key management"},
	{"token", "api-key", "API key (hyphenated)"},
}
