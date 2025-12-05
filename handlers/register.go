package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/aabbtree77/schatzhauser/db"
	"github.com/aabbtree77/schatzhauser/protect"
)

type RegisterHandler struct {
	DB         *sql.DB
	IPRLimiter *protect.IPRateLimiter
}

type RegisterInput struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (h *RegisterHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	/*
		ip := protect.GetIP(r)
		if ip != "" && !h.IPRLimiter.Allow(ip) {
			tooManyRequests(w)
			return
		}
	*/
	var in RegisterInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		badRequest(w, "invalid json")
		return
	}

	if in.Username == "" || in.Password == "" {
		badRequest(w, "username and password required")
		return
	}

	store := db.NewStore(h.DB)

	hashed, err := hashPassword(in.Password)
	if err != nil {
		internalError(w, "cannot hash password")
		return
	}

	user, err := store.CreateUser(r.Context(), db.CreateUserParams{
		Username:     in.Username,
		PasswordHash: hashed,
	})
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			writeJSON(w, http.StatusConflict, map[string]any{
				"status":  "error",
				"message": "username already taken",
			})
			return
		}
		internalError(w, "cannot create user")
		return
	}

	created(w, map[string]any{
		"id":       user.ID,
		"username": user.Username,
	})
}
