package httputils

import (
	"net/http"
	"strings"
	"time"
)

const (
	// HTTPClientDefaultTimeout is the default timeout for HTTP clients.
	HTTPClientDefaultTimeout = 60 * time.Second
	// HTTPHeaderAccessControlAllowHeaders is the CORS header that lists the allowed header names.
	HTTPHeaderAccessControlAllowHeaders = "Access-Control-Allow-Headers"
	// HTTPHeaderAccessControlAllowMethods is the CORS header that lists the allowed methods.
	HTTPHeaderAccessControlAllowMethods = "Access-Control-Allow-Methods"
	// HTTPHeaderAccessControlAllowOrigin is the CORS header that specifies the allowed origin URL.
	HTTPHeaderAccessControlAllowOrigin = "Access-Control-Allow-Origin"
	// HTTPHeaderContentType is the header that defines the content type of the body.
	HTTPHeaderContentType = "Content-Type"
	// HTTPHeaderAuthorization is the header used for authentication/authorization credentials.
	HTTPHeaderAuthorization = "Authorization"
)

// HTTPClient represents HTTP clients.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// MiddlewareFunc takes in a HandlerFunc and returns a HandlerFunc.
// Middleware is used to add a processing step to handlers.
type MiddlewareFunc func(HandlerFunc) HandlerFunc

// HandlerFunc takes in a request and response writer and implements a handler.
type HandlerFunc func(w *ResponseWriter, r *http.Request)

// ServeHTTP calls the handler function.
func (f HandlerFunc) ServeHTTP(w *ResponseWriter, r *http.Request) {
	f(w, r)
}

// GetAuthorizationHeader parses HTTP authorization header.
func GetAuthorizationHeader(header http.Header, authType string) (string, bool) {
	token, ok := strings.CutPrefix(strings.TrimSpace(header.Get(HTTPHeaderAuthorization)), authType)
	if !ok {
		return "", false
	}

	return strings.TrimSpace(token), true
}

// IsHTTPSuccess determines whether or not a given status code is 2xx.
func IsHTTPSuccess(statusCode int) bool {
	return statusCode >= http.StatusOK && statusCode < http.StatusMultipleChoices
}

// NewHTTPClient creates and returns a new HTTP client.
func NewHTTPClient(modifier func(c *http.Client)) *http.Client {
	client := &http.Client{
		Timeout: HTTPClientDefaultTimeout,
	}

	if modifier != nil {
		modifier(client)
	}

	return client
}
