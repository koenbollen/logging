package logging

import (
	"bufio"
	"context"
	"math/rand"
	"net"
	"net/http"
	"strconv"
	"strings"

	"go.uber.org/zap"
)

// Middleware wraps the given next http.Handler. Requests made through this
// middleware are annotated with the given logger (to the r.Context()) and
// after the request has been service a _info_ logentry is triggered for the
// request served.
func Middleware(next http.Handler, logger *zap.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := "r" + strconv.FormatInt(rand.Int63(), 36)
		l := logger.With(zap.String("rid", requestID))
		ww := &wrapper{ResponseWriter: w}
		r = r.WithContext(context.WithValue(r.Context(), keyLogger, l))
		r = r.WithContext(context.WithValue(r.Context(), keyRequestID, requestID))
		if r.Header.Get("X-Forwarded-For") != "" {
			r.RemoteAddr = strings.TrimSpace(strings.Split(r.Header.Get("X-Forwarded-For"), ",")[0])
		}
		defer func() {
			logger.Info("served",
				zap.String("rid", requestID),
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.String("remote", r.RemoteAddr),
				zap.Int("status", ww.Status()),
			)
		}()

		next.ServeHTTP(ww, r)
	})
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
	w.ResponseWriter.WriteHeader(code)
	// Check after in case there's error handling in the wrapped ResponseWriter.
	if w.wroteHeader {
		return
	}
	w.status = code
	w.wroteHeader = true
}

func (w *wrapper) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return w.ResponseWriter.(http.Hijacker).Hijack()
}
