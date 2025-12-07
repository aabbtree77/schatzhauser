package main

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/aabbtree77/schatzhauser/config"
	"github.com/aabbtree77/schatzhauser/handlers"
	"github.com/aabbtree77/schatzhauser/protect"
)

func registerRoutes(mux *http.ServeMux, db *sql.DB,
	cfg *config.Config) {

	mux.Handle("/register", &handlers.RegisterHandler{
		DB:                  db,
		IPRLimiter:          protect.NewIPRateLimiter(tomlSect2IPRConfig(cfg.IPRateLimiter.Register)),
		AccountPerIPLimiter: cfg.AccountPerIPLimiter,
	})

	mux.Handle("/login", &handlers.LoginHandler{
		DB:         db,
		IPRLimiter: protect.NewIPRateLimiter(tomlSect2IPRConfig(cfg.IPRateLimiter.Login)),
	})

	mux.Handle("/logout", &handlers.LogoutHandler{
		DB:         db,
		IPRLimiter: protect.NewIPRateLimiter(tomlSect2IPRConfig(cfg.IPRateLimiter.Logout)),
	})

	mux.Handle("/profile", &handlers.ProfileHandler{
		DB:         db,
		IPRLimiter: protect.NewIPRateLimiter(tomlSect2IPRConfig(cfg.IPRateLimiter.Profile)),
	})
}

// Don't miss to pass sec.Enable here as the default will always be false
// and the IP rate limiter won't run inside the handler.
// There is no way to specify default Enable inside protect/ip_rate_limiter.go.

func tomlSect2IPRConfig(sec config.IPRateLimiterSection) protect.IPRateLimiterConfig {
	IPRConfig := protect.IPRateLimiterConfig{}
	IPRConfig.Enable = sec.Enable
	IPRConfig.MaxRequests = sec.MaxRequests
	IPRConfig.Window = time.Duration(sec.WindowMS) * time.Millisecond
	return IPRConfig
}
