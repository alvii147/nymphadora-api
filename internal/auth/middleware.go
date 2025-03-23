package auth

import (
	"context"
	"errors"
	"net/http"

	"github.com/alvii147/nymphadora-api/pkg/api"
	"github.com/alvii147/nymphadora-api/pkg/cryptocore"
	"github.com/alvii147/nymphadora-api/pkg/errutils"
	"github.com/alvii147/nymphadora-api/pkg/httputils"
)

// NewJWTAuthMiddleware creates a middleware using JWTAuthMiddleware.
func NewJWTAuthMiddleware(crypto cryptocore.Crypto) httputils.MiddlewareFunc {
	return func(next httputils.HandlerFunc) httputils.HandlerFunc {
		return JWTAuthMiddleware(next, crypto)
	}
}

// JWTAuthMiddleware parses and validates JWT from authorization header.
// If authentication fails, it returns 401.
// If authentication is successful, it sets the user UUID in context.
func JWTAuthMiddleware(next httputils.HandlerFunc, crypto cryptocore.Crypto) httputils.HandlerFunc {
	return httputils.HandlerFunc(func(w *httputils.ResponseWriter, r *http.Request) {
		token, ok := httputils.GetAuthorizationHeader(r.Header, "Bearer")
		if !ok {
			w.WriteJSON(
				api.ErrorResponse{
					Code:   api.ErrCodeMissingCredentials,
					Detail: api.ErrDetailMissingToken,
				},
				http.StatusUnauthorized,
			)

			return
		}

		claims, ok := crypto.ValidateAuthJWT(token, cryptocore.JWTTypeAccess)
		if !ok {
			w.WriteJSON(
				api.ErrorResponse{
					Code:   api.ErrCodeInvalidCredentials,
					Detail: api.ErrDetailInvalidToken,
				},
				http.StatusUnauthorized,
			)

			return
		}

		next.ServeHTTP(w, r.Clone(context.WithValue(r.Context(), AuthContextKeyUserUUID, claims.Subject)))
	})
}

// NewAPIKeyAuthMiddleware creates a middleware using APIKeyAuthMiddleware.
func NewAPIKeyAuthMiddleware(svc Service) httputils.MiddlewareFunc {
	return func(next httputils.HandlerFunc) httputils.HandlerFunc {
		return APIKeyAuthMiddleware(next, svc)
	}
}

// APIKeyAuthMiddleware authenticates a user using provided API Key.
// If authentication fails, it returns 401.
// If authentication is successful, it sets the user UUID in context.
func APIKeyAuthMiddleware(next httputils.HandlerFunc, svc Service) httputils.HandlerFunc {
	return httputils.HandlerFunc(func(w *httputils.ResponseWriter, r *http.Request) {
		rawKey, ok := httputils.GetAuthorizationHeader(r.Header, "X-API-Key")
		if !ok {
			w.WriteJSON(
				api.ErrorResponse{
					Code:   api.ErrCodeMissingCredentials,
					Detail: api.ErrDetailMissingToken,
				},
				http.StatusUnauthorized,
			)

			return
		}

		apiKey, err := svc.FindAPIKey(r.Context(), rawKey)
		if err != nil {
			switch {
			case errors.Is(err, errutils.ErrAPIKeyNotFound):
				w.WriteJSON(
					api.ErrorResponse{
						Code:   api.ErrCodeInvalidCredentials,
						Detail: api.ErrDetailInvalidToken,
					},
					http.StatusUnauthorized,
				)
			default:
				w.WriteJSON(
					api.ErrorResponse{
						Code:   api.ErrCodeInvalidCredentials,
						Detail: api.ErrDetailInvalidToken,
					},
					http.StatusUnauthorized,
				)
			}

			return
		}

		next.ServeHTTP(w, r.Clone(context.WithValue(r.Context(), AuthContextKeyUserUUID, apiKey.UserUUID)))
	})
}
