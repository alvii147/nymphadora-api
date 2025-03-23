package server

import (
	"github.com/alvii147/nymphadora-api/internal/auth"
	"github.com/alvii147/nymphadora-api/pkg/logging"
)

// route sets up routes for the controller.
func (ctrl *Controller) route() {
	loggerMiddleware := logging.NewLoggerMiddleware(ctrl.logger)
	jwtMiddleware := auth.NewJWTAuthMiddleware(ctrl.crypto)
	apiKeyMiddleware := auth.NewAPIKeyAuthMiddleware(ctrl.authService)

	ctrl.router.GET("/ping", ctrl.HandlePing, loggerMiddleware)

	ctrl.router.POST("/auth/users", ctrl.HandleCreateUser, loggerMiddleware)
	ctrl.router.GET("/auth/users/me", ctrl.HandleGetUserMe, jwtMiddleware, loggerMiddleware)
	ctrl.router.PATCH("/auth/users/me", ctrl.HandleUpdateUser, jwtMiddleware, loggerMiddleware)
	ctrl.router.GET("/api/v1/auth/users/me", ctrl.HandleGetUserMe, apiKeyMiddleware, loggerMiddleware)
	ctrl.router.POST("/auth/users/activate", ctrl.HandleActivateUser, loggerMiddleware)

	ctrl.router.POST("/auth/tokens", ctrl.HandleCreateJWT, loggerMiddleware)
	ctrl.router.POST("/auth/tokens/refresh", ctrl.HandleRefreshJWT, loggerMiddleware)
	ctrl.router.POST("/auth/tokens/validate", ctrl.HandleValidateJWT, loggerMiddleware)

	ctrl.router.POST("/auth/api-keys", ctrl.HandleCreateAPIKey, jwtMiddleware, loggerMiddleware)
	ctrl.router.GET("/auth/api-keys", ctrl.HandleListAPIKeys, jwtMiddleware, loggerMiddleware)
	ctrl.router.PATCH("/auth/api-keys/{id}", ctrl.HandleUpdateAPIKey, jwtMiddleware, loggerMiddleware)
	ctrl.router.DELETE("/auth/api-keys/{id}", ctrl.HandleDeleteAPIKey, jwtMiddleware, loggerMiddleware)

	ctrl.router.POST("/code/space", ctrl.HandleCreateCodeSpace, jwtMiddleware, loggerMiddleware)
	ctrl.router.GET("/code/space", ctrl.HandleListCodeSpaces, jwtMiddleware, loggerMiddleware)
	ctrl.router.GET("/code/space/{name}", ctrl.HandleGetCodeSpace, jwtMiddleware, loggerMiddleware)
	ctrl.router.PATCH("/code/space/{name}", ctrl.HandleUpdateCodeSpace, jwtMiddleware, loggerMiddleware)
	ctrl.router.POST("/code/space/{name}/run", ctrl.HandleRunCodeSpace, jwtMiddleware, loggerMiddleware)
	ctrl.router.POST("/api/v1/code/space/{name}/run", ctrl.HandleRunCodeSpace, apiKeyMiddleware, loggerMiddleware)
	ctrl.router.GET("/code/space/{name}/access", ctrl.HandleListCodespaceUsers, jwtMiddleware, loggerMiddleware)
	ctrl.router.POST("/code/space/{name}/access", ctrl.HandleInviteCodeSpaceUser, jwtMiddleware, loggerMiddleware)
	ctrl.router.DELETE("/code/space/{name}/access", ctrl.HandleRemoveCodeSpaceUser, jwtMiddleware, loggerMiddleware)
	ctrl.router.DELETE(
		"/code/space/{name}/access/accept",
		ctrl.HandleAcceptCodeSpaceUserInvitation,
		jwtMiddleware,
		loggerMiddleware,
	)
}
