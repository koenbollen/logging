package logging_test

import (
	"context"
	"io"
	"os"
	"testing"

	"github.com/koenbollen/logging"
	"go.uber.org/zap"
)

func ExampleLogger_Info() {
	os.Setenv("ENV", "test") // output to stdout and disable timestamps
	logger := logging.New(context.Background(), "test", "info")

	logger.Info("hello, world!!")

	// Output:
	// {"severity":"INFO","caller":"logging/logger_test.go:17","message":"hello, world!!","env":"test","service":"test","component":"info"}
}

func ExampleLogger_Error() {
	os.Setenv("ENV", "test") // output to stdout and disable timestamps
	logger := logging.New(context.Background(), "test", "error")

	logger.Error("failed reading file", zap.Error(io.EOF))

	// Output:
	// {"severity":"ERROR","caller":"logging/logger_test.go:27","message":"failed reading file","env":"test","service":"test","component":"error","error":"EOF","stacktrace":"github.com/koenbollen/logging_test.ExampleLogger_Error\n\t/home/koen/koenbollen/logging/logger_test.go:27\ntesting.runExample\n\t/usr/local/go/src/testing/run_example.go:63\ntesting.runExamples\n\t/usr/local/go/src/testing/example.go:40\ntesting.(*M).Run\n\t/usr/local/go/src/testing/testing.go:2029\nmain.main\n\t_testmain.go:53\nruntime.main\n\t/usr/local/go/src/runtime/proc.go:271"}
}

func TestLogger_FatalPanic(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	logger := logging.New(ctx, "test", "fatal")

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
