package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"sync"

	"github.com/alvii147/nymphadora-api/pkg/api"
	"github.com/alvii147/nymphadora-api/pkg/errutils"
	"github.com/alvii147/nymphadora-api/pkg/httputils"
)

// APIKeyIDParamKey is the URL parameter used for API key ID.
const APIKeyIDParamKey = "id"

// GetAPIKeyIDParam extracts the API key ID from the parameters of a request.
func GetAPIKeyIDParam(r *http.Request) (int64, error) {
	param := r.PathValue(APIKeyIDParamKey)
	apiKeyID, err := strconv.ParseInt(param, 10, 64)
	if err != nil {
		return 0, errutils.FormatErrorf(err, "strconv.Atoi failed for param %s", param)
	}

	return apiKeyID, nil
}

// handleCreateUser handles creation of new users.
// Methods: POST
// URL: /auth/users.
func (ctrl *Controller) HandleCreateUser(w *httputils.ResponseWriter, r *http.Request) {
	var req api.CreateUserRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		ctrl.logger.LogWarn(errutils.FormatError(err, "json.Decoder.Decode failed"))
		w.WriteJSON(
			api.ErrorResponse{
				Code:   api.ErrCodeInvalidRequest,
				Detail: api.ErrDetailInvalidRequestData,
			},
			http.StatusBadRequest,
		)

		return
	}

	validationPassed, validationFailures := req.Validate()
	if !validationPassed {
		ctrl.logger.LogWarn(errutils.FormatErrorf(nil, "validation failed: %v", validationFailures))
		w.WriteJSON(
			api.ErrorResponse{
				Code:               api.ErrCodeInvalidRequest,
				Detail:             api.ErrDetailInvalidRequestData,
				ValidationFailures: validationFailures,
			},
			http.StatusBadRequest,
		)

		return
	}

	var wg sync.WaitGroup
	user, err := ctrl.authService.CreateUser(
		r.Context(),
		&wg,
		req.Email,
		req.Password,
		req.FirstName,
		req.LastName,
	)
	if err != nil {
		ctrl.logger.LogError(errutils.FormatError(err))
		switch {
		case errors.Is(err, errutils.ErrUserAlreadyExists):
			w.WriteJSON(
				api.ErrorResponse{
					Code:   api.ErrCodeResourceExists,
					Detail: api.ErrDetailUserExists,
				},
				http.StatusConflict,
			)
		default:
			w.WriteJSON(
				api.ErrorResponse{
					Code:   api.ErrCodeInternalServerError,
					Detail: api.ErrDetailInternalServerError,
				},
				http.StatusInternalServerError,
			)
		}

		return
	}

	w.WriteJSON(
		api.CreateUserResponse{
			UUID:      user.UUID,
			Email:     user.Email,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		},
		http.StatusCreated,
	)
}

// HandleActivateUser handles activation of users.
// Methods: POST
// URL: /auth/users/activate.
func (ctrl *Controller) HandleActivateUser(w *httputils.ResponseWriter, r *http.Request) {
	var req api.ActivateUserRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		ctrl.logger.LogWarn(errutils.FormatError(err, "json.Decoder.Decode failed"))
		w.WriteJSON(
			api.ErrorResponse{
				Code:   api.ErrCodeInvalidRequest,
				Detail: api.ErrDetailInvalidRequestData,
			},
			http.StatusBadRequest,
		)

		return
	}

	validationPassed, validationFailures := req.Validate()
	if !validationPassed {
		ctrl.logger.LogWarn(errutils.FormatError(nil, "validation failed: %v", validationFailures))
		w.WriteJSON(
			api.ErrorResponse{
				Code:               api.ErrCodeInvalidRequest,
				Detail:             api.ErrDetailInvalidRequestData,
				ValidationFailures: validationFailures,
			},
			http.StatusBadRequest,
		)

		return
	}

	err = ctrl.authService.ActivateUser(r.Context(), req.Token)
	if err != nil {
		ctrl.logger.LogError(errutils.FormatError(err))
		switch {
		case errors.Is(err, errutils.ErrInvalidToken):
			w.WriteJSON(
				api.ErrorResponse{
					Code:   api.ErrCodeInvalidRequest,
					Detail: api.ErrDetailInvalidToken,
				},
				http.StatusBadRequest,
			)
		case errors.Is(err, errutils.ErrUserNotFound):
			w.WriteJSON(
				api.ErrorResponse{
					Code:   api.ErrCodeResourceNotFound,
					Detail: api.ErrDetailUserNotFound,
				},
				http.StatusNotFound,
			)
		default:
			w.WriteJSON(
				api.ErrorResponse{
					Code:   api.ErrCodeInternalServerError,
					Detail: api.ErrDetailInternalServerError,
				},
				http.StatusInternalServerError,
			)
		}

		return
	}

	w.WriteJSON(nil, http.StatusOK)
}

