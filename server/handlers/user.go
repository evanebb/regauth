package handlers

import (
	"context"
	"errors"
	"github.com/evanebb/regauth/auth/local"
	"github.com/evanebb/regauth/oas"
	"github.com/evanebb/regauth/user"
	"github.com/google/uuid"
	"log/slog"
	"net/http"
	"time"
)

type UserHandler struct {
	logger           *slog.Logger
	userStore        user.Store
	credentialsStore local.UserCredentialsStore
}

func (h UserHandler) CreateUser(ctx context.Context, req *oas.UserRequest) (*oas.UserResponse, error) {
	if err := h.requireRole(ctx, user.RoleAdmin); err != nil {
		return nil, err
	}

	id, err := uuid.NewV7()
	if err != nil {
		h.logger.ErrorContext(ctx, "could not generate UUID", "error", err)
		return nil, newInternalServerErrorResponse()
	}

	newUser := user.User{
		ID:        id,
		Username:  user.Username(req.Username),
		Role:      user.Role(req.Role),
		CreatedAt: time.Now(),
	}

	if err := newUser.IsValid(); err != nil {
		return nil, newErrorResponse(http.StatusBadRequest, err.Error())
	}

	_, err = h.userStore.GetByUsername(ctx, req.Username)
	if err == nil {
		return nil, newErrorResponse(http.StatusBadRequest, "user already exists")
	}

	if !errors.Is(err, user.ErrNotFound) {
		h.logger.ErrorContext(ctx, "could not get user", slog.Any("error", err))
		return nil, newInternalServerErrorResponse()
	}

	if err := h.userStore.Create(ctx, newUser); err != nil {
		h.logger.ErrorContext(ctx, "could not create user", slog.Any("error", err))
		return nil, newInternalServerErrorResponse()
	}

	resp := convertToUserResponse(newUser)
	return &resp, nil
}

func (h UserHandler) ListUsers(ctx context.Context) ([]oas.UserResponse, error) {
	if err := h.requireRole(ctx, user.RoleAdmin); err != nil {
		return nil, err
	}

	users, err := h.userStore.GetAll(ctx)
	if err != nil {
		h.logger.ErrorContext(ctx, "could not get users", slog.Any("error", err))
		return nil, newInternalServerErrorResponse()
	}

	return convertSlice(users, convertToUserResponse), nil
}

func (h UserHandler) GetUser(ctx context.Context, params oas.GetUserParams) (*oas.UserResponse, error) {
	if err := h.requireRole(ctx, user.RoleAdmin); err != nil {
		return nil, err
	}

	u, err := h.getUserFromRequest(ctx, params.Username)
	if err != nil {
		return nil, err
	}

	resp := convertToUserResponse(u)
	return &resp, nil
}

func (h UserHandler) DeleteUser(ctx context.Context, params oas.DeleteUserParams) error {
	if err := h.requireRole(ctx, user.RoleAdmin); err != nil {
		return err
	}

	currentUser, ok := AuthenticatedUserFromContext(ctx)
	if !ok {
		h.logger.ErrorContext(ctx, "could not parse user from request context")
		return newInternalServerErrorResponse()
	}

	u, err := h.getUserFromRequest(ctx, params.Username)
	if err != nil {
		return err
	}

	if currentUser.ID == u.ID {
		return newErrorResponse(http.StatusBadRequest, "cannot delete current user")
	}

	if err := h.userStore.DeleteByID(ctx, u.ID); err != nil {
		h.logger.ErrorContext(ctx, "could not delete user", slog.Any("error", err))
		return newInternalServerErrorResponse()
	}

	return nil
}

func (h UserHandler) ChangeUserPassword(ctx context.Context, req *oas.UserPasswordChangeRequest, params oas.ChangeUserPasswordParams) error {
	if err := h.requireRole(ctx, user.RoleAdmin); err != nil {
		return err
	}

	u, err := h.getUserFromRequest(ctx, params.Username)
	if err != nil {
		return err
	}

	credentials := local.UserCredentials{UserID: u.ID}
	if err := credentials.SetPassword(req.Password); err != nil {
		if errors.Is(err, local.ErrWeakPassword) {
			return newErrorResponse(http.StatusBadRequest, err.Error())
		}

		h.logger.ErrorContext(ctx, "could not set password", slog.Any("error", err))
		return newInternalServerErrorResponse()
	}

	if err := h.credentialsStore.Save(ctx, credentials); err != nil {
		h.logger.ErrorContext(ctx, "could not update user credentials", slog.Any("error", err))
		return newInternalServerErrorResponse()
	}

	return nil
}

func (h UserHandler) getUserFromRequest(ctx context.Context, username string) (user.User, error) {
	u, err := h.userStore.GetByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			return user.User{}, newErrorResponse(http.StatusNotFound, "user not found")
		}

		h.logger.ErrorContext(ctx, "could not get user", slog.Any("error", err))
		return user.User{}, newInternalServerErrorResponse()
	}

	return u, nil
}

func (h UserHandler) requireRole(ctx context.Context, role user.Role) error {
	u, ok := AuthenticatedUserFromContext(ctx)
	if !ok {
		h.logger.ErrorContext(ctx, "could not parse user from request context")
		return newInternalServerErrorResponse()
	}

	if u.Role != role {
		return newErrorResponse(http.StatusForbidden, "insufficient permission")
	}

	return nil
}

func convertToUserResponse(u user.User) oas.UserResponse {
	return oas.UserResponse{
		ID:        u.ID,
		Username:  string(u.Username),
		Role:      oas.UserResponseRole(u.Role),
		CreatedAt: u.CreatedAt,
	}
}
