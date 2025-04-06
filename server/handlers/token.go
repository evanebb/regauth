package handlers

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"github.com/evanebb/regauth/oas"
	"github.com/evanebb/regauth/token"
	"github.com/google/uuid"
	"log/slog"
	"net/http"
	"time"
)

type TokenHandler struct {
	logger     *slog.Logger
	tokenStore token.Store
}

func (h TokenHandler) CreatePersonalAccessToken(ctx context.Context, req *oas.PersonalAccessTokenRequest) (*oas.PersonalAccessTokenCreationResponse, error) {
	u, ok := AuthenticatedUserFromContext(ctx)
	if !ok {
		h.logger.ErrorContext(ctx, "could not parse user from request context")
		return nil, newInternalServerErrorResponse()
	}

	id, err := uuid.NewV7()
	if err != nil {
		h.logger.ErrorContext(ctx, "could not generate UUID", "error", err)
		return nil, newInternalServerErrorResponse()
	}

	pat := token.PersonalAccessToken{
		ID:             id,
		Description:    token.Description(req.Description),
		Permission:     token.Permission(req.Permission),
		ExpirationDate: req.ExpirationDate,
		UserID:         u.ID,
		CreatedAt:      time.Now(),
	}

	if err := pat.IsValid(); err != nil {
		return nil, newErrorResponse(http.StatusBadRequest, err.Error())
	}

	randBytes := make([]byte, 42)
	_, _ = rand.Read(randBytes)
	plainTextToken := "registry_pat_" + base64.URLEncoding.EncodeToString(randBytes)

	if err := h.tokenStore.Create(ctx, pat, plainTextToken); err != nil {
		h.logger.ErrorContext(ctx, "could not create personal access token", slog.Any("error", err))
		return nil, newInternalServerErrorResponse()
	}

	return &oas.PersonalAccessTokenCreationResponse{
		ID:             pat.ID,
		Description:    string(pat.Description),
		Permission:     oas.PersonalAccessTokenCreationResponsePermission(pat.Permission),
		ExpirationDate: pat.ExpirationDate,
		Token:          plainTextToken,
		CreatedAt:      time.Now(),
	}, nil
}

func (h TokenHandler) ListPersonalAccessTokens(ctx context.Context) ([]oas.PersonalAccessTokenResponse, error) {
	u, ok := AuthenticatedUserFromContext(ctx)
	if !ok {
		h.logger.ErrorContext(ctx, "could not parse user from request context")
		return nil, newInternalServerErrorResponse()
	}

	tokens, err := h.tokenStore.GetAllByUser(ctx, u.ID)
	if err != nil {
		h.logger.ErrorContext(ctx, "could not get personal access tokens for user", slog.Any("error", err))
		return nil, newInternalServerErrorResponse()
	}

	return convertSlice(tokens, convertToPersonalAccessTokenResponse), nil
}

func (h TokenHandler) GetPersonalAccessToken(ctx context.Context, params oas.GetPersonalAccessTokenParams) (*oas.PersonalAccessTokenResponse, error) {
	pat, err := h.getPersonalAccessTokenFromRequest(ctx, params.ID)
	if err != nil {
		return nil, err
	}

	resp := convertToPersonalAccessTokenResponse(pat)
	return &resp, nil
}

func (h TokenHandler) DeletePersonalAccessToken(ctx context.Context, params oas.DeletePersonalAccessTokenParams) error {
	pat, err := h.getPersonalAccessTokenFromRequest(ctx, params.ID)
	if err != nil {
		return err
	}

	if err := h.tokenStore.DeleteByID(ctx, pat.ID); err != nil {
		h.logger.ErrorContext(ctx, "could not delete personal access token", slog.Any("error", err))
		return newInternalServerErrorResponse()
	}

	return nil
}

func (h TokenHandler) getPersonalAccessTokenFromRequest(ctx context.Context, id uuid.UUID) (token.PersonalAccessToken, error) {
	u, ok := AuthenticatedUserFromContext(ctx)
	if !ok {
		h.logger.ErrorContext(ctx, "could not parse user from request context")
		return token.PersonalAccessToken{}, newInternalServerErrorResponse()
	}

	pat, err := h.tokenStore.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, token.ErrNotFound) {
			return token.PersonalAccessToken{}, newErrorResponse(http.StatusNotFound, "personal access token not found")
		}

		h.logger.ErrorContext(ctx, "could not get personal access token", slog.Any("error", err))
		return token.PersonalAccessToken{}, newInternalServerErrorResponse()
	}

	if pat.UserID != u.ID {
		return token.PersonalAccessToken{}, newErrorResponse(http.StatusNotFound, "personal access token not found")
	}

	return pat, nil
}

func convertToPersonalAccessTokenResponse(t token.PersonalAccessToken) oas.PersonalAccessTokenResponse {
	return oas.PersonalAccessTokenResponse{
		ID:             t.ID,
		Description:    string(t.Description),
		Permission:     oas.PersonalAccessTokenResponsePermission(t.Permission),
		ExpirationDate: t.ExpirationDate,
		CreatedAt:      t.CreatedAt,
	}
}
