package main

import (
	"net/http"
	"time"

	"github.com/go-kit/kit/log"

	"github.com/peterbourgon/sympatico/internal/ctxlog"
)

type loggingMiddleware struct {
	next   http.Handler
	logger log.Logger
}

func newLoggingMiddleware(next http.Handler, logger log.Logger) *loggingMiddleware {
	return &loggingMiddleware{next, logger}
}

func (mw *loggingMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var (
		iw          = &interceptingWriter{http.StatusOK, w}
		ctx, ctxlog = ctxlog.New(r.Context(), "http_method", r.Method, "http_path", r.URL.Path)
	)

	defer func(begin time.Time) {
		ctxlog.Log("http_status_code", iw.code, "http_duration", time.Since(begin))
		mw.logger.Log(ctxlog.Keyvals()...)
	}(time.Now())

	mw.next.ServeHTTP(iw, r.WithContext(ctx))
}

type interceptingWriter struct {
	code int
	http.ResponseWriter
}

func (iw *interceptingWriter) WriteHeader(code int) {
	iw.code = code
	iw.ResponseWriter.WriteHeader(code)
}
