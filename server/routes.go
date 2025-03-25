package server

import (
	"github.com/evanebb/regauth/api"
	"github.com/evanebb/regauth/auth"
	"github.com/evanebb/regauth/auth/local"
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

	r.Mount("/v1", v1ApiRouter(logger, repoStore, userStore, teamStore, tokenStore, credentialsStore))

	return r
}

func v1ApiRouter(
	logger *slog.Logger,
	repoStore repository.Store,
	userStore user.Store,
	teamStore user.TeamStore,
	tokenStore token.Store,
	credentialsStore local.UserCredentialsStore,
) chi.Router {
	r := chi.NewRouter()

	r.NotFound(handlers.NotFound)

	excludedRoutes := middleware.ExcludedRoutes{{"/v1/tokens", "POST"}}
	authMiddleware := middleware.TokenAuthentication(logger, tokenStore, userStore, excludedRoutes)
	r.Use(authMiddleware)

	r.Route("/repositories", func(r chi.Router) {
		r.Post("/", handlers.CreateRepository(logger, repoStore, teamStore))
		r.Get("/", handlers.ListRepositories(logger, repoStore))

		r.Route("/{namespace}/{name}", func(r chi.Router) {
			r.Use(handlers.RepositoryParser(logger, repoStore, teamStore))
			r.Get("/", handlers.GetRepository(logger))
			r.Delete("/", handlers.DeleteRepository(logger, repoStore))
		})
	})

	r.Route("/tokens", func(r chi.Router) {
		// special case: since this is a fully API-driven application, users can create personal access tokens using
		// their username and password. Otherwise there would be no way for a user to access the rest of the API :)
		userPassMiddleware := middleware.UsernamePasswordAuthentication(logger, userStore, credentialsStore)
		r.With(userPassMiddleware).Post("/", handlers.CreateToken(logger, tokenStore))

		r.Get("/", handlers.ListTokens(logger, tokenStore))
		r.Route("/{id}", func(r chi.Router) {
			r.Use(handlers.PersonalAccessTokenParser(logger, tokenStore))
			r.Get("/", handlers.GetToken(logger))
			r.Delete("/", handlers.DeleteToken(logger, tokenStore))
		})
	})

	r.Route("/teams", func(r chi.Router) {
		r.Post("/", handlers.CreateTeam(logger, teamStore))
		r.Get("/", handlers.ListTeams(logger, teamStore))
		r.Route("/{name}", func(r chi.Router) {
			r.Use(handlers.TeamParser(logger, teamStore))
			r.Get("/", handlers.GetTeam(logger))
			r.Delete("/", handlers.DeleteTeam(logger, teamStore))

			r.Route("/members", func(r chi.Router) {
				r.Get("/", handlers.ListTeamMembers(logger, teamStore))
				r.Post("/", handlers.AddTeamMember(logger, teamStore, userStore))
				r.Delete("/{username}", handlers.RemoveTeamMember(logger, teamStore, userStore))
			})
		})
	})

	r.Route("/users", func(r chi.Router) {
		r.Use(handlers.RequireRole(logger, user.RoleAdmin))
		r.Post("/", handlers.CreateUser(logger, userStore))
		r.Get("/", handlers.ListUsers(logger, userStore))
		r.Route("/{username}", func(r chi.Router) {
			r.Use(handlers.UserParser(logger, userStore))
			r.Get("/", handlers.GetUser(logger))
			r.Delete("/", handlers.DeleteUser(logger, userStore))
			r.Post("/password", handlers.ChangeUserPassword(logger, credentialsStore))
		})
	})

	return r
}
