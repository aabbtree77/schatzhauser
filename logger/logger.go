package logger

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/aabbtree77/schatzhauser/config"
)

var (
	Log   *slog.Logger // default logger after Init
	debug bool
)

// Init configures the global logger.
// Dev → pretty text logs. Prod → JSON logs.
func Init(cfg config.Config) {
	debug = cfg.Debug

	var handler slog.Handler

	if debug {
		// Pretty human-readable logs for development.
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})
	} else {
		// Structured JSON for production.
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
	}

	Log = slog.New(handler)
}

// ------- Helper functions (simple wrappers) -------

func Debug(msg string, args ...any) {
	if debug && Log != nil {
		Log.Debug(msg, args...)
	}
}

func Info(msg string, args ...any) {
	if Log != nil {
		Log.Info(msg, args...)
	}
}

func Warn(msg string, args ...any) {
	if Log != nil {
		Log.Warn(msg, args...)
	}
}

func Error(msg string, args ...any) {
	if Log != nil {
		Log.Error(msg, args...)
	}
}

// Panic-safe — if Log is nil (Init never called), set one quickly.
func ensure() *slog.Logger {
	if Log == nil {
		Log = slog.New(slog.NewJSONHandler(os.Stdout, nil))
	}
	return Log
}

// ------- Extra helper: request logging -------

// Request logs method, path, IP, user-agent
func Request(r *http.Request) {
	ensure().Info("request",
		"method", r.Method,
		"path", r.URL.Path,
		"ip", clientIP(r),
		"ua", r.UserAgent(),
	)
}

func clientIP(r *http.Request) string {
	// Simple and safe; no trust in X-Forwarded-For for MVP.
	ip := r.RemoteAddr
	return ip
}

// For contexts that require a logger instance
func FromContext(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value("logger").(*slog.Logger); ok {
		return l
	}
	return ensure()
}
