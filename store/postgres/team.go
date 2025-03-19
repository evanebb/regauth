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
			teams.uuid,
			teams.name
		FROM teams
		JOIN team_members ON teams.uuid = team_members.team_uuid
		JOIN users ON team_members.user_uuid = users.uuid
		WHERE users.uuid = $1
		`
	rows, err := s.QuerierFromContext(ctx).Query(ctx, query, userID)
	defer rows.Close()
	if err != nil {
		return teams, err
	}

	for rows.Next() {
		var t user.Team

		if err := rows.Scan(&t.ID, &t.Name); err != nil {
			return teams, err
		}

		teams = append(teams, t)
	}

	return teams, nil
}

func (s TeamStore) GetByID(ctx context.Context, id uuid.UUID) (user.Team, error) {
	var t user.Team

	query := "SELECT uuid, name FROM teams WHERE uuid = $1"
	err := s.QuerierFromContext(ctx).QueryRow(ctx, query, id).Scan(&t.ID, &t.Name)
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

	query := "SELECT uuid, name FROM teams WHERE name = $1"
	err := s.QuerierFromContext(ctx).QueryRow(ctx, query, name).Scan(&t.ID, &t.Name)
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

	query := "INSERT INTO teams (uuid, name) VALUES ($1, $2)"
	if _, err := tx.Exec(ctx, query, t.ID, t.Name); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}

	nsQuery := "INSERT INTO namespaces (uuid, name, team_uuid) VALUES ($1, $2, $3)"
	if _, err := tx.Exec(ctx, nsQuery, uuid.New(), t.Name, t.ID); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}

	_ = tx.Commit(ctx)
	return nil
}

func (s TeamStore) DeleteByID(ctx context.Context, id uuid.UUID) error {
	query := "DELETE FROM teams WHERE uuid = $1"
	_, err := s.QuerierFromContext(ctx).Exec(ctx, query, id)
	return err
}

func (s TeamStore) GetTeamMember(ctx context.Context, teamID uuid.UUID, userID uuid.UUID) (user.TeamMember, error) {
	var tm user.TeamMember

	query := `
		SELECT
			team_members.user_uuid,
			team_members.team_uuid,
			users.username,
			team_members.role
		FROM team_members
		JOIN users ON team_members.user_uuid = users.uuid
		WHERE users.uuid = $1
		`
	err := s.QuerierFromContext(ctx).QueryRow(ctx, query, userID).Scan(&tm.UserID, &tm.TeamID, &tm.Username, &tm.Role)
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
			team_members.user_uuid,
			team_members.team_uuid,
			users.username,
			team_members.role
		FROM team_members
		JOIN teams ON team_members.team_uuid = teams.uuid
		JOIN users ON team_members.user_uuid = users.uuid
		WHERE teams.uuid = $1
		`
	rows, err := s.QuerierFromContext(ctx).Query(ctx, query, teamID)
	defer rows.Close()
	if err != nil {
		return members, err
	}

	for rows.Next() {
		var tm user.TeamMember

		if err := rows.Scan(&tm.UserID, &tm.TeamID, &tm.Username, &tm.Role); err != nil {
			return members, err
		}

		members = append(members, tm)
	}

	return members, nil
}

func (s TeamStore) AddTeamMember(ctx context.Context, m user.TeamMember) error {
	query := "INSERT INTO team_members (user_uuid, team_uuid, role) VALUES ($1, $2, $3)"
	_, err := s.QuerierFromContext(ctx).Exec(ctx, query, m.UserID, m.TeamID, m.Role)
	return err
}

func (s TeamStore) RemoveTeamMember(ctx context.Context, teamID uuid.UUID, userID uuid.UUID) error {
	query := "DELETE FROM team_members WHERE team_uuid = $1 AND user_uuid = $2"
	_, err := s.QuerierFromContext(ctx).Exec(ctx, query, teamID, userID)
	return err
}
