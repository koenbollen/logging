package logging

import (
	"context"
	"log/slog"
	"os"
	"runtime/debug"

	"github.com/jussi-kalliokoski/slogdriver"
	"github.com/lmittmann/tint"
)

// New creates a new *slog.Logger configured based on the environment variables
// ENV, DEBUG and VERSION.
// When env is "local" the resulting logger will output in a human readable
// format (colored when output is a tty). When ENV is not "local" the logger
// will output JSON log-entries compatible with Stackdriver.
// The env "test" will output to os.Stdout.
//
// The given service, component and VERSION are attached to the logger as
// default fields.
//
// The severity level filter is _info_, otherwise _debug_ when
// ENV is "local" or the env DEBUG is not empty.
//
// see: https://cloud.google.com/logging/docs/reference/v2/rest/v2/LogEntry
func New(ctx context.Context, service, component string) *slog.Logger {
	env := os.Getenv("ENV")
	if env == "" {
		env = "local"
	}

	var handler slog.Handler

	switch env {
	case "local":
		useColor := false
		if fileInfo, _ := os.Stdout.Stat(); fileInfo != nil && (fileInfo.Mode()&os.ModeCharDevice) != 0 {
			useColor = true
		}
		handler = tint.NewHandler(os.Stderr, &tint.Options{
			NoColor:    !useColor,
			Level:      slog.LevelDebug,
			TimeFormat: "15:04:05.000",
		})
	case "test":
		handler = slogdriver.NewHandler(os.Stdout, slogdriver.Config{
			Level: slog.LevelDebug,
		})
	default:
		level := slog.LevelInfo
		if os.Getenv("DEBUG") != "" {
			level = slog.LevelDebug
		}
		handler = slogdriver.NewHandler(os.Stderr, slogdriver.Config{
			Level: level,
		})
	}

	logger := slog.New(handler)
	logger = logger.With("env", env)
	logger = logger.With("service", service)
	if component != "" {
		logger = logger.With("component", component)
	}
	if version, found := getVersionFromEnvOrRuntime(); found {
		logger = logger.With("version", version)
	}

	slog.SetDefault(logger)
	return logger
}

func getVersionFromEnvOrRuntime() (string, bool) {
	if version, found := os.LookupEnv("VERSION"); found && version != "" {
		return version, true
	}
	info, _ := debug.ReadBuildInfo()
	var revision string
	var modified bool
	for _, v := range info.Settings {
		switch v.Key {
		case "vcs.revision":
			revision = v.Value
		case "vcs.modified":
			modified = v.Value == "true"
		}
	}
	if revision != "" {
		version := revision
		if modified {
			version += "-modified"
		}
		return version, true
	}
	return "", false
}
