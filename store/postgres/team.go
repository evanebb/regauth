package postgres

import (
	"context"
	"errors"
	"github.com/evanebb/regauth/user"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TeamStore struct {
	TransactionStore
}

func NewTeamStore(db *pgxpool.Pool) TeamStore {
	return TeamStore{TransactionStore{db: db}}
}

func (s TeamStore) GetAllByUser(ctx context.Context, userID uuid.UUID) ([]user.Team, error) {
	var teams []user.Team

	query := `
		SELECT
			teams.id,
			teams.name,
			teams.created_at
		FROM teams
		JOIN team_members ON teams.id = team_members.team_id
		JOIN users ON team_members.user_id = users.id
		WHERE users.id = $1
		`
	rows, err := s.QuerierFromContext(ctx).Query(ctx, query, userID)
	if err != nil {
		return teams, err
	}

	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (user.Team, error) {
		var t user.Team

		if err := rows.Scan(&t.ID, &t.Name, &t.CreatedAt); err != nil {
			return t, err
		}

		return t, t.IsValid()
	})
}

func (s TeamStore) GetByID(ctx context.Context, id uuid.UUID) (user.Team, error) {
	var t user.Team

	query := "SELECT id, name, created_at FROM teams WHERE id = $1"
	err := s.QuerierFromContext(ctx).QueryRow(ctx, query, id).Scan(&t.ID, &t.Name, &t.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return t, user.ErrTeamNotFound
		}

		return t, err
	}

	return t, nil
}

func (s TeamStore) GetByName(ctx context.Context, name string) (user.Team, error) {
	var t user.Team

	query := "SELECT id, name, created_at FROM teams WHERE name = $1"
	err := s.QuerierFromContext(ctx).QueryRow(ctx, query, name).Scan(&t.ID, &t.Name, &t.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return t, user.ErrTeamNotFound
		}

		return t, err
	}

	return t, nil
}

func (s TeamStore) Create(ctx context.Context, t user.Team) error {
	_, err := s.GetByID(ctx, t.ID)
	if err == nil {
		return user.ErrTeamAlreadyExists
	}
	if !errors.Is(err, user.ErrTeamNotFound) {
		return err
	}

	tx, err := s.QuerierFromContext(ctx).Begin(ctx)
	if err != nil {
		return err
	}

	query := "INSERT INTO teams (id, name) VALUES ($1, $2)"
	if _, err := tx.Exec(ctx, query, t.ID, t.Name); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}

	nsQuery := "INSERT INTO namespaces (name, team_id) VALUES ($1, $2)"
	if _, err := tx.Exec(ctx, nsQuery, t.Name, t.ID); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}

	_ = tx.Commit(ctx)
	return nil
}

func (s TeamStore) DeleteByID(ctx context.Context, id uuid.UUID) error {
	query := "DELETE FROM teams WHERE id = $1"
	_, err := s.QuerierFromContext(ctx).Exec(ctx, query, id)
	return err
}

func (s TeamStore) GetTeamMember(ctx context.Context, teamID uuid.UUID, userID uuid.UUID) (user.TeamMember, error) {
	var tm user.TeamMember

	query := `
		SELECT
			team_members.user_id,
			team_members.team_id,
			users.username,
			team_members.role,
			team_members.created_at
		FROM team_members
		JOIN users ON team_members.user_id = users.id
		WHERE team_id = $1
		AND users.id = $2
		`
	err := s.QuerierFromContext(ctx).QueryRow(ctx, query, teamID, userID).Scan(&tm.UserID, &tm.TeamID, &tm.Username, &tm.Role, &tm.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return tm, user.ErrTeamMemberNotFound
		}

		return tm, err
	}

	return tm, nil
}

func (s TeamStore) GetTeamMembers(ctx context.Context, teamID uuid.UUID) ([]user.TeamMember, error) {
	var members []user.TeamMember

	query := `
		SELECT
			team_members.user_id,
			team_members.team_id,
			users.username,
			team_members.role,
			team_members.created_at
		FROM team_members
		JOIN teams ON team_members.team_id = teams.id
		JOIN users ON team_members.user_id = users.id
		WHERE teams.id = $1
		`
	rows, err := s.QuerierFromContext(ctx).Query(ctx, query, teamID)
	if err != nil {
		return members, err
	}

	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (user.TeamMember, error) {
		var tm user.TeamMember

		if err := rows.Scan(&tm.UserID, &tm.TeamID, &tm.Username, &tm.Role, &tm.CreatedAt); err != nil {
			return tm, err
		}

		return tm, tm.IsValid()
	})
}

func (s TeamStore) AddTeamMember(ctx context.Context, m user.TeamMember) error {
	_, err := s.GetTeamMember(ctx, m.TeamID, m.UserID)
	if err == nil {
		return user.ErrTeamMemberAlreadyExists
	}
	if !errors.Is(err, user.ErrTeamMemberNotFound) {
		return err
	}

	query := "INSERT INTO team_members (user_id, team_id, role) VALUES ($1, $2, $3)"
	_, err = s.QuerierFromContext(ctx).Exec(ctx, query, m.UserID, m.TeamID, m.Role)
	return err
}

func (s TeamStore) RemoveTeamMember(ctx context.Context, teamID uuid.UUID, userID uuid.UUID) error {
	query := "DELETE FROM team_members WHERE team_id = $1 AND user_id = $2"
	_, err := s.QuerierFromContext(ctx).Exec(ctx, query, teamID, userID)
	return err
}
