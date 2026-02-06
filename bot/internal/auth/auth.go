package auth

import (
	"fmt"
	"net/http"

	"parkjunwoo.com/claribot/internal/db"
)

const (
	issuerName  = "Claribot"
	accountName = "admin"
)

// Auth provides authentication operations backed by the global DB auth table.
type Auth struct {
	db *db.DB
}

// New creates a new Auth instance.
func New(d *db.DB) *Auth {
	return &Auth{db: d}
}

// SetupResult contains the result of the initial setup.
type SetupResult struct {
	TOTPURI string // Provisioning URI for QR code (only on first call before TOTP verification)
	Token   string // JWT token (set after TOTP verification succeeds)
}

// Setup performs initial setup: sets password and configures TOTP.
// First call (no TOTP code): stores password hash, generates TOTP secret, returns URI for QR.
// Second call (with TOTP code): verifies TOTP, marks setup complete, returns JWT.
func (a *Auth) Setup(password, totpCode string) (*SetupResult, error) {
	if a.isSetupCompleted() {
		return nil, fmt.Errorf("setup already completed")
	}

	// If TOTP code is provided, this is the verification step
	if totpCode != "" {
		return a.completeSetup(password, totpCode)
	}

	// First step: store password and generate TOTP secret
	return a.initSetup(password)
}

// initSetup stores the password hash and generates a TOTP secret.
func (a *Auth) initSetup(password string) (*SetupResult, error) {
	hash, err := HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	secret, uri, err := GenerateSecret(issuerName, accountName)
	if err != nil {
		return nil, fmt.Errorf("generate totp: %w", err)
	}

	// Store password hash and TOTP secret (upsert)
	for _, kv := range []struct{ key, value string }{
		{"password_hash", hash},
		{"totp_secret", secret},
	} {
		_, err := a.db.Exec(
			"INSERT INTO auth (key, value) VALUES (?, ?) ON CONFLICT(key) DO UPDATE SET value = ?",
			kv.key, kv.value, kv.value,
		)
		if err != nil {
			return nil, fmt.Errorf("store %s: %w", kv.key, err)
		}
	}

	return &SetupResult{TOTPURI: uri}, nil
}

// completeSetup verifies password + TOTP, marks setup complete, and returns a JWT.
func (a *Auth) completeSetup(password, totpCode string) (*SetupResult, error) {
	if err := a.verifyPassword(password); err != nil {
		return nil, err
	}

	var secret string
	err := a.db.QueryRow("SELECT value FROM auth WHERE key = 'totp_secret'").Scan(&secret)
	if err != nil {
		return nil, fmt.Errorf("totp secret not found: run setup without totp_code first")
	}

	if !ValidateTOTP(totpCode, secret) {
		return nil, fmt.Errorf("invalid TOTP code")
	}

	// Mark setup as completed
	_, err = a.db.Exec(
		"INSERT INTO auth (key, value) VALUES ('setup_completed', '1') ON CONFLICT(key) DO UPDATE SET value = '1'",
	)
	if err != nil {
		return nil, fmt.Errorf("mark setup completed: %w", err)
	}

	token, err := a.GenerateToken()
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}

	return &SetupResult{Token: token}, nil
}

// Login verifies password and TOTP code, then returns a JWT token.
func (a *Auth) Login(password, totpCode string) (string, error) {
	if !a.isSetupCompleted() {
		return "", fmt.Errorf("setup not completed")
	}

	if err := a.verifyPassword(password); err != nil {
		return "", err
	}

	var secret string
	err := a.db.QueryRow("SELECT value FROM auth WHERE key = 'totp_secret'").Scan(&secret)
	if err != nil {
		return "", fmt.Errorf("totp secret not found")
	}

	if !ValidateTOTP(totpCode, secret) {
		return "", fmt.Errorf("invalid TOTP code")
	}

	return a.GenerateToken()
}

// StatusResult contains the current authentication status.
type StatusResult struct {
	SetupCompleted  bool `json:"setup_completed"`
	IsAuthenticated bool `json:"is_authenticated"`
}

// Status returns whether setup is completed and if the request is authenticated.
func (a *Auth) Status(r *http.Request) StatusResult {
	result := StatusResult{
		SetupCompleted: a.isSetupCompleted(),
	}

	token := GetTokenFromRequest(r)
	if token != "" {
		if _, err := a.ValidateToken(token); err == nil {
			result.IsAuthenticated = true
		}
	}

	return result
}

// IsAuthenticated checks if the given request has a valid JWT token.
func (a *Auth) IsAuthenticated(r *http.Request) bool {
	token := GetTokenFromRequest(r)
	if token == "" {
		return false
	}
	_, err := a.ValidateToken(token)
	return err == nil
}

// IsSetupCompleted returns whether initial setup has been completed.
func (a *Auth) IsSetupCompleted() bool {
	return a.isSetupCompleted()
}

func (a *Auth) isSetupCompleted() bool {
	var value string
	err := a.db.QueryRow("SELECT value FROM auth WHERE key = 'setup_completed'").Scan(&value)
	return err == nil && value == "1"
}

// GetTOTPSetupURI returns the TOTP provisioning URI for the pending setup.
// Returns error if setup is already completed or no secret is stored.
func (a *Auth) GetTOTPSetupURI() (string, error) {
	if a.isSetupCompleted() {
		return "", fmt.Errorf("setup already completed")
	}

	var secret string
	err := a.db.QueryRow("SELECT value FROM auth WHERE key = 'totp_secret'").Scan(&secret)
	if err != nil {
		return "", fmt.Errorf("no totp secret found: run setup first")
	}

	// Reconstruct the provisioning URI
	uri := fmt.Sprintf("otpauth://totp/%s:%s?secret=%s&issuer=%s&algorithm=SHA1&digits=6&period=30",
		issuerName, accountName, secret, issuerName)
	return uri, nil
}

func (a *Auth) verifyPassword(password string) error {
	var hash string
	err := a.db.QueryRow("SELECT value FROM auth WHERE key = 'password_hash'").Scan(&hash)
	if err != nil {
		return fmt.Errorf("password not set")
	}
	if err := CheckPassword(hash, password); err != nil {
		return fmt.Errorf("invalid password")
	}
	return nil
}
