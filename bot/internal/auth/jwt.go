package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	cookieName      = "claribot_token"
	tokenExpiration = 24 * time.Hour
)

// getOrCreateJWTSecret retrieves the JWT secret from DB, or generates and stores a new one.
func (a *Auth) getOrCreateJWTSecret() ([]byte, error) {
	var secret string
	err := a.db.QueryRow("SELECT value FROM auth WHERE key = 'jwt_secret'").Scan(&secret)
	if err == nil {
		return []byte(secret), nil
	}

	// Generate a new 32-byte random secret
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return nil, fmt.Errorf("generate jwt secret: %w", err)
	}
	secret = hex.EncodeToString(b)

	_, err = a.db.Exec("INSERT INTO auth (key, value) VALUES ('jwt_secret', ?)", secret)
	if err != nil {
		return nil, fmt.Errorf("store jwt secret: %w", err)
	}
	return []byte(secret), nil
}

// GenerateToken creates a signed JWT token with a 24-hour expiration.
func (a *Auth) GenerateToken() (string, error) {
	secret, err := a.getOrCreateJWTSecret()
	if err != nil {
		return "", err
	}

	claims := jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenExpiration)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		Issuer:    "claribot",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

// ValidateToken validates a JWT token string and returns the claims if valid.
func (a *Auth) ValidateToken(tokenStr string) (*jwt.RegisteredClaims, error) {
	secret, err := a.getOrCreateJWTSecret()
	if err != nil {
		return nil, err
	}

	token, err := jwt.ParseWithClaims(tokenStr, &jwt.RegisteredClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return secret, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return claims, nil
}

// SetTokenCookie sets the JWT token as an HTTP-only cookie.
func SetTokenCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(tokenExpiration.Seconds()),
	})
}

// ClearTokenCookie removes the JWT token cookie.
func ClearTokenCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}

// GetTokenFromRequest extracts the JWT token from the request cookie.
func GetTokenFromRequest(r *http.Request) string {
	cookie, err := r.Cookie(cookieName)
	if err != nil {
		return ""
	}
	return cookie.Value
}
