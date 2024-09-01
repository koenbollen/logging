package logging

import (
	"bufio"
	"context"
	"log/slog"
	"math/rand"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

// Middleware wraps the given next http.Handler. Requests made through this
// middleware are annotated with the given logger (to the r.Context()) and
// after the request has been service a _info_ logentry is triggered for the
// request served.
func Middleware(next http.Handler, logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ww := &wrapper{ResponseWriter: w}
		requestID := "r" + strconv.FormatInt(rand.Int63(), 36)
		ignored := uint32(0)
		includeQuery := uint32(0)
		l := logger.With("rid", requestID)

		ctx := r.Context()
		ctx = context.WithValue(ctx, keyLogger, l)
		ctx = context.WithValue(ctx, keyRequestID, requestID)
		ctx = context.WithValue(ctx, keyIgnoredToggle, &ignored)
		ctx = context.WithValue(ctx, keyIncludeQueryToggle, &includeQuery)
		r = r.WithContext(ctx)
		if r.Header.Get("X-Forwarded-For") != "" {
			r.RemoteAddr = strings.TrimSpace(strings.Split(r.Header.Get("X-Forwarded-For"), ",")[0])
		}

		start := time.Now()
		defer func() {
			if atomic.LoadUint32(&ignored) != 0 {
				return
			}
			took := time.Since(start)
			attrs := []slog.Attr{
				slog.String("rid", requestID),
				slog.Duration("took", took),
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.String("remote", r.RemoteAddr),
				slog.Int("status", ww.Status()),
			}
			if atomic.LoadUint32(&includeQuery) != 0 && r.URL.RawQuery != "" {
				attrs = append(attrs, slog.String("query", r.URL.RawQuery))
			}
			logger.LogAttrs(ctx, slog.LevelInfo, "served", attrs...)
		}()

		next.ServeHTTP(ww, r)
	})
}

func IgnoreRequest(r *http.Request) {
	if ignored, ok := r.Context().Value(keyIgnoredToggle).(*uint32); ok {
		atomic.StoreUint32(ignored, 1)
	}
}

func IncludeRequestQuery(r *http.Request) {
	if includeQuery, ok := r.Context().Value(keyIncludeQueryToggle).(*uint32); ok {
		atomic.StoreUint32(includeQuery, 1)
	}
}

type wrapper struct {
	http.ResponseWriter
	http.Flusher
	http.Hijacker

	status      int
	wroteHeader bool
}

func (w *wrapper) Status() int {
	return w.status
}

func (w *wrapper) Flush() {
	w.ResponseWriter.(http.Flusher).Flush()
}

func (w *wrapper) Write(p []byte) (n int, err error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(p)
}

func (w *wrapper) WriteHeader(code int) {
	if w.wroteHeader {
		return
	}
	w.ResponseWriter.WriteHeader(code)
	w.status = code
	w.wroteHeader = true
}

func (w *wrapper) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return w.ResponseWriter.(http.Hijacker).Hijack()
}
