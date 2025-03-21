package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/evanebb/regauth/server/middleware"
	"github.com/evanebb/regauth/server/response"
	"github.com/evanebb/regauth/user"
	"github.com/evanebb/regauth/util"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"log/slog"
	"net/http"
)

type teamCtxKey struct{}
type teamMemberCtxKey struct{}

func TeamParser(l *slog.Logger, s user.TeamStore) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			u, ok := middleware.AuthenticatedUserFromContext(r.Context())
			if !ok {
				l.ErrorContext(r.Context(), "could not parse user from request context")
				response.WriteJSONError(w, http.StatusUnauthorized, "authentication failed")
				return
			}

			name := chi.URLParam(r, "name")
			if name == "" {
				response.WriteJSONError(w, http.StatusBadRequest, "no team name given")
				return
			}

			team, err := s.GetByName(r.Context(), name)
			if err != nil {
				if errors.Is(err, user.ErrTeamNotFound) {
					response.WriteJSONError(w, http.StatusNotFound, "team not found")
					return
				}

				l.ErrorContext(r.Context(), "could not get team", slog.Any("error", err))
				response.WriteJSONError(w, http.StatusInternalServerError, "internal server error")
				return
			}

			teamMember, err := s.GetTeamMember(r.Context(), team.ID, u.ID)
			if err != nil {
				if errors.Is(err, user.ErrTeamMemberNotFound) {
					response.WriteJSONError(w, http.StatusNotFound, "team not found")
					return
				}

				l.ErrorContext(r.Context(), "could not get team member", slog.Any("error", err))
				response.WriteJSONError(w, http.StatusInternalServerError, "internal server error")
				return
			}

			ctx := withTeam(r.Context(), team)
			ctx = withTeamMember(ctx, teamMember)
			next.ServeHTTP(w, r.WithContext(ctx))
		}
		return http.HandlerFunc(fn)
	}
}

// withTeam sets the given team.Team in the context.
// Use teamFromContext to retrieve the team.
func withTeam(ctx context.Context, t user.Team) context.Context {
	return context.WithValue(ctx, teamCtxKey{}, t)
}

// teamFromContext parses the current team.Team from the given request context.
// This requires the team to have been previously set in the context, for example by the TeamParser middleware.
func teamFromContext(ctx context.Context) (user.Team, bool) {
	val, ok := ctx.Value(teamCtxKey{}).(user.Team)
	return val, ok
}

// withTeamMember sets team membership information for the current user for the current team in the context.
// Use teamMemberFromContext to retrieve it.
func withTeamMember(ctx context.Context, tm user.TeamMember) context.Context {
	return context.WithValue(ctx, teamMemberCtxKey{}, tm)
}

// teamMemberFromContext parses team membership information for the current user from the given request context.
// This requires this to have been previously set in the context.
func teamMemberFromContext(ctx context.Context) (user.TeamMember, bool) {
	val, ok := ctx.Value(teamMemberCtxKey{}).(user.TeamMember)
	return val, ok
}

func CreateTeam(l *slog.Logger, s user.TeamStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, ok := middleware.AuthenticatedUserFromContext(r.Context())
		if !ok {
			l.ErrorContext(r.Context(), "could not parse user from request context")
			response.WriteJSONError(w, http.StatusUnauthorized, "authenticated failed")
			return
		}

		var team user.Team
		if err := json.NewDecoder(r.Body).Decode(&team); err != nil {
			response.WriteJSONError(w, http.StatusBadRequest, "invalid JSON body given")
			return
		}

		_, err := s.GetByName(r.Context(), team.Name)
		if err == nil {
			response.WriteJSONError(w, http.StatusBadRequest, "team already exists")
			return
		}

		if !errors.Is(err, user.ErrTeamNotFound) {
			l.ErrorContext(r.Context(), "could not get team", slog.Any("error", err))
			response.WriteJSONError(w, http.StatusInternalServerError, "internal server error")
			return
		}

		team.ID = uuid.New()

		member := user.TeamMember{
			UserID:   u.ID,
			TeamID:   team.ID,
			Username: u.Username,
			Role:     user.TeamMemberRoleAdmin,
		}
		if err := member.IsValid(); err != nil {
			// shouldn't ever happen, just check it anyway
			response.WriteJSONError(w, http.StatusBadRequest, err.Error())
		}

		err = s.Tx(r.Context(), func(ctx context.Context) error {
			if err := s.Create(ctx, team); err != nil {
				return err
			}

			// add the current user as an admin to the team
			if err := s.AddTeamMember(ctx, member); err != nil {
				return err
			}

			return nil
		})

		if err != nil {
			l.Error("could not create team", "error", err)
			response.WriteJSONError(w, http.StatusInternalServerError, "internal server error")
			return
		}

		response.WriteJSONResponse(w, http.StatusOK, team)
	}
}