// HandleGetUserMe handles retrieval of currently authenticated user.
// Methods: GET
// URL: /auth/users/me, /api/v1/auth/users/me.
func (ctrl *Controller) HandleGetUserMe(w *httputils.ResponseWriter, r *http.Request) {
	user, err := ctrl.authService.GetAuthenticatedUser(r.Context())
	if err != nil {
		ctrl.logger.LogError(errutils.FormatError(err))
		switch {
		case errors.Is(err, errutils.ErrUserNotFound):
			w.WriteJSON(
				api.ErrorResponse{
					Code:   api.ErrCodeResourceNotFound,
					Detail: api.ErrDetailUserNotFound,
				},
				http.StatusNotFound,
			)
		default:
			w.WriteJSON(
				api.ErrorResponse{
					Code:   api.ErrCodeInternalServerError,
					Detail: api.ErrDetailInternalServerError,
				},
				http.StatusInternalServerError,
			)
		}

		return
	}

	w.WriteJSON(
		api.GetUserMeResponse{
			UUID:      user.UUID,
			Email:     user.Email,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		},
		http.StatusOK,
	)
}

// HandleUpdateUser handles updating of currently authenticated user.
// Methods: PATCH
// URL: /auth/users/me.
func (ctrl *Controller) HandleUpdateUser(w *httputils.ResponseWriter, r *http.Request) {
	var req api.UpdateUserRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		ctrl.logger.LogWarn(errutils.FormatError(err, "json.Decoder.Decode failed"))
		w.WriteJSON(
			api.ErrorResponse{
				Code:   api.ErrCodeInvalidRequest,
				Detail: api.ErrDetailInvalidRequestData,
			},
			http.StatusBadRequest,
		)

		return
	}

	validationPassed, validationFailures := req.Validate()
	if !validationPassed {
		ctrl.logger.LogWarn(errutils.FormatError(nil, "validation failed: %v", validationFailures))
		w.WriteJSON(
			api.ErrorResponse{
				Code:               api.ErrCodeInvalidRequest,
				Detail:             api.ErrDetailInvalidRequestData,
				ValidationFailures: validationFailures,
			},
			http.StatusBadRequest,
		)

		return
	}

	user, err := ctrl.authService.UpdateAuthenticatedUser(r.Context(), req.FirstName, req.LastName)
	if err != nil {
		ctrl.logger.LogError(errutils.FormatError(err))
		switch {
		case errors.Is(err, errutils.ErrUserNotFound):
			w.WriteJSON(
				api.ErrorResponse{
					Code:   api.ErrCodeResourceNotFound,
					Detail: api.ErrDetailUserNotFound,
				},
				http.StatusNotFound,
			)
		default:
			w.WriteJSON(
				api.ErrorResponse{
					Code:   api.ErrCodeInternalServerError,
					Detail: api.ErrDetailInternalServerError,
				},
				http.StatusInternalServerError,
			)
		}

		return
	}

	w.WriteJSON(
		api.UpdateUserResponse{
			UUID:      user.UUID,
			Email:     user.Email,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		},
		http.StatusOK,
	)
}

