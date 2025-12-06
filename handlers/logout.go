package handlers

import (
	"database/sql"
	"net/http"

	"github.com/aabbtree77/schatzhauser/db"
	"github.com/aabbtree77/schatzhauser/protect"
)

type LogoutHandler struct {
	DB         *sql.DB
	IPRLimiter *protect.IPRateLimiter
}

func (h *LogoutHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if h.IPRLimiter.Enable {
		ip := protect.GetIP(r)
		if ip != "" && !h.IPRLimiter.Allow(ip) {
			tooManyRequests(w)
			return
		}
	}

	// Attempt to get cookie
	c, err := r.Cookie(SessionCookieName)
	if err != nil {
		// No cookie — nothing to do. Return success (idempotent).
		writeJSON(w, http.StatusOK, map[string]any{
			"status":  "ok",
			"message": "no session",
		})
		return
	}

	token := c.Value
	store := db.NewStore(h.DB)

	// Delete session from DB (best-effort)
	_ = store.DeleteSessionByToken(r.Context(), token)

	// Clear cookie on client
	clearSessionCookie(w, r)

	writeJSON(w, http.StatusOK, map[string]any{
		"status":  "ok",
		"message": "logged out",
	})
}
