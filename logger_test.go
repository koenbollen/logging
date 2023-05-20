package logging_test

import (
	"context"
	"io"
	"testing"

	"github.com/koenbollen/logging"
	"go.uber.org/zap"
)

func TestLogger_FatalPanic(t *testing.T) {
	logger := logging.New(context.Background(), "test", "fatal")
	defer logger.Sync()

	defer func() {
		recovered := recover()
		if recovered == nil {
			t.Fatal("expected panic, got nil instead")
		}
		if recovered != "log message" {
			t.Fatalf("expected EOF, got %v instead", recovered)
		}
	}()

	logger.Fatal("log message", zap.Error(io.EOF))
}