// HandleCreateJWT handles authentication of User and creation of authentication JWTs.
// Methods: POST
// URL: /auth/tokens.
func (ctrl *Controller) HandleCreateJWT(w *httputils.ResponseWriter, r *http.Request) {
	var req api.CreateTokenRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		ctrl.logger.LogWarn(errutils.FormatError(err, "json.Decoder.Decode failed"))
		w.WriteJSON(
			api.ErrorResponse{
				Code:   api.ErrCodeInvalidRequest,
				Detail: api.ErrDetailInvalidRequestData,
			},
			http.StatusBadRequest,
		)

		return
	}

	validationPassed, validationFailures := req.Validate()
	if !validationPassed {
		ctrl.logger.LogWarn(errutils.FormatError(nil, "validation failed: %v", validationFailures))
		w.WriteJSON(
			api.ErrorResponse{
				Code:               api.ErrCodeInvalidRequest,
				Detail:             api.ErrDetailInvalidRequestData,
				ValidationFailures: validationFailures,
			},
			http.StatusBadRequest,
		)

		return
	}

	accessToken, refreshToken, err := ctrl.authService.CreateJWT(
		r.Context(),
		req.Email,
		req.Password,
	)
	if err != nil {
		ctrl.logger.LogError(errutils.FormatError(err))
		switch {
		case errors.Is(err, errutils.ErrInvalidCredentials):
			w.WriteJSON(
				api.ErrorResponse{
					Code:   api.ErrCodeInvalidCredentials,
					Detail: api.ErrDetailInvalidEmailOrPassword,
				},
				http.StatusUnauthorized,
			)
		default:
			w.WriteJSON(
				api.ErrorResponse{
					Code:   api.ErrCodeInternalServerError,
					Detail: api.ErrDetailInternalServerError,
				},
				http.StatusInternalServerError,
			)
		}

		return
	}

	w.WriteJSON(
		api.CreateTokenResponse{
			Access:  accessToken,
			Refresh: refreshToken,
		},
		http.StatusCreated,
	)
}

// HandleRefreshJWT handles validation of refresh JWTs and creation of new access JWTs.
// Methods: POST
// URL: /auth/tokens/refresh.
func (ctrl *Controller) HandleRefreshJWT(w *httputils.ResponseWriter, r *http.Request) {
	var req api.RefreshTokenRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		ctrl.logger.LogWarn(errutils.FormatError(err, "json.Decoder.Decode failed"))
		w.WriteJSON(
			api.ErrorResponse{
				Code:   api.ErrCodeInvalidRequest,
				Detail: api.ErrDetailInvalidRequestData,
			},
			http.StatusBadRequest,
		)

		return
	}

	validationPassed, validationFailures := req.Validate()
	if !validationPassed {
		ctrl.logger.LogWarn(errutils.FormatError(nil, "validation failed: %v", validationFailures))
		w.WriteJSON(
			api.ErrorResponse{
				Code:               api.ErrCodeInvalidRequest,
				Detail:             api.ErrDetailInvalidRequestData,
				ValidationFailures: validationFailures,
			},
			http.StatusBadRequest,
		)

		return
	}

	accessToken, err := ctrl.authService.RefreshJWT(r.Context(), req.Refresh)
	if err != nil {
		ctrl.logger.LogError(errutils.FormatError(err))
		switch {
		case errors.Is(err, errutils.ErrInvalidToken):
			w.WriteJSON(
				api.ErrorResponse{
					Code:   api.ErrCodeInvalidRequest,
					Detail: api.ErrDetailInvalidRequestData,
				},
				http.StatusBadRequest,
			)
		default:
			w.WriteJSON(
				api.ErrorResponse{
					Code:   api.ErrCodeInternalServerError,
					Detail: api.ErrDetailInternalServerError,
				},
				http.StatusInternalServerError,
			)
		}

		return
	}

	w.WriteJSON(
		api.RefreshTokenResponse{
			Access: accessToken,
		},
		http.StatusCreated,
	)
}

