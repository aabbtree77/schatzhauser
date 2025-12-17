package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/aabbtree77/schatzhauser/db"
	"github.com/aabbtree77/schatzhauser/internal/config"
	"github.com/aabbtree77/schatzhauser/internal/guards"
	"github.com/aabbtree77/schatzhauser/internal/httpx"
)

type RegisterHandler struct {
	DB *sql.DB

	// Injected request-level guards (PoW, body size, etc.)
	Guards []guards.Guard

	// Domain config
	AccountPerIPLimiter config.AccountPerIPLimiterConfig
}

type RegisterInput struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

/*
────────────────────────────────────────────────────────────
HTTP entrypoint
────────────────────────────────────────────────────────────
*/

func (h *RegisterHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Guards (request-level, stateless)
	for _, g := range h.Guards {
		if !g.Check(w, r) {
			return
		}
	}

	// Decode input
	var in RegisterInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		httpx.BadRequest(w, "invalid json")
		return
	}

	in.Username = strings.TrimSpace(in.Username)
	if in.Username == "" || in.Password == "" {
		httpx.BadRequest(w, "username and password required")
		return
	}

	h.register(w, r, in)
}

/*
────────────────────────────────────────────────────────────
Domain logic
────────────────────────────────────────────────────────────
*/

func (h *RegisterHandler) register(
	w http.ResponseWriter,
	r *http.Request,
	in RegisterInput,
) {
	ip := strings.TrimSpace(guards.GetIP(r))

	store := db.NewStore(h.DB)

	tx, err := h.DB.BeginTx(r.Context(), &sql.TxOptions{})
	if err != nil {
		httpx.InternalError(w, "cannot begin transaction")
		return
	}
	defer tx.Rollback()

	txStore := store.WithTx(tx)

	// Force SQLite write lock early to prevent IP races
	if err := txStore.TouchUsersTable(r.Context(), ip); err != nil {
		httpx.InternalError(w, "cannot lock users table")
		return
	}

	limiter := accountPerIPLimiter{
		Enable:      h.AccountPerIPLimiter.Enable,
		MaxAccounts: h.AccountPerIPLimiter.MaxAccounts,
		CountFn:     txStore.CountUsersByIP,
	}

	ok, err := limiter.allow(r.Context(), ip)
	if err != nil {
		httpx.InternalError(w, "cannot check ip usage")
		return
	}
	if !ok {
		httpx.TooManyRequests(w)
		return
	}

	hash, err := HashPassword(in.Password)
	if err != nil {
		httpx.InternalError(w, "cannot hash password")
		return
	}

	user, err := txStore.CreateUserWithIP(
		r.Context(),
		db.CreateUserWithIPParams{
			Username:     in.Username,
			PasswordHash: hash,
			Ip:           ip,
		},
	)
	if err != nil {
		if IsUniqueConstraint(err) {
			httpx.WriteJSON(w, http.StatusConflict, map[string]any{
				"status":  "error",
				"message": "username already taken",
			})
			return
		}
		httpx.InternalError(w, "cannot create user")
		return
	}

	if err := tx.Commit(); err != nil {
		httpx.InternalError(w, "cannot commit transaction")
		return
	}

	httpx.Created(w, map[string]any{
		"id":       user.ID,
		"username": user.Username,
	})
}

/*
────────────────────────────────────────────────────────────
Local helper (domain-level, not a guard)
────────────────────────────────────────────────────────────
*/

type accountPerIPLimiter struct {
	Enable      bool
	MaxAccounts int
	CountFn     func(ctx context.Context, ip string) (int64, error)
}

func (l *accountPerIPLimiter) allow(ctx context.Context, ip string) (bool, error) {
	if !l.Enable {
		return true, nil
	}
	if l.MaxAccounts <= 0 || ip == "" {
		return true, nil
	}

	n, err := l.CountFn(ctx, ip)
	if err != nil {
		return false, err
	}
	return n < int64(l.MaxAccounts), nil
}
