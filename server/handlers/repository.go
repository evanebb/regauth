package handlers

import (
	"context"
	"errors"
	"github.com/evanebb/regauth/httputil"
	"github.com/evanebb/regauth/repository"
	"github.com/evanebb/regauth/session"
	"github.com/evanebb/regauth/template"
	"github.com/google/uuid"
	"github.com/gorilla/sessions"
	"log/slog"
	"net/http"
	"strings"
)

type repositoryCtxKey struct{}

// RepositoryParser is a middleware that will look up the requested repository from the ID in the path, checks if it
// belongs to the user and sets it in the request context.
func RepositoryParser(l *slog.Logger, t template.Templater, repoStore repository.Store) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			u, ok := httputil.LoggedInUserFromContext(r.Context())
			if !ok {
				l.Error("no user in request context")
				w.WriteHeader(http.StatusInternalServerError)
				t.RenderBase(w, r, nil, "errors/500.gohtml")
				return
			}

			id, err := getUUIDFromRequest(r)
			if err != nil {
				l.Debug("could not get UUID from request", "error", err)
				w.WriteHeader(http.StatusBadRequest)
				t.RenderBase(w, r, nil, "errors/400.gohtml")
				return
			}

			repo, err := repoStore.GetByID(r.Context(), id)
			if err != nil {
				if errors.Is(err, repository.ErrNotFound) {
					l.Error("repository not found", "error", err, "repositoryId", id)
					w.WriteHeader(http.StatusNotFound)
					t.RenderBase(w, r, nil, "errors/404.gohtml")
					return
				}

				l.Error("failed to get repository", "error", err, "repositoryId", id)
				w.WriteHeader(http.StatusInternalServerError)
				t.RenderBase(w, r, nil, "errors/500.gohtml")
				return
			}

			if repo.OwnerID != u.ID {
				l.Debug("repo does not belong to user", "repositoryId", repo.ID, "userId", u.ID)
				w.WriteHeader(http.StatusNotFound)
				t.RenderBase(w, r, nil, "errors/400.gohtml")
				return
			}

			ctx := withRepository(r.Context(), repo)
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
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
// This requires the repository to have been previously set in the context, for example by RepositoryParser.
func repositoryFromContext(ctx context.Context) (repository.Repository, bool) {
	val, ok := ctx.Value(repositoryCtxKey{}).(repository.Repository)
	return val, ok
}

func Explore(l *slog.Logger, t template.Templater, s repository.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		repositories, err := s.GetAllPublic(r.Context())
		if err != nil {
			l.Error("failed to get public repositories", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			t.Render(w, r, nil, "errors/500.gohtml")
		}

		repositories = filterRepositoriesByName(repositories, r.URL.Query().Get("q"))
		paginated := PaginateRequest(r, repositories, 10)

		if shouldRenderPartials(r) {
			t.Render(w, r, paginated, "partial", "explore.partial.gohtml")
		} else {
			t.RenderBase(w, r, paginated, "explore.gohtml", "explore.partial.gohtml")
		}
	}
}

func UserRepositoryOverview(l *slog.Logger, t template.Templater, s repository.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, ok := httputil.LoggedInUserFromContext(r.Context())
		if !ok {
			l.Error("no user in request context")
			w.WriteHeader(http.StatusInternalServerError)
			t.Render(w, r, nil, "errors/500.gohtml")
			return
		}

		repositories, err := s.GetAllByOwner(r.Context(), u.ID)
		if err != nil {
			l.Error("failed to get repositories for user", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			t.Render(w, r, nil, "errors/500.gohtml")
			return
		}

		repositories = filterRepositoriesByName(repositories, r.URL.Query().Get("q"))
		paginated := PaginateRequest(r, repositories, 10)

		if shouldRenderPartials(r) {
			t.Render(w, r, paginated, "partial", "repositories/overview.partial.gohtml")
		} else {
			t.RenderBase(w, r, paginated, "repositories/overview.gohtml", "repositories/overview.partial.gohtml")
		}
	}
}

func filterRepositoriesByName(r []repository.Repository, name string) []repository.Repository {
	if name == "" {
		return r
	}

	var filtered []repository.Repository
	for _, repo := range r {
		fullName := repo.Namespace + "/" + repo.Name
		if strings.Contains(fullName, name) {
			filtered = append(filtered, repo)
		}
	}

	return filtered
}

func CreateRepositoryPage(l *slog.Logger, t template.Templater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, ok := httputil.LoggedInUserFromContext(r.Context())
		if !ok {
			l.Error("no user in request context")
			w.WriteHeader(http.StatusInternalServerError)
			t.Render(w, r, nil, "errors/500.gohtml")
		}

		data := struct {
			Namespace string
		}{
			u.Username.String(),
		}

		t.RenderBase(w, r, data, "repositories/create.gohtml")
	}
}

func CreateRepository(l *slog.Logger, t template.Templater, repoStore repository.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, ok := httputil.LoggedInUserFromContext(r.Context())
		if !ok {
			l.Error("no user in request context")
			w.WriteHeader(http.StatusInternalServerError)
			t.RenderBase(w, r, nil, "errors/500.gohtml")
			return
		}

		repo := repository.Repository{
			ID:         uuid.New(),
			Namespace:  u.Username.String(),
			Name:       r.PostFormValue("name"),
			Visibility: repository.Visibility(r.PostFormValue("visibility")),
			OwnerID:    u.ID,
		}

		err := repo.IsValid()
		if err != nil {
			l.Debug("invalid repository given", "error", err)
			w.WriteHeader(http.StatusBadRequest)
			t.RenderBase(w, r, nil, "errors/400.gohtml")
			return
		}

		err = repoStore.Create(r.Context(), repo)
		if err != nil {
			l.Error("failed to create repository", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			t.RenderBase(w, r, nil, "errors/500.gohtml")
			return
		}

		http.Redirect(w, r, "/ui/repositories/"+repo.ID.String(), http.StatusFound)
	}
}

func ViewRepository(l *slog.Logger, t template.Templater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		repo, ok := repositoryFromContext(r.Context())
		if !ok {
			l.Error("no repository in request context")
			w.WriteHeader(http.StatusInternalServerError)
			t.RenderBase(w, r, nil, "errors/500.gohtml")
			return
		}

		t.RenderBase(w, r, repo, "repositories/view.gohtml")
	}
}

func DeleteRepository(l *slog.Logger, t template.Templater, repoStore repository.Store, sessionStore sessions.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		repo, ok := repositoryFromContext(r.Context())
		if !ok {
			l.Error("no repository in request context")
			w.WriteHeader(http.StatusInternalServerError)
			t.RenderBase(w, r, nil, "errors/500.gohtml")
			return
		}

		err := repoStore.DeleteByID(r.Context(), repo.ID)
		if err != nil {
			l.Error("failed to delete repository", "error", err, "repositoryId", repo.ID)
			w.WriteHeader(http.StatusInternalServerError)
			t.RenderBase(w, r, nil, "errors/500.gohtml")
			return
		}

		s, _ := sessionStore.Get(r, "session")
		s.AddFlash(session.NewFlash(session.FlashTypeSuccess, "Successfully deleted repository!"))
		err = s.Save(r, w)
		if err != nil {
			l.Error("failed to save session", "error", err)
			http.Error(w, "authentication failed", http.StatusUnauthorized)
			return
		}

		http.Redirect(w, r, "/ui/repositories", http.StatusFound)
	}
}
