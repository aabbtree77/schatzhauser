package handlers

import (
	"database/sql"
	"net/http"

	"github.com/aabbtree77/schatzhauser/db"
	"github.com/aabbtree77/schatzhauser/protect"
)

type ProfileHandler struct {
	DB         *sql.DB
	IPRLimiter *protect.IPRateLimiter
}

func (h *ProfileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ip := protect.GetIP(r)
	if ip != "" && !h.IPRLimiter.Allow(ip) {
		tooManyRequests(w)
		return
	}

	// Validate session cookie & fetch session row
	sess, err := getSessionFromRequest(r.Context(), h.DB, r)
	if err != nil {
		unauthorized(w, "unauthorized")
		return
	}

	// Load user
	store := db.NewStore(h.DB)
	user, err := store.GetUserByID(r.Context(), sess.UserID)
	if err != nil {
		unauthorized(w, "unauthorized")
		return
	}

	// Return safe user info
	writeJSON(w, http.StatusOK, map[string]any{
		"status": "ok",
		"user": map[string]any{
			"id":       user.ID,
			"username": user.Username,
			"created":  user.CreatedAt,
		},
	})
}
