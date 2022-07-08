package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/koenbollen/logging"
	"go.uber.org/zap"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	logger := logging.New(ctx, "example", "api")
	ctx = logging.WithLogger(ctx, logger)
	logger.Info("init")
	defer logger.Info("fin")

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		logging.IgnoreRequest(r)
		w.WriteHeader(http.StatusOK)
	})

	http.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		logger := logging.GetLogger(r.Context())

		if r.URL.RawQuery == "error" {
			logger.Error("error", zap.Error(io.ErrUnexpectedEOF))
			http.Error(w, "error", http.StatusInternalServerError)
			panic(http.ErrAbortHandler)
		}

		logger.Debug("test endpoint hit", zap.String("qs", r.URL.RawQuery))
		w.WriteHeader(http.StatusCreated)
		fmt.Fprintln(w, "Hello!")
	})

	server := &http.Server{
		Addr:    ":8080",
		Handler: logging.Middleware(http.DefaultServeMux, logger),
	}
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("failed to listen or serve", zap.Error(err))
		}
	}()
	logger.Info("listening", zap.String("addr", server.Addr))

	<-ctx.Done()

	ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		logger.Fatal("failed to shutdown server", zap.Error(err))
	}
}
