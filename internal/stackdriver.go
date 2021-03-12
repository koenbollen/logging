package internal

import (
	"time"

	"go.uber.org/zap/zapcore"
)

var zapLevelToStackdriver = map[zapcore.Level]string{
	zapcore.DebugLevel:  "DEBUG",
	zapcore.InfoLevel:   "INFO",
	zapcore.WarnLevel:   "WARNING",
	zapcore.ErrorLevel:  "ERROR",
	zapcore.DPanicLevel: "CRITICAL",
	zapcore.PanicLevel:  "ALERT",
	zapcore.FatalLevel:  "EMERGENCY",
}

// EncoderConfig is the Stackdriver compatible zapcore.EncoderConfig instance.
var EncoderConfig = zapcore.EncoderConfig{
	NameKey:        "logger",
	LevelKey:       "severity",
	TimeKey:        "timestamp",
	CallerKey:      "caller",
	MessageKey:     "message",
	StacktraceKey:  "stacktrace",
	EncodeTime:     UTCTimeEncoder,
	EncodeLevel:    EncodeLevel,
	LineEnding:     zapcore.DefaultLineEnding,
	EncodeCaller:   zapcore.ShortCallerEncoder,
	EncodeDuration: zapcore.SecondsDurationEncoder,
}

func UTCTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.UTC().Format(time.RFC3339Nano))
}

func EncodeLevel(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(zapLevelToStackdriver[l])
}
