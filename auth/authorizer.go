package auth

import (
	"context"
	"errors"
	"github.com/evanebb/regauth/repository"
	"github.com/evanebb/regauth/token"
	"github.com/evanebb/regauth/user"
	"log/slog"
	"slices"
	"strings"
)

type Authorizer interface {
	AuthorizeAccess(ctx context.Context, u *user.User, p *token.PersonalAccessToken, requestedAccess Access) (Access, error)
}

func NewAuthorizer(logger *slog.Logger, repoStore repository.Store, teamStore user.TeamStore) Authorizer {
	return authorizer{logger: logger, repoStore: repoStore, teamStore: teamStore}
}

type ResourceActions struct {
	Type    string   `json:"type"`
	Name    string   `json:"name"`
	Actions []string `json:"actions"`
}

type Access []ResourceActions

type authorizer struct {
	logger    *slog.Logger
	repoStore repository.Store
	teamStore user.TeamStore
}

func (a authorizer) AuthorizeAccess(ctx context.Context, u *user.User, p *token.PersonalAccessToken, requestedAccess Access) (Access, error) {
	var authorizedNamespaces []string
	if u != nil {
		authorizedNamespaces = append(authorizedNamespaces, u.Username.String())

		teams, err := a.teamStore.GetAllByUser(ctx, u.ID)
		if err != nil {
			return Access{}, err
		}

		for _, team := range teams {
			authorizedNamespaces = append(authorizedNamespaces, team.Name)
		}
	}

	grantedAccess := Access{}
	for _, requestedActions := range requestedAccess {
		grantedActions, err := a.authorizeResourceActions(ctx, authorizedNamespaces, p, requestedActions)
		if err != nil {
			if errors.Is(err, ErrAccessNotGranted) {
				continue
			}
			return grantedAccess, err
		}
		grantedAccess = append(grantedAccess, grantedActions)
	}

	return grantedAccess, nil
}

func (a authorizer) authorizeResourceActions(
	ctx context.Context,
	authorizedNamespaces []string,
	p *token.PersonalAccessToken,
	r ResourceActions,
) (ResourceActions, error) {
	var granted ResourceActions
	if r.Type != "repository" {
		// Only authorize access to repositories
		a.logger.Debug("cannot grant access to non-repository requests", "type", r.Type, "name", r.Name)
		return granted, ErrAccessNotGranted
	}

	split := strings.Split(r.Name, "/")
	if len(split) != 2 {
		// Only support repositories like 'namespace/name'
		a.logger.Debug("malformed repository name given", "repository", r.Name)
		return granted, ErrAccessNotGranted
	}

	repo, err := a.repoStore.GetByNamespaceAndName(ctx, split[0], split[1])
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			// If the repository cannot be found, grant no access
			a.logger.Debug("repository not found, no access granted", "repository", r.Name)
			return granted, errors.Join(ErrAccessNotGranted, err)
		}
		return granted, err
	}

	var allowedActions []string
	// First, determine the actions that are allowed for the user
	if slices.Contains(authorizedNamespaces, repo.Namespace) {
		// the repository is in an authorized namespace, allow all actions
		a.logger.Debug("repository is in authorized namespace, all actions allowed", "repository", r.Name)
		allowedActions = []string{"pull", "push", "delete"}
	} else if repo.Visibility == repository.VisibilityPublic {
		// If the user does not own this repository but it is public, pull access is allowed
		a.logger.Debug("user does not own public repository, allowing pull access", "repository", r.Name, "repository", r.Name)
		allowedActions = []string{"pull"}
	}

	// Remove actions that are not allowed by the assigned token permissions or not requested by the user
	for _, allowedAction := range allowedActions {
		if !slices.Contains(r.Actions, allowedAction) {
			a.logger.Debug("action not requested", "action", allowedAction)
			continue
		}

		if p != nil {
			if !slices.Contains(p.Permission.GetAllowedActions(), allowedAction) {
				a.logger.Debug("action not allowed by personal access token permission", "action", allowedAction, "repository", r.Name)
				continue
			}
		}

		a.logger.Debug("action granted", "action", allowedAction, "repository", r.Name)
		granted.Actions = append(granted.Actions, allowedAction)
	}

	if len(granted.Actions) == 0 {
		a.logger.Debug("no actions granted, removing access from result entirely", "repository", r.Name)
		return granted, ErrAccessNotGranted
	}

	granted.Name = r.Name
	granted.Type = r.Type
	return granted, nil
}
