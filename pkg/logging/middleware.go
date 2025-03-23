package logging

import (
	"net/http"

	"github.com/alvii147/nymphadora-api/pkg/httputils"
)

// LogTraffic logs HTTP traffic, including http method, URL, protocol, and status code.
func LogTraffic(logger Logger, w *httputils.ResponseWriter, request *http.Request) {
	switch {
	case w.StatusCode < http.StatusBadRequest:
		logger.LogInfo(request.Method, request.URL, request.Proto, w.StatusCode, http.StatusText(w.StatusCode))
	case w.StatusCode < http.StatusInternalServerError:
		logger.LogWarn(request.Method, request.URL, request.Proto, w.StatusCode, http.StatusText(w.StatusCode))
	default:
		logger.LogError(request.Method, request.URL, request.Proto, w.StatusCode, http.StatusText(w.StatusCode))
	}
}

// NewLoggerMiddleware creates a middleware using LoggerMiddleware.
func NewLoggerMiddleware(logger Logger) httputils.MiddlewareFunc {
	return func(next httputils.HandlerFunc) httputils.HandlerFunc {
		return LoggerMiddleware(next, logger)
	}
}

// LoggerMiddleware handles logging of HTTP traffic.
func LoggerMiddleware(next httputils.HandlerFunc, logger Logger) httputils.HandlerFunc {
	return httputils.HandlerFunc(func(w *httputils.ResponseWriter, r *http.Request) {
		defer func() {
			LogTraffic(logger, w, r)
		}()

		next.ServeHTTP(w, r)
	})
}
