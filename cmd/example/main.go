package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/koenbollen/logging"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	logger := logging.New(ctx, "example", "api")
	ctx = logging.WithLogger(ctx, logger)
	logger.Info("init")
	defer logger.Info("fin")

	http.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		logging.IgnoreRequest(r)
		w.WriteHeader(http.StatusOK)
	})

	http.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		logger := logging.GetLogger(r.Context())

		if r.URL.RawQuery == "error" {
			logger.Error("error", "err", io.ErrUnexpectedEOF)
			http.Error(w, "error", http.StatusInternalServerError)
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
			panic(http.ErrAbortHandler)
		}

		if r.URL.RawQuery == "fatal" {
			logger.Error("fatal", "err", io.ErrUnexpectedEOF)
			panic(io.ErrUnexpectedEOF)
		}

		logger.Debug("test endpoint hit", "qs", r.URL.RawQuery)
		w.WriteHeader(http.StatusCreated)
		fmt.Fprintln(w, "Hello!")
	})

	server := &http.Server{
		Addr:    ":8080",
		Handler: logging.Middleware(http.DefaultServeMux, logger),
	}
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("failed to listen or serve", "err", err)
			return
		}
	}()
	logger.Info("listening", "addr", server.Addr)

	<-ctx.Done()

	ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("failed to shutdown server", "err", err)
		return
	}
}
