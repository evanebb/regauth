package server

import (
	"github.com/evanebb/regauth/api"
	"github.com/evanebb/regauth/auth"
	"github.com/evanebb/regauth/auth/local"
	"github.com/evanebb/regauth/oas"
	"github.com/evanebb/regauth/repository"
	"github.com/evanebb/regauth/server/handlers"
	"github.com/evanebb/regauth/server/middleware"
	"github.com/evanebb/regauth/token"
	"github.com/evanebb/regauth/user"
	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"log/slog"
	"net/http"
)

func baseRouter(
	logger *slog.Logger,
	repoStore repository.Store,
	userStore user.Store,
	teamStore user.TeamStore,
	tokenStore token.Store,
	credentialsStore local.UserCredentialsStore,
	authenticator auth.Authenticator,
	authorizer auth.Authorizer,
	accessTokenConfig auth.AccessTokenConfiguration,
	tokenPrefix string,
) chi.Router {
	r := chi.NewRouter()

	loggerMiddleware := middleware.Logger(logger)
	r.Use(chiMiddleware.RequestID, loggerMiddleware, chiMiddleware.Recoverer)

	// Note: if more extensive (and sensitive) information is ever added to the /health endpoint, it should listen on a
	// separate port from the main server, so that clients cannot directly access it!
	r.Get("/health", handlers.Health())

	r.Handle("/", http.RedirectHandler("/reference/", http.StatusMovedPermanently))
	r.Handle("/reference", http.RedirectHandler("/reference/", http.StatusMovedPermanently))
	r.Handle("/reference/*", http.StripPrefix("/reference/", http.FileServer(http.FS(api.Files))))

	r.Handle("/token", handlers.GenerateRegistryToken(logger, authenticator, authorizer, accessTokenConfig))

	handler := handlers.NewHandler(logger, repoStore, userStore, teamStore, tokenStore, credentialsStore, tokenPrefix)
	securityHandler := handlers.NewSecurityHandler(logger, tokenStore, userStore, credentialsStore)
	apiServer, err := oas.NewServer(handler, securityHandler, oas.WithNotFound(handlers.NotFound))
	if err != nil {
		panic(err)
	}

	r.Mount("/v1/", apiServer)

	return r
}
