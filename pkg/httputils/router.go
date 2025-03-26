package httputils

import (
	"fmt"
	"net/http"
	"strings"
)

// Router represents a handler that can register handlers under specific HTTP methods.
type Router interface {
	GET(pattern string, handler HandlerFunc, middleware ...MiddlewareFunc)
	POST(pattern string, handler HandlerFunc, middleware ...MiddlewareFunc)
	PUT(pattern string, handler HandlerFunc, middleware ...MiddlewareFunc)
	PATCH(pattern string, handler HandlerFunc, middleware ...MiddlewareFunc)
	DELETE(pattern string, handler HandlerFunc, middleware ...MiddlewareFunc)
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}

// router implements Router.
type router struct {
	*http.ServeMux
	corsHeaderNames []string
	corsMethods     map[string][]string
	corsOrigin      *string
	rootHeaders     map[string]string
}

// WithRouterCORSHeaderNames can be used with NewRouter to define a list of allowed headers.
func WithRouterCORSHeaderNames(headerNames ...string) func(r *router) {
	return func(r *router) {
		r.corsHeaderNames = headerNames
	}
}

// WithRouterCORSOrigin can be used with NewRouter to specify an allowed origin URL.
func WithRouterCORSOrigin(origin *string) func(r *router) {
	return func(r *router) {
		r.corsOrigin = origin
	}
}

// WithRouterCORSOrigin can be used with NewRouter to add top-level headers.
func WithRouterRootHeader(name string, value string) func(r *router) {
	return func(r *router) {
		r.rootHeaders[name] = value
	}
}

// NewRouter creates and returns a new router.
func NewRouter(opts ...func(r *router)) *router {
	r := &router{
		ServeMux:        http.NewServeMux(),
		corsHeaderNames: []string{"*"},
		corsMethods:     make(map[string][]string),
		corsOrigin:      nil,
		rootHeaders:     make(map[string]string),
	}

	for _, opt := range opts {
		opt(r)
	}

	return r
}

// handleOptions is a handler for the OPTIONS method.
// This handle just returns a 200 status code.
func (r *router) handleOptions(w *ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)
}

// rootMiddleware handles setting of top-level headers.
func (r *router) rootMiddleware(next HandlerFunc, pattern string) HandlerFunc {
	return HandlerFunc(func(w *ResponseWriter, req *http.Request) {
		if r.corsHeaderNames != nil {
			w.Header().Set(HTTPHeaderAccessControlAllowHeaders, strings.Join(r.corsHeaderNames, ", "))
		}

		corsMethods := append(r.corsMethods[pattern], http.MethodOptions)
		w.Header().Set(HTTPHeaderAccessControlAllowMethods, strings.Join(corsMethods, ", "))

		if r.corsOrigin != nil {
			w.Header().Set(HTTPHeaderAccessControlAllowOrigin, *r.corsOrigin)
		}

		for headerName, headerValue := range r.rootHeaders {
			w.Header().Set(headerName, headerValue)
		}

		next.ServeHTTP(w, req)
	})
}

// chainMiddleware wraps a handler with a chain of middleware.
func (r *router) chainMiddleware(handler HandlerFunc, middleware ...MiddlewareFunc) HandlerFunc {
	h := handler
	for _, mw := range middleware {
		h = mw(h)
	}

	return h
}

// addRoute registers a pattern and HTTP method with a handler and set of middleware.
// If not already done, it also registers the pattern for the OPTIONS method.
func (r *router) addRoute(
	method string,
	pattern string,
	handler HandlerFunc,
	middleware ...MiddlewareFunc,
) {
	_, registered := r.corsMethods[pattern]
	r.corsMethods[pattern] = append(r.corsMethods[pattern], method)

	if !registered {
		r.Handle(
			fmt.Sprintf("%s %s", http.MethodOptions, pattern),
			ResponseWriterMiddleware(
				r.rootMiddleware(
					r.handleOptions,
					pattern,
				),
			),
		)
	}

	r.Handle(
		fmt.Sprintf("%s %s", method, pattern),
		ResponseWriterMiddleware(
			r.rootMiddleware(
				r.chainMiddleware(
					handler,
					middleware...,
				),
				pattern,
			),
		),
	)
}

// GET wraps a handler with a set of middleware
// and registers it for a given pattern and HTTP GET method.
func (r *router) GET(
	pattern string,
	handler HandlerFunc,
	middleware ...MiddlewareFunc,
) {
	r.addRoute(http.MethodGet, pattern, handler, middleware...)
}

// POST wraps a handler with a set of middleware
// and registers it for a given pattern and HTTP POST method.
func (r *router) POST(
	pattern string,
	handler HandlerFunc,
	middleware ...MiddlewareFunc,
) {
	r.addRoute(http.MethodPost, pattern, handler, middleware...)
}

// PUT wraps a handler with a set of middleware
// and registers it for a given pattern and HTTP PUT method.
func (r *router) PUT(
	pattern string,
	handler HandlerFunc,
	middleware ...MiddlewareFunc,
) {
	r.addRoute(http.MethodPut, pattern, handler, middleware...)
}

// PATCH wraps a handler with a set of middleware
// and registers it for a given pattern and HTTP PATCH method.
func (r *router) PATCH(
	pattern string,
	handler HandlerFunc,
	middleware ...MiddlewareFunc,
) {
	r.addRoute(http.MethodPatch, pattern, handler, middleware...)
}

// DELETE wraps a handler with a set of middleware
// and registers it for a given pattern and HTTP DELETE method.
func (r *router) DELETE(
	pattern string,
	handler HandlerFunc,
	middleware ...MiddlewareFunc,
) {
	r.addRoute(http.MethodDelete, pattern, handler, middleware...)
}
