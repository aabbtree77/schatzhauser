package main

import (
	"database/sql"
	"net/http"

	"github.com/aabbtree77/schatzhauser/handlers"
	"github.com/aabbtree77/schatzhauser/protect"
)

func registerRoutes(mux *http.ServeMux, db *sql.DB,
	registerIPRLimiter *protect.IPRateLimiter,
	loginIPRLimiter *protect.IPRateLimiter,
	logoutIPRLimiter *protect.IPRateLimiter,
	profileIPRLimiter *protect.IPRateLimiter) {

	mux.Handle("/register", &handlers.RegisterHandler{
		DB:         db,
		IPRLimiter: registerIPRLimiter,
	})

	mux.Handle("/login", &handlers.LoginHandler{
		DB:         db,
		IPRLimiter: loginIPRLimiter,
	})

	mux.Handle("/logout", &handlers.LogoutHandler{
		DB:         db,
		IPRLimiter: logoutIPRLimiter,
	})

	mux.Handle("/profile", &handlers.ProfileHandler{
		DB:         db,
		IPRLimiter: profileIPRLimiter,
	})
}
