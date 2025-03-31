package server

import (
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/alvii147/nymphadora-api/internal/auth"
	"github.com/alvii147/nymphadora-api/internal/code"
	"github.com/alvii147/nymphadora-api/internal/config"
	"github.com/alvii147/nymphadora-api/internal/database"
	"github.com/alvii147/nymphadora-api/internal/templatesmanager"
	"github.com/alvii147/nymphadora-api/pkg/cryptocore"
	"github.com/alvii147/nymphadora-api/pkg/env"
	"github.com/alvii147/nymphadora-api/pkg/errutils"
	"github.com/alvii147/nymphadora-api/pkg/httputils"
	"github.com/alvii147/nymphadora-api/pkg/logging"
	"github.com/alvii147/nymphadora-api/pkg/mailclient"
	"github.com/alvii147/nymphadora-api/pkg/piston"
	"github.com/alvii147/nymphadora-api/pkg/timekeeper"
)

// HTTPServerTimeout is the timeout set for the HTTP server.
const HTTPServerTimeout = 30 * time.Second

// Controller handles server API operations.
type Controller struct {
	config       *config.Config
	timeProvider timekeeper.Provider
	router       httputils.Router
	dbPool       database.Pool
	logger       logging.Logger
	crypto       cryptocore.Crypto
	mailClient   mailclient.Client
	tmplManager  templatesmanager.Manager
	authService  auth.Service
	codeService  code.Service
}

// NewController sets up the server and returns a new controller.
func NewController() (*Controller, error) {
	cfg := &config.Config{}
	err := env.NewConfig(cfg)
	if err != nil {
		return nil, errutils.FormatError(err)
	}

	timeProvider := timekeeper.NewSystemProvider()

	router := httputils.NewRouter(
		httputils.WithRouterCORSHeaderNames("*"),
		httputils.WithRouterCORSOrigin(&cfg.FrontendBaseURL),
		httputils.WithRouterRootHeader(httputils.HTTPHeaderContentType, "application/json"),
	)
	dbPool, err := database.NewPool(
		cfg.PostgresHostname,
		cfg.PostgresPort,
		cfg.PostgresUsername,
		cfg.PostgresPassword,
		cfg.PostgresDatabaseName,
	)
	if err != nil {
		return nil, errutils.FormatError(err)
	}

	logger := logging.NewLogger(os.Stdout, os.Stderr)

	var mailClient mailclient.Client
	switch cfg.MailClientType {
	case mailclient.ClientTypeSMTP:
		mailClient = mailclient.NewSMTPClient(
			cfg.SMTPHostname,
			cfg.SMTPPort,
			cfg.SMTPUsername,
			cfg.SMTPPassword,
			timeProvider,
		)
	case mailclient.ClientTypeConsole:
		mailClient = mailclient.NewConsoleClient("support@nymphadora.com", timeProvider, os.Stdout)
	default:
		return nil, errutils.FormatErrorf(nil, "unknown mail client type %s", cfg.MailClientType)
	}

	tmplManager := templatesmanager.NewManager()

	crypto := cryptocore.NewCrypto(timeProvider, cfg.SecretKey)

	authRepository := auth.NewRepository(timeProvider)
	authService := auth.NewService(
		cfg,
		timeProvider,
		dbPool,
		logger,
		crypto,
		mailClient,
		tmplManager,
		authRepository,
	)

	pistonClient := piston.NewClient(nil, httputils.NewHTTPClient(nil))
	codeRepository := code.NewRepository(timeProvider)
	codeService := code.NewService(
		cfg,
		timeProvider,
		dbPool,
		crypto,
		mailClient,
		tmplManager,
		pistonClient,
		codeRepository,
		authRepository,
	)

	ctrl := &Controller{
		config:       cfg,
		timeProvider: timeProvider,
		router:       router,
		dbPool:       dbPool,
		logger:       logger,
		crypto:       crypto,
		mailClient:   mailClient,
		tmplManager:  tmplManager,
		authService:  authService,
		codeService:  codeService,
	}

	ctrl.route()

	return ctrl, nil
}

// Serve runs the Controller server.
func (ctrl *Controller) Serve() error {
	addr := fmt.Sprintf("%s:%d", ctrl.config.Hostname, ctrl.config.Port)
	httpSrv := &http.Server{
		Addr:              addr,
		Handler:           ctrl.router,
		ReadHeaderTimeout: HTTPServerTimeout,
	}

	ctrl.logger.LogInfo("Nymphadora API server running on", addr)
	err := httpSrv.ListenAndServe()
	if err != nil {
		return errutils.FormatErrorf(err, "httpSrv.ListenAndServe failed for addr %s", addr)
	}

	return nil
}

// ServeHTTP takes in a given request and writes a response.
func (ctrl *Controller) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctrl.router.ServeHTTP(w, r)
}

// Close closes the Controller and its connections.
func (ctrl *Controller) Close() {
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		ctrl.dbPool.Close()
	}()

	wg.Wait()
}
