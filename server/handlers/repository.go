package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/evanebb/regauth/repository"
	"github.com/evanebb/regauth/server/middleware"
	"github.com/evanebb/regauth/server/response"
	"github.com/evanebb/regauth/user"
	"github.com/evanebb/regauth/util"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"log/slog"
	"net/http"
)

type repositoryCtxKey struct{}

func RepositoryParser(l *slog.Logger, repoStore repository.Store, teamStore user.TeamStore) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			u, ok := middleware.AuthenticatedUserFromContext(r.Context())
			if !ok {
				l.ErrorContext(r.Context(), "could not parse user from request context")
				response.WriteJSONError(w, http.StatusUnauthorized, "authentication failed")
				return
			}

			namespace, name := chi.URLParam(r, "namespace"), chi.URLParam(r, "name")
			if namespace == "" || name == "" {
				response.WriteJSONError(w, http.StatusBadRequest, "no repository namespace or name given")
				return
			}

			repo, err := repoStore.GetByNamespaceAndName(r.Context(), namespace, name)
			if err != nil {
				if errors.Is(err, repository.ErrNotFound) {
					response.WriteJSONError(w, http.StatusNotFound, "repository not found")
					return
				}

				l.ErrorContext(r.Context(), "could not get repository", slog.Any("error", err))
				response.WriteJSONError(w, http.StatusInternalServerError, "internal server error")
				return
			}

			teams, err := teamStore.GetAllByUser(r.Context(), u.ID)
			if err != nil {
				l.ErrorContext(r.Context(), "failed to get teams for user", slog.Any("error", err))
				response.WriteJSONError(w, http.StatusInternalServerError, "internal server error")
				return
			}

			authorized := false
			for _, team := range teams {
				if repo.Namespace == string(team.Name) {
					authorized = true
				}
			}

			if repo.Namespace == string(u.Username) {
				authorized = true
			}

			if !authorized {
				response.WriteJSONError(w, http.StatusForbidden, "not authorized for given namespace")
				return
			}

			ctx := withRepository(r.Context(), repo)
			next.ServeHTTP(w, r.WithContext(ctx))
		}
		return http.HandlerFunc(fn)
	}
}

// withRepository sets the given repository.Repository in the context.
// Use repositoryFromContext to retrieve the repository.
func withRepository(ctx context.Context, r repository.Repository) context.Context {
	return context.WithValue(ctx, repositoryCtxKey{}, r)
}

// repositoryFromContext parses the current repository.Repository from the given request context.
// This requires the repository to have been previously set in the context, for example by the RepositoryParser middleware.
func repositoryFromContext(ctx context.Context) (repository.Repository, bool) {
	val, ok := ctx.Value(repositoryCtxKey{}).(repository.Repository)
	return val, ok
}

func CreateRepository(l *slog.Logger, repoStore repository.Store, teamStore user.TeamStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, ok := middleware.AuthenticatedUserFromContext(r.Context())
		if !ok {
			l.ErrorContext(r.Context(), "could not parse user from request context")
			response.WriteJSONError(w, http.StatusUnauthorized, "authenticated failed")
			return
		}

		var repo repository.Repository
		if err := json.NewDecoder(r.Body).Decode(&repo); err != nil {
			response.WriteJSONError(w, http.StatusBadRequest, "invalid JSON body given")
			return
		}

		teams, err := teamStore.GetAllByUser(r.Context(), u.ID)
		if err != nil {
			l.ErrorContext(r.Context(), "failed to get teams for user", slog.Any("error", err))
			response.WriteJSONError(w, http.StatusInternalServerError, "internal server error")
			return
		}

		authorized := false
		for _, team := range teams {
			if repo.Namespace == string(team.Name) {
				authorized = true
			}
		}

		if repo.Namespace == string(u.Username) {
			authorized = true
		}

		if !authorized {
			response.WriteJSONError(w, http.StatusForbidden, "not authorized to create repository in given namespace")
			return
		}

		_, err = repoStore.GetByNamespaceAndName(r.Context(), repo.Namespace, string(repo.Name))
		if err == nil {
			response.WriteJSONError(w, http.StatusBadRequest, "repository already exists")
			return
		}

		if !errors.Is(err, repository.ErrNotFound) {
			l.Error("could not get repository", "error", err)
			response.WriteJSONError(w, http.StatusInternalServerError, "internal server error")
			return
		}

		repo.ID, err = uuid.NewV7()
		if err != nil {
			l.Error("could not generate UUID", "error", err)
			response.WriteJSONError(w, http.StatusInternalServerError, "internal server error")
			return
		}

		if err := repo.IsValid(); err != nil {
			response.WriteJSONError(w, http.StatusBadRequest, err.Error())
			return
		}

		if err := repoStore.Create(r.Context(), repo); err != nil {
			l.Error("could not create repository", "error", err)
			response.WriteJSONError(w, http.StatusInternalServerError, "internal server error")
			return
		}

		response.WriteJSONResponse(w, http.StatusOK, repo)
	}
}

func ListRepositories(l *slog.Logger, s repository.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, ok := middleware.AuthenticatedUserFromContext(r.Context())
		if !ok {
			l.ErrorContext(r.Context(), "could not parse user from request context")
			response.WriteJSONError(w, http.StatusUnauthorized, "authentication failed")
			return
		}

		repos, err := s.GetAllByUser(r.Context(), u.ID)
		if err != nil {
			l.ErrorContext(r.Context(), "could not get repositories for user", slog.Any("error", err))
			response.WriteJSONError(w, http.StatusInternalServerError, "internal server error")
			return
		}

		response.WriteJSONResponse(w, http.StatusOK, util.NilSliceToEmpty(repos))
	}
}

func GetRepository(l *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		repo, ok := repositoryFromContext(r.Context())
		if !ok {
			l.ErrorContext(r.Context(), "could not parse repository from request context")
			response.WriteJSONError(w, http.StatusInternalServerError, "internal server error")
			return
		}

		response.WriteJSONResponse(w, http.StatusOK, repo)
	}
}

func DeleteRepository(l *slog.Logger, s repository.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		repo, ok := repositoryFromContext(r.Context())
		if !ok {
			l.ErrorContext(r.Context(), "could not parse repository from request context")
			response.WriteJSONError(w, http.StatusInternalServerError, "internal server error")
			return
		}

		if err := s.DeleteByID(r.Context(), repo.ID); err != nil {
			l.ErrorContext(r.Context(), "could not delete repository", slog.Any("error", err))
			response.WriteJSONError(w, http.StatusInternalServerError, "internal server error")
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
