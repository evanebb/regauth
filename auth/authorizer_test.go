package auth

import (
	"github.com/evanebb/regauth/repository"
	"github.com/evanebb/regauth/store/memory"
	"github.com/evanebb/regauth/token"
	"github.com/evanebb/regauth/user"
	"github.com/google/uuid"
	"io"
	"log/slog"
	"slices"
	"strings"
	"testing"
)

func TestAuthorizer_AuthorizeAccess(t *testing.T) {
	t.Parallel()

	// authorizer wants a logger, and since we don't care about the logs currently, so just discard them
	logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))

	t.Run("no access requested", func(t *testing.T) {
		t.Parallel()
		repoStore := memory.NewRepositoryStore()
		teamStore := memory.NewTeamStore()
		a := NewAuthorizer(logger, repoStore, teamStore)

		requestedAccess := Access{}
		grantedAccess, err := a.AuthorizeAccess(t.Context(), nil, nil, requestedAccess)
		if err != nil {
			t.Fatalf("expected err to be nil, got %q", err)
		}

		if !compareAccess(grantedAccess, requestedAccess) {
			t.Fatalf("expected %+v, got %+v", requestedAccess, grantedAccess)
		}
	})

	t.Run("no access granted for non-repository types", func(t *testing.T) {
		t.Parallel()
		repoStore := memory.NewRepositoryStore()
		teamStore := memory.NewTeamStore()
		a := NewAuthorizer(logger, repoStore, teamStore)

		requestedAccess := Access{
			{
				Type: "foo",
			},
		}
		grantedAccess, err := a.AuthorizeAccess(t.Context(), nil, nil, requestedAccess)
		if err != nil {
			t.Fatalf("expected err to be nil, got %q", err)
		}

		expectedAccess := Access{}
		if !compareAccess(grantedAccess, expectedAccess) {
			t.Fatalf("expected %+v, got %+v", expectedAccess, grantedAccess)
		}
	})

	t.Run("no access granted for invalid repository names", func(t *testing.T) {
		t.Parallel()
		repoStore := memory.NewRepositoryStore()
		teamStore := memory.NewTeamStore()
		a := NewAuthorizer(logger, repoStore, teamStore)

		requestedAccess := Access{
			{
				Type: "repository",
				Name: "invalid",
			},
		}
		grantedAccess, err := a.AuthorizeAccess(t.Context(), nil, nil, requestedAccess)
		if err != nil {
			t.Fatalf("expected err to be nil, got %q", err)
		}

		expectedAccess := Access{}
		if !compareAccess(grantedAccess, expectedAccess) {
			t.Fatalf("expected %+v, got %+v", expectedAccess, grantedAccess)
		}
	})

	t.Run("no access granted for unknown repositories", func(t *testing.T) {
		t.Parallel()
		repoStore := memory.NewRepositoryStore()
		teamStore := memory.NewTeamStore()
		a := NewAuthorizer(logger, repoStore, teamStore)

		requestedAccess := Access{
			{
				Type: "repository",
				Name: "foo/bar",
			},
		}
		grantedAccess, err := a.AuthorizeAccess(t.Context(), nil, nil, requestedAccess)
		if err != nil {
			t.Fatalf("expected err to be nil, got %q", err)
		}

		expectedAccess := Access{}
		if !compareAccess(grantedAccess, expectedAccess) {
			t.Fatalf("expected %+v, got %+v", expectedAccess, grantedAccess)
		}
	})

	t.Run("grant pull access for public repositories", func(t *testing.T) {
		t.Parallel()
		repoStore := memory.NewRepositoryStore()
		teamStore := memory.NewTeamStore()
		a := NewAuthorizer(logger, repoStore, teamStore)

		repo := repository.Repository{
			ID:         uuid.New(),
			Namespace:  "user",
			Name:       "public-repo",
			Visibility: repository.VisibilityPublic,
		}
		if err := repoStore.Create(t.Context(), repo); err != nil {
			t.Fatalf("could not create repository: %q", err)
		}

		requestedAccess := Access{
			{
				Type:    "repository",
				Name:    "user/public-repo",
				Actions: []string{"pull", "push"},
			},
		}
		grantedAccess, err := a.AuthorizeAccess(t.Context(), nil, nil, requestedAccess)
		if err != nil {
			t.Fatalf("expected err to be nil, got %q", err)
		}

		expectedAccess := Access{
			{
				Type:    "repository",
				Name:    "user/public-repo",
				Actions: []string{"pull"},
			},
		}
		if !compareAccess(grantedAccess, expectedAccess) {
			t.Fatalf("expected %+v, got %+v", expectedAccess, grantedAccess)
		}
	})

	t.Run("remove access from result if no actions requested", func(t *testing.T) {
		t.Parallel()
		repoStore := memory.NewRepositoryStore()
		teamStore := memory.NewTeamStore()
		a := NewAuthorizer(logger, repoStore, teamStore)

		repo := repository.Repository{
			ID:         uuid.New(),
			Namespace:  "user",
			Name:       "public-repo",
			Visibility: repository.VisibilityPublic,
		}
		if err := repoStore.Create(t.Context(), repo); err != nil {
			t.Fatalf("could not create repository: %q", err)
		}

		requestedAccess := Access{
			{
				Type: "repository",
				Name: "user/public-repo",
			},
		}
		grantedAccess, err := a.AuthorizeAccess(t.Context(), nil, nil, requestedAccess)
		if err != nil {
			t.Fatalf("expected err to be nil, got %q", err)
		}

		expectedAccess := Access{}
		if !compareAccess(grantedAccess, expectedAccess) {
			t.Fatalf("expected %+v, got %+v", expectedAccess, grantedAccess)
		}
	})

	t.Run("full access granted for owned repositories", func(t *testing.T) {
		t.Parallel()
		repoStore := memory.NewRepositoryStore()
		teamStore := memory.NewTeamStore()
		a := NewAuthorizer(logger, repoStore, teamStore)

		repo1 := repository.Repository{
			ID:         uuid.New(),
			Namespace:  "user",
			Name:       "myrepo",
			Visibility: repository.VisibilityPrivate,
		}
		repo2 := repository.Repository{
			ID:         uuid.New(),
			Namespace:  "myteam",
			Name:       "anotherone",
			Visibility: repository.VisibilityPrivate,
		}
		if err := repoStore.Create(t.Context(), repo1); err != nil {
			t.Fatalf("could not create repository: %q", err)
		}
		if err := repoStore.Create(t.Context(), repo2); err != nil {
			t.Fatalf("could not create repository: %q", err)
		}

		teamID := uuid.New()
		team := user.Team{
			ID:   teamID,
			Name: "myteam",
		}
		if err := teamStore.Create(t.Context(), team); err != nil {
			t.Fatalf("could not create team: %q", err)
		}

		userID := uuid.New()
		teamMember := user.TeamMember{
			UserID: userID,
			TeamID: teamID,
		}
		if err := teamStore.AddTeamMember(t.Context(), teamMember); err != nil {
			t.Fatalf("could not add team member: %q", err)
		}

		u := &user.User{
			ID:       userID,
			Username: "user",
		}
		tok := &token.PersonalAccessToken{Permission: token.PermissionReadWriteDelete}

		requestedAccess := Access{
			{
				Type:    "repository",
				Name:    "user/myrepo",
				Actions: []string{"pull", "push", "delete"},
			},
			{
				Type:    "repository",
				Name:    "myteam/anotherone",
				Actions: []string{"pull", "push"},
			},
		}
		grantedAccess, err := a.AuthorizeAccess(t.Context(), u, tok, requestedAccess)
		if err != nil {
			t.Fatalf("expected err to be nil, got %q", err)
		}

		if !compareAccess(grantedAccess, requestedAccess) {
			t.Fatalf("expected %+v, got %+v", requestedAccess, grantedAccess)
		}
	})

	tokenTestCases := []struct {
		desc            string
		permission      token.Permission
		expectedActions []string
	}{
		{
			"read-only",
			token.PermissionReadOnly,
			[]string{"pull"},
		},
		{
			"read-write",
			token.PermissionReadWrite,
			[]string{"pull", "push"},
		},
		{
			"read-write-delete",
			token.PermissionReadWriteDelete,
			[]string{"pull", "push", "delete"},
		},
	}

	for _, c := range tokenTestCases {
		t.Run("restrict permissions for "+c.desc+" token for owned repository", func(t *testing.T) {
			t.Parallel()
			repoStore := memory.NewRepositoryStore()
			teamStore := memory.NewTeamStore()
			a := NewAuthorizer(logger, repoStore, teamStore)

			repo := repository.Repository{
				ID:         uuid.New(),
				Namespace:  "user",
				Name:       "myrepo",
				Visibility: repository.VisibilityPrivate,
			}
			if err := repoStore.Create(t.Context(), repo); err != nil {
				t.Fatalf("could not create repository: %q", err)
			}

			u := &user.User{
				ID:       uuid.New(),
				Username: "user",
			}
			tok := &token.PersonalAccessToken{Permission: c.permission}

			requestedAccess := Access{
				{
					Type:    "repository",
					Name:    "user/myrepo",
					Actions: []string{"pull", "push", "delete"},
				},
			}
			grantedAccess, err := a.AuthorizeAccess(t.Context(), u, tok, requestedAccess)
			if err != nil {
				t.Fatalf("expected err to be nil, got %q", err)
			}

			expectedAccess := Access{
				{
					Type:    "repository",
					Name:    "user/myrepo",
					Actions: c.expectedActions,
				},
			}
			if !compareAccess(grantedAccess, expectedAccess) {
				t.Fatalf("expected %+v, got %+v", expectedAccess, grantedAccess)
			}
		})
	}

}

// compareAccess is a bit of garbage code to check two Access instances for equality.
func compareAccess(a1 Access, a2 Access) bool {
	// first, stuff the second access instance into a map for easy lookups
	// we store an int value in the map, so multiple entries with the same values are counted properly
	a2Map := make(map[string]int)
	for _, e := range a2 {
		s := actionsToString(e)
		if _, ok := a2Map[s]; !ok {
			a2Map[s] = 0
		}

		a2Map[s]++
	}

	for _, entry := range a1 {
		s := actionsToString(entry)
		if _, ok := a2Map[s]; !ok {
			// if the current set of resource actions is not present in the map, they are not the same
			return false
		}

		a2Map[s]--
		if a2Map[s] <= 0 {
			delete(a2Map, s)
		}
	}

	if len(a2Map) > 0 {
		// check if there are any entries left in a2 that are not in a1, if so, they are not the same
		return false
	}

	return true
}

func actionsToString(ra ResourceActions) string {
	tmp := make([]string, len(ra.Actions))
	copy(tmp, ra.Actions)
	slices.Sort(tmp)
	return ra.Type + ra.Name + strings.Join(tmp, "")
}
