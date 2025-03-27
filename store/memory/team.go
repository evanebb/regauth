package memory

import (
	"context"
	"fmt"
	"github.com/evanebb/regauth/user"
	"github.com/google/uuid"
	"sync"
)

type TeamStore struct {
	TransactionStore
	mu          sync.RWMutex
	teams       map[uuid.UUID]user.Team
	teamMembers map[uuid.UUID]map[uuid.UUID]user.TeamMember
}

func NewTeamStore() *TeamStore {
	return &TeamStore{
		teams:       make(map[uuid.UUID]user.Team),
		teamMembers: make(map[uuid.UUID]map[uuid.UUID]user.TeamMember),
	}
}

func (s *TeamStore) GetAllByUser(ctx context.Context, userID uuid.UUID) ([]user.Team, error) {
	var teams []user.Team

	s.mu.RLock()
	defer s.mu.RUnlock()

	for teamID, members := range s.teamMembers {
		for memberUserID, _ := range members {
			if memberUserID != userID {
				continue
			}

			team, ok := s.teams[teamID]
			if !ok {
				return teams, fmt.Errorf("user is member of team %s, but team does not exist", teamID)
			}

			if err := team.IsValid(); err != nil {
				return teams, err
			}

			teams = append(teams, team)
		}
	}

	return teams, nil
}

func (s *TeamStore) GetByID(ctx context.Context, id uuid.UUID) (user.Team, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	team, ok := s.teams[id]
	if !ok {
		return user.Team{}, user.ErrTeamNotFound
	}

	return team, team.IsValid()
}

func (s *TeamStore) GetByName(ctx context.Context, name string) (user.Team, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, team := range s.teams {
		if string(team.Name) == name {
			return team, team.IsValid()
		}
	}

	return user.Team{}, user.ErrTeamNotFound
}

func (s *TeamStore) Create(ctx context.Context, t user.Team) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.teams[t.ID]; ok {
		return user.ErrTeamAlreadyExists
	}

	s.teams[t.ID] = t

	return nil
}

func (s *TeamStore) DeleteByID(ctx context.Context, id uuid.UUID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.teams, id)

	return nil
}

func (s *TeamStore) GetTeamMember(ctx context.Context, teamID uuid.UUID, userID uuid.UUID) (user.TeamMember, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	member, ok := s.teamMembers[teamID][userID]
	if !ok {
		return user.TeamMember{}, user.ErrTeamMemberNotFound
	}

	return member, member.IsValid()
}

func (s *TeamStore) GetTeamMembers(ctx context.Context, teamID uuid.UUID) ([]user.TeamMember, error) {
	var members []user.TeamMember

	s.mu.RLock()
	defer s.mu.RUnlock()

	if _, ok := s.teams[teamID]; !ok {
		return members, user.ErrTeamNotFound
	}

	memberMap, ok := s.teamMembers[teamID]
	if !ok {
		return members, user.ErrTeamNotFound
	}

	members = make([]user.TeamMember, 0, len(memberMap))
	for _, m := range memberMap {
		if err := m.IsValid(); err != nil {
			return members, err
		}

		members = append(members, m)
	}

	return members, nil
}

func (s *TeamStore) AddTeamMember(ctx context.Context, m user.TeamMember) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.teams[m.TeamID]; !ok {
		return user.ErrTeamNotFound
	}

	if _, ok := s.teamMembers[m.TeamID]; !ok {
		s.teamMembers[m.TeamID] = make(map[uuid.UUID]user.TeamMember)
	}

	s.teamMembers[m.TeamID][m.UserID] = m

	return nil
}

func (s *TeamStore) RemoveTeamMember(ctx context.Context, teamID uuid.UUID, userID uuid.UUID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.teams[teamID]; !ok {
		return user.ErrTeamNotFound
	}

	if _, ok := s.teamMembers[teamID]; !ok {
		// team has no team members yet, just return nil
		return nil
	}

	delete(s.teamMembers[teamID], userID)

	return nil
}
