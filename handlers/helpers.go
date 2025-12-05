package handlers

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"time"

	"github.com/aabbtree77/schatzhauser/db"
	"golang.org/x/crypto/bcrypt"
)

const (
	SessionCookieName = "schatz_sess"
	SessionDuration   = 30 * 24 * time.Hour // 30 days
	TokenByteLen      = 32
)

// ---------- JSON helpers ----------

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func badRequest(w http.ResponseWriter, msg string) {
	writeJSON(w, http.StatusBadRequest, map[string]any{
		"status":  "error",
		"message": msg,
	})
}

func unauthorized(w http.ResponseWriter, msg string) {
	writeJSON(w, http.StatusUnauthorized, map[string]any{
		"status":  "error",
		"message": msg,
	})
}

func internalError(w http.ResponseWriter, msg string) {
	writeJSON(w, http.StatusInternalServerError, map[string]any{
		"status":  "error",
		"message": msg,
	})
}

func created(w http.ResponseWriter, v any) {
	writeJSON(w, http.StatusCreated, v)
}

// ---------- password helpers ----------

func hashPassword(pwd string) (string, error) {
	h, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
	return string(h), err
}

func comparePassword(hash, pwd string) bool {
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(pwd)); err != nil {
		return false
	}
	return true
}

// ---------- token helpers ----------

func generateSessionToken() (string, error) {
	b := make([]byte, TokenByteLen)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// ---------- cookie helpers ----------

// setSessionCookie writes a session cookie with token and expiry.
// Cookie Secure bit is set only if the request is TLS.
func setSessionCookie(w http.ResponseWriter, r *http.Request, token string, expires time.Time) {
	c := &http.Cookie{
		Name:     SessionCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Expires:  expires,
		MaxAge:   int(time.Until(expires).Seconds()),
	}
	// Only set Secure if the incoming request is TLS (so local dev works).
	if r.TLS != nil {
		c.Secure = true
	}
	http.SetCookie(w, c)
}

// clearSessionCookie removes cookie in client
func clearSessionCookie(w http.ResponseWriter, r *http.Request) {
	c := &http.Cookie{
		Name:     SessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
	}
	// same Secure logic
	if r.TLS != nil {
		c.Secure = true
	}
	http.SetCookie(w, c)
}

// ---------- session helpers ----------

// getSessionFromRequest reads session cookie, loads session row and validates expiry.
// Returns session (sqlc type) and nil error when valid; if not valid returns nil and an error.
func getSessionFromRequest(ctx context.Context, sqlDB *sql.DB, r *http.Request) (*db.Session, error) {
	c, err := r.Cookie(SessionCookieName)
	if err != nil {
		return nil, err
	}
	token := c.Value
	store := db.NewStore(sqlDB)

	sess, err := store.GetSessionByToken(ctx, token)
	if err != nil {
		return nil, err
	}

	// Check expiry: sqlc will map expires_at to time.Time
	if time.Now().After(sess.ExpiresAt) {
		// session expired: delete it and return error
		_ = store.DeleteSessionByToken(ctx, token)
		return nil, sql.ErrNoRows
	}

	return &sess, nil
}

func tooManyRequests(w http.ResponseWriter) {
	writeJSON(w, http.StatusTooManyRequests, map[string]any{
		"status":  "error",
		"message": "rate limit exceeded",
	})
}
