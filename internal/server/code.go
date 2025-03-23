package server

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/alvii147/nymphadora-api/internal/code"
	"github.com/alvii147/nymphadora-api/pkg/api"
	"github.com/alvii147/nymphadora-api/pkg/errutils"
	"github.com/alvii147/nymphadora-api/pkg/httputils"
)

// CodeSpaceNameParamKey is the URL parameter used for code space name.
const CodeSpaceNameParamKey = "name"

// GetCodeSpaceNameParam extracts the code space name from the parameters of a request.
func GetCodeSpaceNameParam(r *http.Request) string {
	param := r.PathValue(CodeSpaceNameParamKey)

	return param
}

// HandleCreateCodeSpace handles creation of new code spaces.
// Methods: POST
// URL: /code/space.
func (ctrl *Controller) HandleCreateCodeSpace(w *httputils.ResponseWriter, r *http.Request) {
	var req api.CreateCodeSpaceRequest
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

	codeSpace, codeSpaceAccess, err := ctrl.codeService.CreateCodeSpace(r.Context(), req.Language)
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

	w.WriteJSON(
		api.CreateCodeSpaceResponse{
			ID:          codeSpace.ID,
			AuthorUUID:  codeSpace.AuthorUUID,
			Name:        codeSpace.Name,
			Language:    codeSpace.Language,
			Contents:    codeSpace.Contents,
			AccessLevel: codeSpaceAccess.Level.String(),
			CreatedAt:   codeSpace.CreatedAt,
			UpdatedAt:   codeSpace.UpdatedAt,
		},
		http.StatusCreated,
	)
}

// HandleListCodeSpaces handles retrieval of code spaces for currently authenticated user.
// Methods: GET
// URL: /code/space.
func (ctrl *Controller) HandleListCodeSpaces(w *httputils.ResponseWriter, r *http.Request) {
	codeSpaces, codeSpaceAccesses, err := ctrl.codeService.ListCodeSpaces(r.Context())
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

	responseBody := api.ListCodeSpacesResponse{
		CodeSpaces: make([]*api.GetCodeSpaceResponse, len(codeSpaces)),
	}

	for i, codeSpace := range codeSpaces {
		responseBody.CodeSpaces[i] = &api.GetCodeSpaceResponse{
			ID:          codeSpace.ID,
			AuthorUUID:  codeSpace.AuthorUUID,
			Name:        codeSpace.Name,
			Language:    codeSpace.Language,
			Contents:    codeSpace.Contents,
			AccessLevel: codeSpaceAccesses[i].Level.String(),
			CreatedAt:   codeSpace.CreatedAt,
			UpdatedAt:   codeSpace.UpdatedAt,
		}
	}

	w.WriteJSON(responseBody, http.StatusOK)
}