// HandleValidateJWT handles validation of access JWTs.
// Methods: POST
// URL: /auth/tokens/validate.
func (ctrl *Controller) HandleValidateJWT(w *httputils.ResponseWriter, r *http.Request) {
	var req api.ValidateTokenRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		ctrl.logger.LogWarn(errutils.FormatError(err, "json.Decoder.Decode failed"))
		w.WriteJSON(
			api.ErrorResponse{
				Code:   api.ErrCodeInvalidRequest,
				Detail: api.ErrDetailInvalidRequestData,
			},
			http.StatusBadRequest,
		)

		return
	}

	validationPassed, validationFailures := req.Validate()
	if !validationPassed {
		ctrl.logger.LogWarn(errutils.FormatError(nil, "validation failed: %v", validationFailures))
		w.WriteJSON(
			api.ErrorResponse{
				Code:               api.ErrCodeInvalidRequest,
				Detail:             api.ErrDetailInvalidRequestData,
				ValidationFailures: validationFailures,
			},
			http.StatusBadRequest,
		)

		return
	}

	valid := ctrl.authService.ValidateJWT(r.Context(), req.Token)

	w.WriteJSON(
		api.ValidateTokenResponse{
			Valid: valid,
		},
		http.StatusOK,
	)
}

// HandleCreateAPIKey handles creation of new user API key.
// Methods: POST
// URL: /auth/api-keys.
func (ctrl *Controller) HandleCreateAPIKey(w *httputils.ResponseWriter, r *http.Request) {
	var req api.CreateAPIKeyRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		ctrl.logger.LogWarn(errutils.FormatError(err, "json.Decoder.Decode failed"))
		w.WriteJSON(
			api.ErrorResponse{
				Code:   api.ErrCodeInvalidRequest,
				Detail: api.ErrDetailInvalidRequestData,
			},
			http.StatusBadRequest,
		)

		return
	}

	validationPassed, validationFailures := req.Validate()
	if !validationPassed {
		ctrl.logger.LogWarn(errutils.FormatError(nil, "validation failed: %v", validationFailures))
		w.WriteJSON(
			api.ErrorResponse{
				Code:               api.ErrCodeInvalidRequest,
				Detail:             api.ErrDetailInvalidRequestData,
				ValidationFailures: validationFailures,
			},
			http.StatusBadRequest,
		)

		return
	}

	apiKey, key, err := ctrl.authService.CreateAPIKey(r.Context(), req.Name, req.ExpiresAt)
	if err != nil {
		ctrl.logger.LogError(errutils.FormatError(err))
		switch {
		case errors.Is(err, errutils.ErrAPIKeyAlreadyExists):
			w.WriteJSON(
				api.ErrorResponse{
					Code:   api.ErrCodeResourceExists,
					Detail: api.ErrDetailAPIKeyExists,
				},
				http.StatusConflict,
			)
		default:
			w.WriteJSON(
				api.ErrorResponse{
					Code:   api.ErrCodeInternalServerError,
					Detail: api.ErrDetailInternalServerError,
				},
				http.StatusInternalServerError,
			)
		}

		return
	}

	w.WriteJSON(
		api.CreateAPIKeyResponse{
			ID:        apiKey.ID,
			RawKey:    key,
			UserUUID:  apiKey.UserUUID,
			Prefix:    apiKey.Prefix,
			Name:      apiKey.Name,
			ExpiresAt: apiKey.ExpiresAt,
			CreatedAt: apiKey.CreatedAt,
			UpdatedAt: apiKey.UpdatedAt,
		},
		http.StatusCreated,
	)
}

// HandleListAPIKeys handles retrieval of API keys for currently authenticated user.
// Methods: GET
// URL: /auth/api-keys.
func (ctrl *Controller) HandleListAPIKeys(w *httputils.ResponseWriter, r *http.Request) {
	apiKeys, err := ctrl.authService.ListAPIKeys(r.Context())
	if err != nil {
		ctrl.logger.LogError(errutils.FormatError(err))
		w.WriteJSON(
			api.ErrorResponse{
				Code:   api.ErrCodeInternalServerError,
				Detail: api.ErrDetailInternalServerError,
			},
			http.StatusInternalServerError,
		)

		return
	}

	responseBody := api.ListAPIKeysResponse{
		Keys: make([]*api.GetAPIKeyResponse, len(apiKeys)),
	}

	for i, apiKey := range apiKeys {
		responseBody.Keys[i] = &api.GetAPIKeyResponse{
			ID:        apiKey.ID,
			UserUUID:  apiKey.UserUUID,
			Prefix:    apiKey.Prefix,
			Name:      apiKey.Name,
			ExpiresAt: apiKey.ExpiresAt,
			CreatedAt: apiKey.CreatedAt,
			UpdatedAt: apiKey.UpdatedAt,
		}
	}

	w.WriteJSON(responseBody, http.StatusOK)
}