func ListTeams(l *slog.Logger, s user.TeamStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, ok := middleware.AuthenticatedUserFromContext(r.Context())
		if !ok {
			l.ErrorContext(r.Context(), "could not parse user from request context")
			response.WriteJSONError(w, http.StatusUnauthorized, "authenticated failed")
			return
		}

		teams, err := s.GetAllByUser(r.Context(), u.ID)
		if err != nil {
			l.ErrorContext(r.Context(), "could not get teams for user", slog.Any("error", err))
			response.WriteJSONError(w, http.StatusInternalServerError, "internal server error")
			return
		}

		response.WriteJSONResponse(w, http.StatusOK, util.NilSliceToEmpty(teams))
	}
}

func GetTeam(l *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		team, ok := teamFromContext(r.Context())
		if !ok {
			l.ErrorContext(r.Context(), "could not parse team from request context")
			response.WriteJSONError(w, http.StatusInternalServerError, "internal server error")
			return
		}

		response.WriteJSONResponse(w, http.StatusOK, team)
	}
}

func DeleteTeam(l *slog.Logger, s user.TeamStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		team, ok := teamFromContext(r.Context())
		if !ok {
			l.ErrorContext(r.Context(), "could not parse team from request context")
			response.WriteJSONError(w, http.StatusInternalServerError, "internal server error")
			return
		}

		if err := s.DeleteByID(r.Context(), team.ID); err != nil {
			l.ErrorContext(r.Context(), "could not delete team", slog.Any("error", err))
			response.WriteJSONError(w, http.StatusInternalServerError, "internal server error")
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

type addTeamMemberRequest struct {
	Username string `json:"username"`
	Role     string `json:"role"`
}

func AddTeamMember(l *slog.Logger, teamStore user.TeamStore, userStore user.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		currentMember, ok := teamMemberFromContext(r.Context())
		if !ok {
			l.ErrorContext(r.Context(), "could not parse team member from request context")
			response.WriteJSONError(w, http.StatusInternalServerError, "internal server error")
			return
		}

		if currentMember.Role != user.TeamMemberRoleAdmin {
			response.WriteJSONError(w, http.StatusForbidden, "insufficient permission")
			return
		}

		var req addTeamMemberRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			response.WriteJSONError(w, http.StatusBadRequest, "invalid JSON body given")
			return
		}

		userToAdd, err := userStore.GetByUsername(r.Context(), req.Username)
		if err != nil {
			response.WriteJSONError(w, http.StatusNotFound, "user not found")
			return
		}

		member := user.TeamMember{
			UserID:   userToAdd.ID,
			TeamID:   currentMember.TeamID,
			Username: userToAdd.Username,
			Role:     user.TeamMemberRole(req.Role),
		}
		if err := member.IsValid(); err != nil {
			response.WriteJSONError(w, http.StatusBadRequest, err.Error())
		}

		if err := teamStore.AddTeamMember(r.Context(), member); err != nil {
			l.ErrorContext(r.Context(), "could not add team member", slog.Any("error", err))
			response.WriteJSONError(w, http.StatusInternalServerError, "internal server error")
		}

		response.WriteJSONResponse(w, http.StatusOK, member)
	}
}

func ListTeamMembers(l *slog.Logger, s user.TeamStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		team, ok := teamFromContext(r.Context())
		if !ok {
			l.ErrorContext(r.Context(), "could not parse team from request context")
			response.WriteJSONError(w, http.StatusInternalServerError, "internal server error")
			return
		}

		members, err := s.GetTeamMembers(r.Context(), team.ID)
		if err != nil {
			l.ErrorContext(r.Context(), "could not get team members", slog.Any("error", err))
			response.WriteJSONError(w, http.StatusInternalServerError, "internal server error")
			return
		}

		response.WriteJSONResponse(w, http.StatusOK, util.NilSliceToEmpty(members))
	}
}

func RemoveTeamMember(l *slog.Logger, teamStore user.TeamStore, userStore user.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		currentMember, ok := teamMemberFromContext(r.Context())
		if !ok {
			l.ErrorContext(r.Context(), "could not parse team member from request context")
			response.WriteJSONError(w, http.StatusInternalServerError, "internal server error")
			return
		}

		if currentMember.Role != user.TeamMemberRoleAdmin {
			response.WriteJSONError(w, http.StatusForbidden, "insufficient permission")
			return
		}

		username := chi.URLParam(r, "username")
		if username == "" {
			response.WriteJSONError(w, http.StatusBadRequest, "no username given")
			return
		}

		userToRemove, err := userStore.GetByUsername(r.Context(), username)
		if err != nil {
			if errors.Is(err, user.ErrNotFound) {
				response.WriteJSONError(w, http.StatusNotFound, "user not found")
				return
			}

			l.ErrorContext(r.Context(), "could not get user", slog.Any("error", err))
			response.WriteJSONError(w, http.StatusInternalServerError, "internal server error")
			return
		}

		if currentMember.UserID == userToRemove.ID {
			response.WriteJSONError(w, http.StatusBadRequest, "cannot remove current user from team")
			return
		}

		if err := teamStore.RemoveTeamMember(r.Context(), currentMember.TeamID, userToRemove.ID); err != nil {
			l.ErrorContext(r.Context(), "could not remove team member", slog.Any("error", err))
			response.WriteJSONError(w, http.StatusInternalServerError, "internal server error")
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
