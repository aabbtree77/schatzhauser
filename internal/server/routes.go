package server

import (
	"database/sql"
	"encoding/base64"
	"net/http"
	"os"
	"time"

	"github.com/aabbtree77/schatzhauser/internal/config"
	"github.com/aabbtree77/schatzhauser/internal/handlers"
	"github.com/aabbtree77/schatzhauser/internal/logger"
	"github.com/aabbtree77/schatzhauser/internal/protect"
)

// RegisterRoutes binds all HTTP routes to the stdlib mux.
func RegisterRoutes(mux *http.ServeMux, db *sql.DB, cfg *config.Config) {

	// Proof of Work Endpoint

	powCfg := protect.PowConfig{
		Enable:     cfg.ProofOfWork.Enable,
		Difficulty: cfg.ProofOfWork.Difficulty,
		TTL:        time.Duration(cfg.ProofOfWork.TTLSeconds) * time.Second,
	}

	if cfg.ProofOfWork.Enable {
		rawKey := cfg.ProofOfWork.SecretKey
		if rawKey == "" {
			logger.Error("proof_of_work.secret_key is missing or empty")
			os.Exit(1)
		}

		decodedKey, err := base64.StdEncoding.DecodeString(rawKey)
		if err != nil {
			logger.Error("proof_of_work.secret_key must be base64: %v", err)
			os.Exit(1)
		}

		if len(decodedKey) < 16 {
			logger.Error("proof_of_work.secret_key too short, need >= 16 bytes after decoding")
			os.Exit(1)
		}

		powCfg.SecretKey = decodedKey
	}

	powHandler := protect.NewPoWHandler(powCfg)
	powKey := powHandler.(*protect.PoWHandler).Key

	mux.Handle("/api/pow/challenge", powHandler)

	ipr := protect.NewIPRateLimiter(tomlSect2IPRConfig(cfg.IPRateLimiter.Register))
	rbl := validateParamsRBodySizeLimiter(cfg.RBodySizeLimiter.Register)

	//The rest of the routes (handlers)

	mux.Handle("/api/register", &handlers.RegisterHandler{
		DB:                  db,
		IPRLimiter:          ipr,
		AccountPerIPLimiter: cfg.AccountPerIPLimiter,
		RBodySizeLimiter:    rbl,
		PoWCfg:              powCfg,
		PoWKey:              powKey,
	})

	ipr = protect.NewIPRateLimiter(tomlSect2IPRConfig(cfg.IPRateLimiter.Login))
	rbl = validateParamsRBodySizeLimiter(cfg.RBodySizeLimiter.Login)

	mux.Handle("/api/login", &handlers.LoginHandler{
		DB:               db,
		IPRLimiter:       ipr,
		RBodySizeLimiter: rbl,
	})

	ipr = protect.NewIPRateLimiter(tomlSect2IPRConfig(cfg.IPRateLimiter.Logout))

	mux.Handle("/api/logout", &handlers.LogoutHandler{
		DB:         db,
		IPRLimiter: ipr,
	})

	ipr = protect.NewIPRateLimiter(tomlSect2IPRConfig(cfg.IPRateLimiter.Profile))

	mux.Handle("/api/profile", &handlers.ProfileHandler{
		DB:         db,
		IPRLimiter: ipr,
	})

}

/* Don't miss to pass sec.Enable here as the default will always be false
and the IP rate limiter won't run inside the handler.
There is no way to specify default Enable inside protect/ip_rate_limiter.go.
*/

func tomlSect2IPRConfig(sec config.IPRateLimiterSection) protect.IPRateLimiterConfig {
	c := protect.IPRateLimiterConfig{}
	c.Enable = sec.Enable
	c.MaxRequests = sec.MaxRequests
	c.Window = time.Duration(sec.WindowMS) * time.Millisecond
	return c
}

/*
This is nicer to do directly inside a handler,
but to save a bit of computation, prevalidate stuff here.
*/

func validateParamsRBodySizeLimiter(sec config.RBodySizeLimiterSection) config.RBodySizeLimiterSection {
	if !sec.Enable {
		sec.MaxRBodyBytes = 0
	}
	sec.MaxRBodyBytes = protect.NormalizePayloadLimit(sec.MaxRBodyBytes)
	return sec
}

/*
TDD: check and set every default value if a param in config.toml is missing.
*/
