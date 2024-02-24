package logging

import (
	"context"
	"os"
	"runtime/debug"

	"github.com/koenbollen/logging/internal"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// New creates a new *zap.Logger configured based on the environment variables
// ENV, DEBUG and VERSION.
// When env is "local" the resulting logger will output in a human readable
// format (colored when output is a tty). When ENV is not "local" the logger
// will output JSON log-entries compatible with Stackdriver.
//
// The given service, component and VERSION are attached to the logger as
// default fields.
//
// The severity level filter is _info_, otherwise _debug_ when
// ENV is "local" or the env DEBUG is not empty.
//
// see: https://cloud.google.com/logging/docs/reference/v2/rest/v2/LogEntry
func New(ctx context.Context, service, component string) *zap.Logger {
	env := os.Getenv("ENV")
	if env == "" {
		env = "local"
	}

	var config zap.Config
	if env == "local" {
		config = zap.NewDevelopmentConfig()
		config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
		if fileInfo, _ := os.Stdout.Stat(); fileInfo != nil && (fileInfo.Mode()&os.ModeCharDevice) != 0 {
			config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		}
	} else {
		config = zap.NewProductionConfig()
		config.Sampling = nil
		config.Development = env != "production"
		config.EncoderConfig = internal.EncoderConfig
	}
	if env == "test" {
		config.OutputPaths = []string{"stdout"}
		config.ErrorOutputPaths = []string{"stdout"}
		config.EncoderConfig.TimeKey = ""
	}

	if os.Getenv("DEBUG") != "" {
		config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	}

	logger, err := config.Build(zap.WithFatalHook(zapcore.WriteThenPanic))
	if err != nil {
		panic(err)
	}
	logger = logger.With(zap.String("env", env))
	logger = logger.With(zap.String("service", service))
	if component != "" {
		logger = logger.With(zap.String("component", component))
	}
	if version, found := getVersionFromEnvOrRuntime(); found {
		logger = logger.With(zap.String("version", version))
	}

	go func() {
		<-ctx.Done()
		_ = logger.Sync()
	}()
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