// HandleUpdateAPIKey handles updating of API keys.
// Methods: PATCH
// URL: /auth/api-keys/{id}.
func (ctrl *Controller) HandleUpdateAPIKey(w *httputils.ResponseWriter, r *http.Request) {
	apiKeyID, err := GetAPIKeyIDParam(r)
	if err != nil {
		w.WriteJSON(
			api.ErrorResponse{
				Code:   api.ErrCodeInvalidRequest,
				Detail: api.ErrDetailInvalidRequestData,
			},
			http.StatusBadRequest,
		)

		return
	}

	var req api.UpdateAPIKeyRequest
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		ctrl.logger.LogWarn(errutils.FormatError(err, "json.Decoder.Decode failed"))
		w.WriteJSON(
			api.ErrorResponse{
				Code:   api.ErrCodeInvalidRequest,
				Detail: api.ErrDetailInvalidRequestData,
			},
			http.StatusBadRequest,
		)

		return
	}

	validationPassed, validationFailures := req.Validate()
	if !validationPassed {
		ctrl.logger.LogWarn(errutils.FormatError(nil, "validation failed: %v", validationFailures))
		w.WriteJSON(
			api.ErrorResponse{
				Code:               api.ErrCodeInvalidRequest,
				Detail:             api.ErrDetailInvalidRequestData,
				ValidationFailures: validationFailures,
			},
			http.StatusBadRequest,
		)

		return
	}

	apiKey, err := ctrl.authService.UpdateAPIKey(r.Context(), apiKeyID, req.Name, req.ExpiresAt)
	if err != nil {
		ctrl.logger.LogError(errutils.FormatError(err))
		switch {
		case errors.Is(err, errutils.ErrAPIKeyNotFound):
			w.WriteJSON(
				api.ErrorResponse{
					Code:   api.ErrCodeResourceNotFound,
					Detail: api.ErrDetailAPIKeyNotFound,
				},
				http.StatusNotFound,
			)
		default:
			w.WriteJSON(
				api.ErrorResponse{
					Code:   api.ErrCodeInternalServerError,
					Detail: api.ErrDetailInternalServerError,
				},
				http.StatusInternalServerError,
			)
		}

		return
	}

	w.WriteJSON(
		api.UpdateAPIKeyResponse{
			ID:        apiKey.ID,
			UserUUID:  apiKey.UserUUID,
			Prefix:    apiKey.Prefix,
			Name:      apiKey.Name,
			ExpiresAt: apiKey.ExpiresAt,
			CreatedAt: apiKey.CreatedAt,
			UpdatedAt: apiKey.UpdatedAt,
		},
		http.StatusOK,
	)
}

// HandleDeleteAPIKey handles deletion of API keys.
// Methods: DELETE
// URL: /auth/api-keys/{id}.
func (ctrl *Controller) HandleDeleteAPIKey(w *httputils.ResponseWriter, r *http.Request) {
	apiKeyID, err := GetAPIKeyIDParam(r)
	if err != nil {
		w.WriteJSON(
			api.ErrorResponse{
				Code:   api.ErrCodeInvalidRequest,
				Detail: api.ErrDetailInvalidRequestData,
			},
			http.StatusBadRequest,
		)

		return
	}

	err = ctrl.authService.DeleteAPIKey(r.Context(), apiKeyID)
	if err != nil {
		ctrl.logger.LogError(errutils.FormatError(err))
		switch {
		case errors.Is(err, errutils.ErrAPIKeyNotFound):
			w.WriteJSON(
				api.ErrorResponse{
					Code:   api.ErrCodeResourceNotFound,
					Detail: api.ErrDetailAPIKeyNotFound,
				},
				http.StatusNotFound,
			)
		default:
			w.WriteJSON(
				api.ErrorResponse{
					Code:   api.ErrCodeInternalServerError,
					Detail: api.ErrDetailInternalServerError,
				},
				http.StatusInternalServerError,
			)
		}

		return
	}

	w.WriteJSON(nil, http.StatusNoContent)
}
