package handlers

import (
	"context"
	"errors"
	"github.com/evanebb/regauth/oas"
	"github.com/evanebb/regauth/user"
	"github.com/google/uuid"
	"log/slog"
	"net/http"
	"time"
)

type TeamHandler struct {
	logger    *slog.Logger
	teamStore user.TeamStore
	userStore user.Store
}

func (h TeamHandler) CreateTeam(ctx context.Context, req *oas.TeamRequest) (*oas.TeamResponse, error) {
	u, ok := AuthenticatedUserFromContext(ctx)
	if !ok {
		h.logger.ErrorContext(ctx, "could not parse user from request context")
		return nil, newErrorResponse(http.StatusInternalServerError, "internal server error")
	}

	_, err := h.teamStore.GetByName(ctx, req.Name)
	if err == nil {
		return nil, newErrorResponse(http.StatusBadRequest, "team already exists")
	}

	if !errors.Is(err, user.ErrTeamNotFound) {
		h.logger.ErrorContext(ctx, "could not get team", slog.Any("error", err))
		return nil, newErrorResponse(http.StatusInternalServerError, "internal server error")
	}

	id, err := uuid.NewV7()
	if err != nil {
		h.logger.ErrorContext(ctx, "could not generate UUID", "error", err)
		return nil, newErrorResponse(http.StatusInternalServerError, "internal server error")
	}

	team := user.Team{
		ID:        id,
		Name:      user.TeamName(req.Name),
		CreatedAt: time.Now(),
	}

	if err := team.IsValid(); err != nil {
		return nil, newErrorResponse(http.StatusBadRequest, err.Error())
	}

	member := user.TeamMember{
		UserID:    u.ID,
		TeamID:    team.ID,
		Username:  u.Username,
		Role:      user.TeamMemberRoleAdmin,
		CreatedAt: time.Now(),
	}

	if err := member.IsValid(); err != nil {
		return nil, newErrorResponse(http.StatusBadRequest, err.Error())
	}

	err = h.teamStore.Tx(ctx, func(txCtx context.Context) error {
		if err := h.teamStore.Create(txCtx, team); err != nil {
			return err
		}

		// add the current user as an admin to the team
		if err := h.teamStore.AddTeamMember(txCtx, member); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		h.logger.ErrorContext(ctx, "could not create team", slog.Any("error", err))
		return nil, newErrorResponse(http.StatusInternalServerError, "internal server error")
	}

	resp := convertToTeamResponse(team)
	return &resp, nil
}

func (h TeamHandler) ListTeams(ctx context.Context) ([]oas.TeamResponse, error) {
	u, ok := AuthenticatedUserFromContext(ctx)
	if !ok {
		h.logger.ErrorContext(ctx, "could not parse user from request context")
		return nil, newErrorResponse(http.StatusInternalServerError, "internal server error")
	}

	teams, err := h.teamStore.GetAllByUser(ctx, u.ID)
	if err != nil {
		h.logger.ErrorContext(ctx, "could not get teams for user", slog.Any("error", err))
		return nil, newErrorResponse(http.StatusInternalServerError, "internal server error")
	}

	return convertSlice(teams, convertToTeamResponse), nil
}

func (h TeamHandler) GetTeam(ctx context.Context, params oas.GetTeamParams) (*oas.TeamResponse, error) {
	team, _, err := h.getTeamAndCurrentMemberFromRequest(ctx, params.Name)
	if err != nil {
		return nil, err
	}

	resp := convertToTeamResponse(team)
	return &resp, nil
}

func (h TeamHandler) DeleteTeam(ctx context.Context, params oas.DeleteTeamParams) error {
	team, member, err := h.getTeamAndCurrentMemberFromRequest(ctx, params.Name)
	if err != nil {
		return err
	}

	if member.Role != user.TeamMemberRoleAdmin {
		return newErrorResponse(http.StatusForbidden, "insufficient permission")
	}

	if err := h.teamStore.DeleteByID(ctx, team.ID); err != nil {
		h.logger.ErrorContext(ctx, "could not delete team", slog.Any("error", err))
		return newInternalServerErrorResponse()
	}

	return nil
}

func (h TeamHandler) AddTeamMember(ctx context.Context, req *oas.TeamMemberRequest, params oas.AddTeamMemberParams) (*oas.TeamMemberResponse, error) {
	team, currentMember, err := h.getTeamAndCurrentMemberFromRequest(ctx, params.Name)
	if err != nil {
		return nil, err
	}

	if currentMember.Role != user.TeamMemberRoleAdmin {
		return nil, newErrorResponse(http.StatusForbidden, "insufficient permission")
	}

	userToAdd, err := h.userStore.GetByUsername(ctx, req.Username)
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			return nil, newErrorResponse(http.StatusNotFound, "user not found")
		}

		h.logger.ErrorContext(ctx, "could not get user", slog.Any("error", err))
		return nil, newInternalServerErrorResponse()
	}

	newMember := user.TeamMember{
		UserID:    userToAdd.ID,
		TeamID:    team.ID,
		Username:  userToAdd.Username,
		Role:      user.TeamMemberRole(req.Role),
		CreatedAt: time.Now(),
	}

	if err := newMember.IsValid(); err != nil {
		return nil, newErrorResponse(http.StatusBadRequest, err.Error())
	}

	if err := h.teamStore.AddTeamMember(ctx, newMember); err != nil {
		h.logger.ErrorContext(ctx, "could not add team member", slog.Any("error", err))
		return nil, newInternalServerErrorResponse()
	}

	resp := convertToTeamMemberResponse(newMember)
	return &resp, nil
}

