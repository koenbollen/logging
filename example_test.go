package logging_test

import (
	"context"

	"github.com/koenbollen/logging"
)

func Example() {
	ctx := context.Background()

	logger := logging.New(ctx, "myservice", "example")
	logger.Info("hello, world!!")

	ctx = logging.WithLogger(ctx, logger)
	err := someOperation(ctx)
	logger.Error("failed", "err", err)
}

func someOperation(ctx context.Context) error {
	return nil
}
