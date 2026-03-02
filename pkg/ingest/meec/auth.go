package meec

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"crypto/tls"
	"encoding/base32"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"math"
	"net/http"
	"sync"
	"time"

	"github.com/dannyota/hotpot/pkg/base/httperr"
)

// TokenSource authenticates with the MEEC API and caches the token.
// Safe for concurrent use. Created once at the provider level and shared
// across all services.
type TokenSource struct {
	httpClient *http.Client
	baseURL    string
	apiVersion string
	username   string
	password   string
	authType   string
	totpSecret string

	mu    sync.Mutex
	token string
}

// NewTokenSource creates a new TokenSource.
func NewTokenSource(baseURL, apiVersion, username, password, authType, totpSecret string, verifySSL bool) *TokenSource {
	baseTransport := http.DefaultTransport.(*http.Transport).Clone()
	if !verifySSL {
		baseTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	return &TokenSource{
		httpClient: &http.Client{Transport: baseTransport},
		baseURL:    baseURL,
		apiVersion: apiVersion,
		username:   username,
		password:   password,
		authType:   authType,
		totpSecret: totpSecret,
	}
}

// Token returns a cached auth token, authenticating if needed.
func (ts *TokenSource) Token() (string, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if ts.token != "" {
		return ts.token, nil
	}

	token, err := ts.authenticate()
	if err != nil {
		return "", err
	}

	ts.token = token
	return token, nil
}

// Invalidate clears the cached token, forcing re-authentication on next call.
func (ts *TokenSource) Invalidate() {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	ts.token = ""
}

// authResponse is the MEEC authentication response.
type authResponse struct {
	Status           string `json:"status"`
	ErrorCode        string `json:"error_code"`
	ErrorDescription string `json:"error_description"`
	MessageResponse  struct {
		Authentication struct {
			AuthData struct {
				AuthToken string `json:"auth_token"`
			} `json:"auth_data"`
			TwoFactorData struct {
				UniqueUserID          string `json:"unique_userID"`
				IsTwoFactorEnabled    bool   `json:"is_TwoFactor_Enabled"`
				OTPValidationRequired bool   `json:"OTP_Validation_Required"`
			} `json:"two_factor_data"`
		} `json:"authentication"`
	} `json:"message_response"`
}

func (ts *TokenSource) authenticate() (string, error) {
	// Step 1: Login with credentials
	resp, err := ts.login()
	if err != nil {
		return "", err
	}

	// If token is returned directly (no 2FA), we're done
	if token := resp.MessageResponse.Authentication.AuthData.AuthToken; token != "" {
		slog.Info("meec auth succeeded")
		return token, nil
	}

	// Step 2: Handle 2FA if required
	tfa := resp.MessageResponse.Authentication.TwoFactorData
	if !tfa.OTPValidationRequired {
		return "", fmt.Errorf("MEEC auth: no token and no 2FA required (unexpected response)")
	}

	if ts.totpSecret == "" {
		return "", fmt.Errorf("MEEC auth: 2FA required but no totp_secret configured")
	}

	otp, err := generateTOTP(ts.totpSecret)
	if err != nil {
		return "", fmt.Errorf("generate TOTP: %w", err)
	}

	token, err := ts.validateOTP(tfa.UniqueUserID, otp)
	if err != nil {
		return "", err
	}

	slog.Info("meec auth succeeded (2FA)")
	return token, nil
}

func (ts *TokenSource) login() (*authResponse, error) {
	endpoint := fmt.Sprintf("/api/%s/desktop/authentication", ts.apiVersion)
	requestURL := ts.baseURL + endpoint

	payload := map[string]string{
		"username":  ts.username,
		"password":  base64.StdEncoding.EncodeToString([]byte(ts.password)),
		"auth_type": ts.authType,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal auth payload: %w", err)
	}

	start := time.Now()
	slog.Debug("meec auth request", "endpoint", endpoint)

	req, err := http.NewRequest("POST", requestURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create auth request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := ts.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute auth request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read auth response: %w", err)
	}

	slog.Info("meec auth response", "status", resp.StatusCode, "durationMs", time.Since(start).Milliseconds())

	var authResp authResponse
	if err := json.Unmarshal(respBody, &authResp); err != nil {
		return nil, fmt.Errorf("parse auth response: %w", err)
	}

	if authResp.Status == "error" {
		slog.Error("meec auth failed", "errorCode", authResp.ErrorCode, "errorDescription", authResp.ErrorDescription)
		if authResp.ErrorCode == "10001" || authResp.ErrorCode == "10002" {
			return nil, &httperr.APIError{Code: http.StatusUnauthorized}
		}
		return nil, fmt.Errorf("MEEC auth error %s: %s", authResp.ErrorCode, authResp.ErrorDescription)
	}

	return &authResp, nil
}

func (ts *TokenSource) validateOTP(uid, otp string) (string, error) {
	endpoint := fmt.Sprintf("/api/%s/desktop/authentication/otpValidate", ts.apiVersion)
	requestURL := ts.baseURL + endpoint

	payload := map[string]string{
		"uid":                uid,
		"otp":                otp,
		"rememberme_enabled": "true",
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal otp payload: %w", err)
	}

	start := time.Now()
	slog.Debug("meec otp validate request", "endpoint", endpoint)

	req, err := http.NewRequest("POST", requestURL, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("create otp request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := ts.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("execute otp request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read otp response: %w", err)
	}

	slog.Info("meec otp response", "status", resp.StatusCode, "durationMs", time.Since(start).Milliseconds())

	var otpResp authResponse
	if err := json.Unmarshal(respBody, &otpResp); err != nil {
		return "", fmt.Errorf("parse otp response: %w", err)
	}

	if otpResp.Status == "error" {
		slog.Error("meec otp validation failed", "errorCode", otpResp.ErrorCode, "errorDescription", otpResp.ErrorDescription)
		return "", fmt.Errorf("MEEC OTP error %s: %s", otpResp.ErrorCode, otpResp.ErrorDescription)
	}

	token := otpResp.MessageResponse.Authentication.AuthData.AuthToken
	if token == "" {
		return "", fmt.Errorf("MEEC OTP succeeded but no token in response")
	}

	return token, nil
}

// generateTOTP generates a 6-digit TOTP code from a base32-encoded secret.
func generateTOTP(secret string) (string, error) {
	key, err := base32.StdEncoding.DecodeString(secret)
	if err != nil {
		return "", fmt.Errorf("decode TOTP secret: %w", err)
	}

	counter := uint64(time.Now().Unix()) / 30
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, counter)

	mac := hmac.New(sha1.New, key)
	mac.Write(buf)
	h := mac.Sum(nil)

	offset := h[len(h)-1] & 0x0F
	code := (binary.BigEndian.Uint32(h[offset:offset+4]) & 0x7FFFFFFF) % uint32(math.Pow10(6))

	return fmt.Sprintf("%06d", code), nil
}
