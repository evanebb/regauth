package postgres

import (
	"errors"
	"github.com/evanebb/regauth/user"
	"github.com/google/uuid"
	"testing"
)

func TestTeamStore_GetAllByUser(t *testing.T) {
	db := getDatabaseConnection(t)
	s := NewTeamStore(db)

	userID, _ := uuid.Parse("0195cd11-2863-71d4-a3c4-032bc264cf81")

	teams, err := s.GetAllByUser(t.Context(), userID)
	if err != nil {
		t.Errorf("expected err to be nil, got %q", err)
	}

	if len(teams) != 2 {
		t.Errorf("expected two teams, got %d", len(teams))
	}
}

func TestTeamStore_GetByID(t *testing.T) {
	db := getDatabaseConnection(t)
	s := NewTeamStore(db)

	t.Run("existing team", func(t *testing.T) {
		teamID, _ := uuid.Parse("0195d46e-cfbf-7324-b9aa-4c9c78d3b722")

		expectedTeam := user.Team{
			ID:   teamID,
			Name: "team-1",
		}

		team, err := s.GetByID(t.Context(), teamID)
		if err != nil {
			t.Errorf("expected err to be nil, got %q", err)
		}

		if team != expectedTeam {
			t.Errorf("expected %+v, got %+v", expectedTeam, team)
		}
	})

	t.Run("team does not exist", func(t *testing.T) {
		if _, err := s.GetByID(t.Context(), uuid.Nil); !errors.Is(err, user.ErrTeamNotFound) {
			t.Errorf("expected %q, got %q", user.ErrTeamNotFound, err)
		}
	})
}

func TestTeamStore_GetByName(t *testing.T) {
	db := getDatabaseConnection(t)
	s := NewTeamStore(db)

	t.Run("existing team", func(t *testing.T) {
		teamID, _ := uuid.Parse("0195d46e-cfbf-7324-b9aa-4c9c78d3b722")

		expectedTeam := user.Team{
			ID:   teamID,
			Name: "team-1",
		}

		team, err := s.GetByName(t.Context(), "team-1")
		if err != nil {
			t.Errorf("expected err to be nil, got %q", err)
		}

		if team != expectedTeam {
			t.Errorf("expected %+v, got %+v", expectedTeam, team)
		}
	})

	t.Run("team does not exist", func(t *testing.T) {
		if _, err := s.GetByName(t.Context(), "foo"); !errors.Is(err, user.ErrTeamNotFound) {
			t.Errorf("expected %q, got %q", user.ErrTeamNotFound, err)
		}
	})
}

func TestTeamStore_Create(t *testing.T) {
	db := getDatabaseConnection(t)
	s := NewTeamStore(db)

	team := user.Team{
		ID:   uuid.New(),
		Name: "new-team",
	}

	t.Run("new team", func(t *testing.T) {
		if err := s.Create(t.Context(), team); err != nil {
			t.Errorf("expected nil, got %q", err)
		}
	})

	t.Run("team already exists", func(t *testing.T) {
		if err := s.Create(t.Context(), team); !errors.Is(err, user.ErrTeamAlreadyExists) {
			t.Errorf("expected %q, got %q", user.ErrTeamAlreadyExists, err)
		}
	})
}

func TestTeamStore_DeleteByID(t *testing.T) {
	db := getDatabaseConnection(t)
	s := NewTeamStore(db)

	teamID, _ := uuid.Parse("0195d46e-cfbf-7324-b9aa-4c9c78d3b722")

	if _, err := s.GetByID(t.Context(), teamID); err != nil {
		t.Errorf("expected team to exist, got %q", err)
	}

	if err := s.DeleteByID(t.Context(), teamID); err != nil {
		t.Errorf("expected nil, got %q", err)
	}

	if _, err := s.GetByID(t.Context(), teamID); !errors.Is(err, user.ErrTeamNotFound) {
		t.Errorf("expected team to be deleted, got %q", err)
	}
}

func TestTeamStore_GetTeamMember(t *testing.T) {
	db := getDatabaseConnection(t)
	s := NewTeamStore(db)

	t.Run("existing team member", func(t *testing.T) {
		teamID, _ := uuid.Parse("0195d46e-cfbf-7324-b9aa-4c9c78d3b722")
		userID, _ := uuid.Parse("0195cd11-2863-71d4-a3c4-032bc264cf81")

		expectedTeamMember := user.TeamMember{
			UserID:   userID,
			TeamID:   teamID,
			Username: "adminuser",
			Role:     user.TeamMemberRoleAdmin,
		}

		member, err := s.GetTeamMember(t.Context(), teamID, userID)
		if err != nil {
			t.Errorf("expected err to be nil, got %q", err)
		}

		if member != expectedTeamMember {
			t.Errorf("expected %+v, got %+v", expectedTeamMember, member)
		}
	})

	t.Run("team member does not exist", func(t *testing.T) {
		if _, err := s.GetTeamMember(t.Context(), uuid.Nil, uuid.Nil); !errors.Is(err, user.ErrTeamMemberNotFound) {
			t.Errorf("expected %q, got %q", user.ErrTeamMemberNotFound, err)
		}
	})
}

func TestTeamStore_GetTeamMembers(t *testing.T) {
	db := getDatabaseConnection(t)
	s := NewTeamStore(db)

	teamID, _ := uuid.Parse("0195d46e-cfbf-7324-b9aa-4c9c78d3b722")

	members, err := s.GetTeamMembers(t.Context(), teamID)
	if err != nil {
		t.Errorf("expected err to be nil, got %q", err)
	}

	if len(members) != 2 {
		t.Errorf("expected two team members, got %d", len(members))
	}
}

func TestTeamStore_AddTeamMember(t *testing.T) {
	db := getDatabaseConnection(t)
	s := NewTeamStore(db)

	teamID, _ := uuid.Parse("0195d46f-fde4-7b27-b542-e41ed0917ace")
	userID, _ := uuid.Parse("0195cd11-2863-721e-a75c-86522539d0ee")

	member := user.TeamMember{
		UserID:   userID,
		TeamID:   teamID,
		Username: "normaluser",
		Role:     user.TeamMemberRoleUser,
	}

	t.Run("new team member", func(t *testing.T) {
		if err := s.AddTeamMember(t.Context(), member); err != nil {
			t.Errorf("expected nil, got %q", err)
		}
	})

	t.Run("team member already exists", func(t *testing.T) {
		if err := s.AddTeamMember(t.Context(), member); !errors.Is(err, user.ErrTeamMemberAlreadyExists) {
			t.Errorf("expected %q, got %q", user.ErrTeamMemberAlreadyExists, err)
		}
	})
}

func TestTeamStore_RemoveTeamMember(t *testing.T) {
	db := getDatabaseConnection(t)
	s := NewTeamStore(db)

	teamID, _ := uuid.Parse("0195d46e-cfbf-7324-b9aa-4c9c78d3b722")
	userID, _ := uuid.Parse("0195cd11-2863-71d4-a3c4-032bc264cf81")

	if _, err := s.GetTeamMember(t.Context(), teamID, userID); err != nil {
		t.Errorf("expected team member to exist, got %q", err)
	}

	if err := s.RemoveTeamMember(t.Context(), teamID, userID); err != nil {
		t.Errorf("expected nil, got %q", err)
	}

	if _, err := s.GetTeamMember(t.Context(), teamID, userID); !errors.Is(err, user.ErrTeamMemberNotFound) {
		t.Errorf("expected team member to be removed, got %q", err)
	}
}
