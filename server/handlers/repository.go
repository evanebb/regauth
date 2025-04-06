package handlers

import (
	"context"
	"errors"
	"fmt"
	"github.com/evanebb/regauth/oas"
	"github.com/evanebb/regauth/repository"
	"github.com/evanebb/regauth/user"
	"github.com/google/uuid"
	"log/slog"
	"net/http"
	"slices"
	"time"
)

type RepositoryHandler struct {
	logger    *slog.Logger
	repoStore repository.Store
	teamStore user.TeamStore
	oas.UnimplementedHandler
}

func (h RepositoryHandler) CreateRepository(ctx context.Context, req *oas.RepositoryRequest) (*oas.RepositoryResponse, error) {
	u, ok := AuthenticatedUserFromContext(ctx)
	if !ok {
		h.logger.ErrorContext(ctx, "could not parse user from request context")
		return nil, newErrorResponse(http.StatusInternalServerError, "internal server error")
	}

	authorizedNamespaces, err := h.getUserNamespaces(ctx, u)
	if err != nil {
		h.logger.ErrorContext(ctx, "failed to get namespaces for user", slog.Any("error", err))
		return nil, newErrorResponse(http.StatusInternalServerError, "internal server error")
	}

	if !slices.Contains(authorizedNamespaces, req.Namespace) {
		return nil, newErrorResponse(http.StatusForbidden, "not authorized for given namespace")
	}

	id, err := uuid.NewV7()
	if err != nil {
		h.logger.ErrorContext(ctx, "could not generate UUID", "error", err)
		return nil, newErrorResponse(http.StatusInternalServerError, "internal server error")
	}

	repo := repository.Repository{
		ID:         id,
		Namespace:  req.Namespace,
		Name:       repository.Name(req.Name),
		Visibility: repository.Visibility(req.Visibility),
		CreatedAt:  time.Now(),
	}

	if err := repo.IsValid(); err != nil {
		return nil, newErrorResponse(http.StatusBadRequest, err.Error())
	}

	if err := h.repoStore.Create(ctx, repo); err != nil {
		h.logger.ErrorContext(ctx, "could not create repository", "error", err)
		return nil, newErrorResponse(http.StatusInternalServerError, "internal server error")
	}

	resp := convertToRepositoryResponse(repo)
	return &resp, nil
}

func (h RepositoryHandler) ListRepositories(ctx context.Context) ([]oas.RepositoryResponse, error) {
	u, ok := AuthenticatedUserFromContext(ctx)
	if !ok {
		h.logger.ErrorContext(ctx, "could not parse user from request context")
		return nil, newErrorResponse(http.StatusInternalServerError, "internal server error")
	}

	namespaces, err := h.getUserNamespaces(ctx, u)
	if err != nil {
		h.logger.ErrorContext(ctx, "failed to get namespaces for user", slog.Any("error", err))
		return nil, newErrorResponse(http.StatusInternalServerError, "internal server error")
	}

	repos, err := h.repoStore.GetAllByNamespace(ctx, namespaces...)
	if err != nil {
		h.logger.ErrorContext(ctx, "failed to get repositories for user", slog.Any("error", err))
		return nil, newErrorResponse(http.StatusInternalServerError, "internal server error")
	}

	return convertSlice(repos, convertToRepositoryResponse), nil
}

func (h RepositoryHandler) GetRepository(ctx context.Context, params oas.GetRepositoryParams) (*oas.RepositoryResponse, error) {
	repo, err := h.getRepositoryFromRequest(ctx, params.Namespace, params.Name)
	if err != nil {
		return nil, err
	}

	resp := convertToRepositoryResponse(repo)
	return &resp, nil
}

func (h RepositoryHandler) DeleteRepository(ctx context.Context, params oas.DeleteRepositoryParams) error {
	repo, err := h.getRepositoryFromRequest(ctx, params.Namespace, params.Name)
	if err != nil {
		return err
	}

	if err := h.repoStore.DeleteByID(ctx, repo.ID); err != nil {
		h.logger.ErrorContext(ctx, "could not delete repository", slog.Any("error", err))
		return newErrorResponse(http.StatusInternalServerError, "internal server error")
	}

	return nil
}

func (h RepositoryHandler) getUserNamespaces(ctx context.Context, u user.User) ([]string, error) {
	teams, err := h.teamStore.GetAllByUser(ctx, u.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get teams for user: %w", err)
	}

	nsCount := len(teams) + 1
	namespaces := make([]string, nsCount)

	namespaces[0] = string(u.Username)
	for i := 1; i < nsCount; i++ {
		namespaces[i] = string(teams[i-1].Name)
	}

	return namespaces, nil
}

func (h RepositoryHandler) getRepositoryFromRequest(ctx context.Context, namespace, name string) (repository.Repository, error) {
	u, ok := AuthenticatedUserFromContext(ctx)
	if !ok {
		h.logger.ErrorContext(ctx, "could not parse user from request context")
		return repository.Repository{}, newErrorResponse(http.StatusInternalServerError, "internal server error")
	}

	authorizedNamespaces, err := h.getUserNamespaces(ctx, u)
	if err != nil {
		h.logger.ErrorContext(ctx, "failed to get namespaces for user", slog.Any("error", err))
		return repository.Repository{}, newErrorResponse(http.StatusInternalServerError, "internal server error")
	}

	if !slices.Contains(authorizedNamespaces, namespace) {
		return repository.Repository{}, newErrorResponse(http.StatusForbidden, "not authorized for given namespace")
	}

	repo, err := h.repoStore.GetByNamespaceAndName(ctx, namespace, name)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return repository.Repository{}, newErrorResponse(http.StatusNotFound, "repository not found")
		}

		h.logger.ErrorContext(ctx, "failed to get repository",
			slog.Any("error", err),
			slog.String("namespace", namespace),
			slog.String("name", name))
		return repository.Repository{}, newErrorResponse(http.StatusInternalServerError, "internal server error")
	}

	return repo, nil
}

func convertToRepositoryResponse(r repository.Repository) oas.RepositoryResponse {
	return oas.RepositoryResponse{
		ID:         r.ID,
		Namespace:  r.Namespace,
		Name:       string(r.Name),
		Visibility: oas.RepositoryResponseVisibility(r.Visibility),
		CreatedAt:  r.CreatedAt,
	}
}
