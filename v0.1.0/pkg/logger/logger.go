package logger

import (
	"log/slog"
	"os"
)

var Log *slog.Logger

// Setup initializes the global logger based on the environment.
// If env is "production", it uses JSON handler.
// Otherwise, it uses Text handler (more human-readable).
func Setup(env string) {
	var handler slog.Handler

	opts := &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}

	if env == "production" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	Log = slog.New(handler)
	slog.SetDefault(Log)
}
