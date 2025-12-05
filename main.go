package main

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/aabbtree77/schatzhauser/config"
	dbpkg "github.com/aabbtree77/schatzhauser/db"
	"github.com/aabbtree77/schatzhauser/logger"
)

func run(ctx context.Context, w io.Writer, args []string) error {

	// Load configuration
	cfg, err := config.LoadConfig("config.toml")
	if err != nil {
		logger.Error("failed to load config", "err", err)
	}
	logger.Init(*cfg)
	logger.Info("starting schatzhauser", "debug", cfg.Debug)

	// Build rate limiters
	registerIPRLimiter := config.BuildIPRLimiter(cfg.IPRateLimiter.Register)
	loginIPRLimiter := config.BuildIPRLimiter(cfg.IPRateLimiter.Login)
	logoutIPRLimiter := config.BuildIPRLimiter(cfg.IPRateLimiter.Logout)
	profileIPRLimiter := config.BuildIPRLimiter(cfg.IPRateLimiter.Profile)

	// -----------------------------------------------------
	// Database (SQLite)
	// -----------------------------------------------------
	db, err := sql.Open("sqlite3", cfg.DBPath+"?_foreign_keys=on")
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}

	// DB sanity check
	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("db ping: %w", err)
	}

	// -----------------------------------------------------
	// Run migrations (manual, SQL files)
	// -----------------------------------------------------
	if err := dbpkg.RunMigrations(db); err != nil {

		return fmt.Errorf("migrations: %w", err)
	}

	// -----------------------------------------------------
	// HTTP server setup
	// -----------------------------------------------------
	mux := http.NewServeMux()

	registerRoutes(mux, db, registerIPRLimiter, loginIPRLimiter, logoutIPRLimiter, profileIPRLimiter)

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	// -----------------------------------------------------
	// Server in goroutine + graceful shutdown
	// -----------------------------------------------------
	go func() {
		logger.Info("listening on :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", "err", err)
		}
	}()

	// Handle Ctrl+C
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt)
	defer stop()

	<-ctx.Done()
	logger.Info("shutting down")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("shutdown error", "err", err)
	}

	_ = db.Close()
	return nil
}

func main() {
	ctx := context.Background()
	if err := run(ctx, os.Stdout, os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