func (h TeamHandler) ListTeamMembers(ctx context.Context, params oas.ListTeamMembersParams) ([]oas.TeamMemberResponse, error) {
	team, _, err := h.getTeamAndCurrentMemberFromRequest(ctx, params.Name)
	if err != nil {
		return nil, err
	}

	members, err := h.teamStore.GetTeamMembers(ctx, team.ID)
	if err != nil {
		h.logger.ErrorContext(ctx, "could not get team members", slog.Any("error", err))
		return nil, newInternalServerErrorResponse()
	}

	return convertSlice(members, convertToTeamMemberResponse), nil
}

func (h TeamHandler) RemoveTeamMember(ctx context.Context, params oas.RemoveTeamMemberParams) error {
	team, currentMember, err := h.getTeamAndCurrentMemberFromRequest(ctx, params.Name)
	if err != nil {
		return err
	}

	if currentMember.Role != user.TeamMemberRoleAdmin {
		return newErrorResponse(http.StatusForbidden, "insufficient permission")
	}

	userToRemove, err := h.userStore.GetByUsername(ctx, params.Username)
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			return newErrorResponse(http.StatusNotFound, "user not found")
		}

		h.logger.ErrorContext(ctx, "could not get user", slog.Any("error", err))
		return newInternalServerErrorResponse()
	}

	if currentMember.UserID == userToRemove.ID {
		return newErrorResponse(http.StatusBadRequest, "cannot remove current user from team")
	}

	if err := h.teamStore.RemoveTeamMember(ctx, team.ID, userToRemove.ID); err != nil {
		h.logger.ErrorContext(ctx, "could not remove team member", slog.Any("error", err))
		return newInternalServerErrorResponse()
	}

	return nil
}

func (h TeamHandler) getTeamAndCurrentMemberFromRequest(ctx context.Context, name string) (user.Team, user.TeamMember, error) {
	u, ok := AuthenticatedUserFromContext(ctx)
	if !ok {
		h.logger.ErrorContext(ctx, "could not parse user from request context")
		return user.Team{}, user.TeamMember{}, newErrorResponse(http.StatusInternalServerError, "internal server error")
	}

	team, err := h.teamStore.GetByName(ctx, name)
	if err != nil {
		if errors.Is(err, user.ErrTeamNotFound) {
			return user.Team{}, user.TeamMember{}, newErrorResponse(http.StatusNotFound, "team not found")
		}

		h.logger.ErrorContext(ctx, "could not get team",
			slog.Any("error", err),
			slog.String("name", name))
		return user.Team{}, user.TeamMember{}, newErrorResponse(http.StatusInternalServerError, "internal server error")
	}

	member, err := h.teamStore.GetTeamMember(ctx, team.ID, u.ID)
	if err != nil {
		if errors.Is(err, user.ErrTeamMemberNotFound) {
			return user.Team{}, user.TeamMember{}, newErrorResponse(http.StatusNotFound, "team not found")
		}

		h.logger.ErrorContext(ctx, "could not get team member",
			slog.Any("error", err),
			slog.String("team", name),
			slog.String("user", string(u.Username)))
		return user.Team{}, user.TeamMember{}, newErrorResponse(http.StatusInternalServerError, "internal server error")
	}

	return team, member, nil
}

func convertToTeamResponse(t user.Team) oas.TeamResponse {
	return oas.TeamResponse{
		ID:        t.ID,
		Name:      string(t.Name),
		CreatedAt: t.CreatedAt,
	}
}

func convertToTeamMemberResponse(tm user.TeamMember) oas.TeamMemberResponse {
	return oas.TeamMemberResponse{
		UserId:    tm.UserID,
		Username:  string(tm.Username),
		Role:      oas.TeamMemberResponseRole(tm.Role),
		CreatedAt: tm.CreatedAt,
	}
}