// HandleGetCodeSpace handles retrieval of a code space.
// Methods: GET
// URL: /code/space/{name}.
func (ctrl *Controller) HandleGetCodeSpace(w *httputils.ResponseWriter, r *http.Request) {
	codeSpaceName := GetCodeSpaceNameParam(r)

	codeSpace, codeSpaceAccess, err := ctrl.codeService.GetCodespace(r.Context(), codeSpaceName)
	if err != nil {
		ctrl.logger.LogError(errutils.FormatError(err))
		switch {
		case errors.Is(err, errutils.ErrCodeSpaceNotFound):
			w.WriteJSON(
				api.ErrorResponse{
					Code:   api.ErrCodeResourceNotFound,
					Detail: api.ErrDetailCodeSpaceNotFound,
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
		api.GetCodeSpaceResponse{
			ID:          codeSpace.ID,
			AuthorUUID:  codeSpace.AuthorUUID,
			Name:        codeSpace.Name,
			Language:    codeSpace.Language,
			Contents:    codeSpace.Contents,
			AccessLevel: codeSpaceAccess.Level.String(),
			CreatedAt:   codeSpace.CreatedAt,
			UpdatedAt:   codeSpace.UpdatedAt,
		},
		http.StatusOK,
	)
}

// HandleUpdateCodeSpace handles updating of code spaces.
// Methods: PATCH
// URL: /code/space/{name}.
func (ctrl *Controller) HandleUpdateCodeSpace(w *httputils.ResponseWriter, r *http.Request) {
	codeSpaceName := GetCodeSpaceNameParam(r)

	var req api.UpdateCodeSpaceRequest
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

	codeSpace, codeSpaceAccess, err := ctrl.codeService.UpdateCodeSpace(r.Context(), codeSpaceName, req.Contents)
	if err != nil {
		ctrl.logger.LogError(errutils.FormatError(err))
		switch {
		case errors.Is(err, errutils.ErrCodeSpaceNotFound):
			w.WriteJSON(
				api.ErrorResponse{
					Code:   api.ErrCodeResourceNotFound,
					Detail: api.ErrDetailCodeSpaceNotFound,
				},
				http.StatusNotFound,
			)
		case errors.Is(err, errutils.ErrCodeSpaceAccessDenied):
			w.WriteJSON(
				api.ErrorResponse{
					Code:   api.ErrCodeAccessDenied,
					Detail: api.ErrDetailCodeSpaceAccessDenied,
				},
				http.StatusForbidden,
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
		api.UpdateCodeSpaceResponse{
			ID:          codeSpace.ID,
			AuthorUUID:  codeSpace.AuthorUUID,
			Name:        codeSpace.Name,
			Language:    codeSpace.Language,
			Contents:    codeSpace.Contents,
			AccessLevel: codeSpaceAccess.Level.String(),
			CreatedAt:   codeSpace.CreatedAt,
			UpdatedAt:   codeSpace.UpdatedAt,
		},
		http.StatusOK,
	)
}

// HandleDeleteCodeSpace handles deletion of code spaces.
// Methods: DELETE
// URL: /code/space/{name}.
func (ctrl *Controller) HandleDeleteCodeSpace(w *httputils.ResponseWriter, r *http.Request) {
	codeSpaceName := GetCodeSpaceNameParam(r)

	err := ctrl.codeService.DeleteCodeSpace(r.Context(), codeSpaceName)
	if err != nil {
		ctrl.logger.LogError(errutils.FormatError(err))
		switch {
		case errors.Is(err, errutils.ErrCodeSpaceNotFound):
			w.WriteJSON(
				api.ErrorResponse{
					Code:   api.ErrCodeResourceNotFound,
					Detail: api.ErrDetailCodeSpaceNotFound,
				},
				http.StatusNotFound,
			)
		case errors.Is(err, errutils.ErrCodeSpaceAccessDenied):
			w.WriteJSON(
				api.ErrorResponse{
					Code:   api.ErrCodeAccessDenied,
					Detail: api.ErrDetailCodeSpaceAccessDenied,
				},
				http.StatusForbidden,
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

// HandleRunCodeSpace handles running of code spaces.
// Methods: POST
// URL: /code/space/{name}/run, /api/v1/code/space/{name}/run.
func (ctrl *Controller) HandleRunCodeSpace(w *httputils.ResponseWriter, r *http.Request) {
	codeSpaceName := GetCodeSpaceNameParam(r)

	pistonResponse, err := ctrl.codeService.RunCodeSpace(r.Context(), codeSpaceName)
	if err != nil {
		ctrl.logger.LogError(errutils.FormatError(err))
		switch {
		case errors.Is(err, errutils.ErrCodeSpaceNotFound):
			w.WriteJSON(
				api.ErrorResponse{
					Code:   api.ErrCodeResourceNotFound,
					Detail: api.ErrDetailCodeSpaceNotFound,
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

	resp := api.RunCodeSpaceResponse{
		Run: api.RunCodeSpaceResultsResponse{
			Stdout: pistonResponse.Compile.Stdout,
			Stderr: pistonResponse.Compile.Stderr,
			Code:   pistonResponse.Compile.Code,
			Signal: pistonResponse.Compile.Signal,
		},
	}

	if pistonResponse.Compile != nil {
		resp.Compile = &api.RunCodeSpaceResultsResponse{
			Stdout: pistonResponse.Compile.Stdout,
			Stderr: pistonResponse.Compile.Stderr,
			Code:   pistonResponse.Compile.Code,
			Signal: pistonResponse.Compile.Signal,
		}
	}

	w.WriteJSON(resp, http.StatusOK)
}

// HandleListCodespaceUsers handles retrieval of users with access to a code space.
// Methods: GET
// URL: /code/space/{name}/access.
func (ctrl *Controller) HandleListCodespaceUsers(w *httputils.ResponseWriter, r *http.Request) {
	codeSpaceName := GetCodeSpaceNameParam(r)

	users, codeSpaceAccesses, err := ctrl.codeService.ListCodespaceUsers(r.Context(), codeSpaceName)
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

	responseBody := api.ListCodespaceUsersResponse{
		Users: make([]*api.GetCodespaceUserResponse, len(users)),
	}

	for i, user := range users {
		responseBody.Users[i] = &api.GetCodespaceUserResponse{
			UserUUID:    user.UUID,
			FirstName:   user.FirstName,
			LastName:    user.LastName,
			CodeSpaceID: codeSpaceAccesses[i].CodeSpaceID,
			AccessLevel: codeSpaceAccesses[i].Level.String(),
		}
	}

	w.WriteJSON(responseBody, http.StatusOK)
}

// HandleInviteCodeSpaceUser handles invitation of users to code spaces.
// Methods: POST
// URL: /code/space/{name}/access.
func (ctrl *Controller) HandleInviteCodeSpaceUser(w *httputils.ResponseWriter, r *http.Request) {
	codeSpaceName := GetCodeSpaceNameParam(r)

	var req api.InviteCodeSpaceUserRequest
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

	err = ctrl.codeService.InviteCodeSpaceUser(
		r.Context(),
		codeSpaceName,
		req.Email,
		code.GetAccessLevelFromString(req.AccessLevel),
	)
	if err != nil {
		ctrl.logger.LogError(errutils.FormatError(err))
		switch {
		case errors.Is(err, errutils.ErrCodeSpaceNotFound):
			w.WriteJSON(
				api.ErrorResponse{
					Code:   api.ErrCodeResourceNotFound,
					Detail: api.ErrDetailCodeSpaceNotFound,
				},
				http.StatusNotFound,
			)
		case errors.Is(err, errutils.ErrCodeSpaceAccessDenied):
			w.WriteJSON(
				api.ErrorResponse{
					Code:   api.ErrCodeAccessDenied,
					Detail: api.ErrDetailCodeSpaceAccessDenied,
				},
				http.StatusForbidden,
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

	w.WriteJSON(nil, http.StatusCreated)
}

// HandleRemoveCodeSpaceUser handles removal of user access to code spaces.
// Methods: DELETE
// URL: /code/space/{name}/access.
func (ctrl *Controller) HandleRemoveCodeSpaceUser(w *httputils.ResponseWriter, r *http.Request) {
	codeSpaceName := GetCodeSpaceNameParam(r)

	var req api.RemoveCodeSpaceUserRequest
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

	err = ctrl.codeService.RemoveCodeSpaceUser(r.Context(), codeSpaceName, req.UserUUID)
	if err != nil {
		ctrl.logger.LogError(errutils.FormatError(err))
		switch {
		case errors.Is(err, errutils.ErrCodeSpaceNotFound):
			w.WriteJSON(
				api.ErrorResponse{
					Code:   api.ErrCodeResourceNotFound,
					Detail: api.ErrDetailCodeSpaceNotFound,
				},
				http.StatusNotFound,
			)
		case errors.Is(err, errutils.ErrCodeSpaceAccessDenied):
			w.WriteJSON(
				api.ErrorResponse{
					Code:   api.ErrCodeAccessDenied,
					Detail: api.ErrDetailCodeSpaceAccessDenied,
				},
				http.StatusForbidden,
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

// HandleAcceptCodeSpaceUserInvitation handles acceptance of code space user invitations.
// Methods: POST
// URL: /code/space/{name}/access/accept.
func (ctrl *Controller) HandleAcceptCodeSpaceUserInvitation(w *httputils.ResponseWriter, r *http.Request) {
	codeSpaceName := GetCodeSpaceNameParam(r)

	var req api.AcceptCodeSpaceUserInvitationRequest
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

	codeSpace, codeSpaceAccess, err := ctrl.codeService.AcceptCodeSpaceUserInvitation(r.Context(), codeSpaceName, req.Token)
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
		api.AcceptCodeSpaceUserInvitationResponse{
			ID:          codeSpace.ID,
			AuthorUUID:  codeSpace.AuthorUUID,
			Name:        codeSpace.Name,
			Language:    codeSpace.Language,
			Contents:    codeSpace.Contents,
			AccessLevel: codeSpaceAccess.Level.String(),
			CreatedAt:   codeSpace.CreatedAt,
			UpdatedAt:   codeSpace.UpdatedAt,
		},
		http.StatusCreated,
	)
}
