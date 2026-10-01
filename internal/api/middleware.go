package api

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Auth handles OTP-based authentication.
type Auth struct {
	secret  string
	mu      sync.Mutex
	otps    map[string]time.Time
	tokens  map[string]bool
}

// NewAuth creates a new auth handler with the given secret.
func NewAuth(secret string) *Auth {
	return &Auth{
		secret: secret,
		otps:   make(map[string]time.Time),
		tokens: make(map[string]bool),
	}
}

// RequestOTP generates an OTP for the given email.
func (a *Auth) RequestOTP(email string) string {
	code := generateOTP()
	a.mu.Lock()
	a.otps[email+":"+code] = time.Now().Add(5 * time.Minute)
	a.mu.Unlock()
	return code
}

// VerifyOTP verifies an OTP and returns a bearer token.
func (a *Auth) VerifyOTP(email, code string) (string, error) {
	key := email + ":" + code
	a.mu.Lock()
	defer a.mu.Unlock()

	expiry, ok := a.otps[key]
	if !ok || time.Now().After(expiry) {
		delete(a.otps, key)
		return "", &authError{"invalid or expired OTP", "request a new OTP via POST /auth/request with email"}
	}
	delete(a.otps, key)

	token := generateToken()
	a.tokens[token] = true
	return token, nil
}

// ValidateToken checks if a bearer token is valid.
func (a *Auth) ValidateToken(token string) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.tokens[token]
}

// Middleware returns an HTTP middleware that checks auth.
func (a *Auth) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip auth for public endpoints
		path := r.URL.Path
		if path == "/help" || path == "/.well-known/agent.md" ||
			path == "/auth/request" || path == "/auth/verify" ||
			path == "/health" || path == "/mcp" {
			next.ServeHTTP(w, r)
			return
		}

		// Check bearer token
		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			writeError(w, r, http.StatusUnauthorized,
				"missing auth token",
				"call POST /auth/request with email to get an OTP, then POST /auth/verify to get a bearer token")
			return
		}
		token := strings.TrimPrefix(auth, "Bearer ")
		if !a.ValidateToken(token) {
			writeError(w, r, http.StatusUnauthorized,
				"invalid or expired token",
				"request a new token via POST /auth/request with email, then POST /auth/verify")
			return
		}
		next.ServeHTTP(w, r)
	})
}

type authError struct {
	msg  string
	hint string
}

func (e *authError) Error() string { return e.msg }

func generateOTP() string {
	b := make([]byte, 3)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func generateToken() string {
	b := make([]byte, 24)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// Constant-time comparison to prevent timing attacks.
func secureCompare(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}
