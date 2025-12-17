package server

import (
	"database/sql"
	"net/http"

	"github.com/aabbtree77/schatzhauser/internal/config"
	"github.com/aabbtree77/schatzhauser/internal/guards"
	"github.com/aabbtree77/schatzhauser/internal/handlers"
)

// RegisterRoutes binds all HTTP routes to the stdlib mux.
func RegisterRoutes(mux *http.ServeMux, db *sql.DB, cfg *config.Config) {

	// ────────────────────────────────────────
	// Proof of Work (shared)
	// ────────────────────────────────────────

	powCfg := guards.PowConfig{
		Enable:     cfg.ProofOfWork.Enable,
		Difficulty: cfg.ProofOfWork.Difficulty,
		TTL:        cfg.ProofOfWork.TTL(),
		SecretKey:  cfg.ProofOfWork.DecodedSecretKey,
	}

	powHandler := guards.NewPoWHandler(powCfg)
	mux.Handle("/api/pow/challenge", powHandler)

	// ────────────────────────────────────────
	// Register
	// ────────────────────────────────────────

	registerIPR := guards.NewIPRateGuard(guards.IPRateLimiterConfig{
		Enable:      cfg.IPRateLimiter.Register.Enable,
		MaxRequests: cfg.IPRateLimiter.Register.MaxRequests,
		Window:      cfg.IPRateLimiter.Register.Window(),
	})

	registerBody := guards.NewBodySizeGuard(
		cfg.RBodySizeLimiter.Register.Enable,
		cfg.RBodySizeLimiter.Register.MaxRBodyBytes,
	)

	registerGuards := []guards.Guard{
		registerIPR,
		registerBody,
		guards.NewPoWGuard(powCfg),
	}

	mux.Handle("/api/register", &handlers.RegisterHandler{
		DB:     db,
		Guards: registerGuards,
		// AccountPerIPLimiter can remain here if needed later
		AccountPerIPLimiter: cfg.AccountPerIPLimiter,
	})

	// ────────────────────────────────────────
	// Login
	// ────────────────────────────────────────

	loginIPR := guards.NewIPRateGuard(guards.IPRateLimiterConfig{
		Enable:      cfg.IPRateLimiter.Login.Enable,
		MaxRequests: cfg.IPRateLimiter.Login.MaxRequests,
		Window:      cfg.IPRateLimiter.Login.Window(),
	})

	loginBody := guards.NewBodySizeGuard(
		cfg.RBodySizeLimiter.Login.Enable,
		cfg.RBodySizeLimiter.Login.MaxRBodyBytes,
	)

	loginGuards := []guards.Guard{
		loginIPR,
		loginBody,
	}

	mux.Handle("/api/login", &handlers.LoginHandler{
		DB:     db,
		Guards: loginGuards,
	})

	// ────────────────────────────────────────
	// Logout
	// ────────────────────────────────────────

	logoutIPR := guards.NewIPRateGuard(guards.IPRateLimiterConfig{
		Enable:      cfg.IPRateLimiter.Logout.Enable,
		MaxRequests: cfg.IPRateLimiter.Logout.MaxRequests,
		Window:      cfg.IPRateLimiter.Logout.Window(),
	})

	logoutGuards := []guards.Guard{
		logoutIPR,
	}

	mux.Handle("/api/logout", &handlers.LogoutHandler{
		DB:     db,
		Guards: logoutGuards,
	})

	// ────────────────────────────────────────
	// Profile
	// ────────────────────────────────────────

	profileIPR := guards.NewIPRateGuard(guards.IPRateLimiterConfig{
		Enable:      cfg.IPRateLimiter.Profile.Enable,
		MaxRequests: cfg.IPRateLimiter.Profile.MaxRequests,
		Window:      cfg.IPRateLimiter.Profile.Window(),
	})

	profileGuards := []guards.Guard{
		profileIPR,
	}

	mux.Handle("/api/profile", &handlers.ProfileHandler{
		DB:     db,
		Guards: profileGuards,
	})
}
