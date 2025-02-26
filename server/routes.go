package server

import (
	"github.com/evanebb/regauth/auth"
	"github.com/evanebb/regauth/auth/local"
	"github.com/evanebb/regauth/pat"
	"github.com/evanebb/regauth/repository"
	"github.com/evanebb/regauth/server/handlers"
	"github.com/evanebb/regauth/server/middleware"
	"github.com/evanebb/regauth/template"
	"github.com/evanebb/regauth/user"
	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-jose/go-jose/v4"
	"github.com/gorilla/sessions"
	"log/slog"
	"net/http"
)

func addRoutes(
	router chi.Router,
	logger *slog.Logger,
	sessionStore sessions.Store,
	templater template.Templater,
	repoStore repository.Store,
	userStore user.Store,
	patStore pat.Store,
	authUserStore local.AuthUserStore,
	authenticator auth.Authenticator,
	authorizer auth.Authorizer,
	tokenIssuer, tokenService string,
	tokenSigner jose.Signer,
	registryHost string,
) {
	loggerMiddleware := middleware.Logger(logger)
	router.Use(chiMiddleware.RequestID, middleware.MethodOverride, loggerMiddleware, chiMiddleware.Recoverer)
	router.Handle("/token", handlers.GenerateToken(logger, authenticator, authorizer, tokenIssuer, tokenService, tokenSigner))

	// Redirect the index to the UI by default
	router.Handle("/", http.RedirectHandler("/ui", http.StatusMovedPermanently))

	// UI routes
	router.Route("/ui", func(r chi.Router) {
		r.NotFound(handlers.NotFound(templater))

		r.Use(handlers.UserSessionParser(sessionStore, userStore))

		// Routes that do not necessarily require authentication
		r.Get("/", handlers.Index(templater))
		r.Get("/login", handlers.LoginPage(templater))
		r.Post("/login", handlers.Login(logger, authUserStore, userStore, sessionStore))
		r.Get("/explore", handlers.Explore(logger, templater, repoStore))

		// Authenticated routes
		r.Route("/", func(r chi.Router) {
			r.Use(handlers.UserAuth)
			r.Get("/logout", handlers.Logout(logger, sessionStore))

			r.Route("/account", func(r chi.Router) {
				r.Get("/", handlers.ManageAccount(templater))
				r.Route("/tokens", func(r chi.Router) {
					r.Get("/", handlers.TokenOverview(logger, templater, patStore))
					r.Get("/create", handlers.CreateTokenPage(templater))
					r.Post("/", handlers.CreateToken(logger, templater, patStore, registryHost))
					r.Route("/{id}", func(r chi.Router) {
						r.Use(handlers.PersonalAccessTokenParser(logger, templater, patStore))
						r.Get("/", handlers.ViewToken(logger, templater, patStore))
						r.Delete("/", handlers.DeleteToken(logger, templater, patStore, sessionStore))
					})
				})
				r.Route("/users", func(r chi.Router) {
					r.Use(handlers.RequireRole(user.RoleAdmin))
					r.Get("/", handlers.UserOverview(logger, templater, userStore))
					r.Get("/create", handlers.CreateUserPage(templater))
					r.Post("/", handlers.CreateUser(logger, templater, userStore, authUserStore))
					r.Route("/{id}", func(r chi.Router) {
						r.Use(handlers.UserParser(logger, templater, userStore))
						r.Get("/", handlers.ViewUser(logger, templater, userStore))
						r.Delete("/", handlers.DeleteUser(logger, templater, userStore, authUserStore, sessionStore))
						r.Post("/reset-password", handlers.ResetUserPassword(logger, templater, authUserStore, sessionStore))
					})
				})
			})

			r.Route("/repositories", func(r chi.Router) {
				r.Get("/", handlers.UserRepositoryOverview(logger, templater, repoStore))
				r.Get("/create", handlers.CreateRepositoryPage(logger, templater))
				r.Post("/", handlers.CreateRepository(logger, templater, repoStore))
				r.Route("/{id}", func(r chi.Router) {
					r.Use(handlers.RepositoryParser(logger, templater, repoStore))
					r.Get("/", handlers.ViewRepository(logger, templater))
					r.Delete("/", handlers.DeleteRepository(logger, templater, repoStore, sessionStore))
				})
			})
		})
	})
}
