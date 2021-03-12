package logging_test

import (
	"context"

	"github.com/koenbollen/logging"
	"go.uber.org/zap"
)

func Example() {
	ctx := context.Background()

	logger := logging.New(ctx, "myservice", "example")
	logger.Info("hello, world!!")

	ctx = logging.WithLogger(ctx, logger)
	err := someOperation(ctx)
	logger.Error("failed", zap.Error(err))
}

func someOperation(ctx context.Context) error {
	return nil
}
