package postgres

import (
	"context"
	"errors"
	"github.com/evanebb/regauth/pat"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PersonalAccessTokenStore struct {
	db *pgxpool.Pool
}

func NewPersonalAccessTokenStore(db *pgxpool.Pool) PersonalAccessTokenStore {
	return PersonalAccessTokenStore{db: db}
}

func (s PersonalAccessTokenStore) GetAllForUser(ctx context.Context, userID uuid.UUID) ([]pat.PersonalAccessToken, error) {
	var tokens []pat.PersonalAccessToken

	query := "SELECT uuid, hash, description, permission, expiration_date, user_uuid FROM personal_access_tokens WHERE user_uuid = $1"
	rows, err := s.db.Query(ctx, query, userID)
	if err != nil {
		return tokens, err
	}

	for rows.Next() {
		var t pat.PersonalAccessToken
		var pt string

		err = rows.Scan(&t.ID, &t.Hash, &t.Description, &pt, &t.ExpirationDate, &t.UserID)
		if err != nil {
			return tokens, err
		}

		t.Permission = permissionFromDatabaseMap[pt]

		err = t.IsValid()
		if err != nil {
			return tokens, err
		}

		tokens = append(tokens, t)
	}

	return tokens, nil
}

func (s PersonalAccessTokenStore) GetByID(ctx context.Context, id uuid.UUID) (pat.PersonalAccessToken, error) {
	var t pat.PersonalAccessToken
	var pt string

	query := "SELECT uuid, hash, description, permission, expiration_date, user_uuid FROM personal_access_tokens WHERE uuid = $1"
	err := s.db.QueryRow(ctx, query, id).Scan(&t.ID, &t.Hash, &t.Description, &pt, &t.ExpirationDate, &t.UserID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return t, pat.ErrNotFound
		}

		return t, err
	}

	t.Permission = permissionFromDatabaseMap[pt]
	return t, t.IsValid()
}

func (s PersonalAccessTokenStore) Create(ctx context.Context, t pat.PersonalAccessToken) error {
	_, err := s.GetByID(ctx, t.ID)
	if err == nil {
		return pat.ErrAlreadyExists
	}
	if !errors.Is(err, pat.ErrNotFound) {
		return err
	}

	query := "INSERT INTO personal_access_tokens (uuid, hash, description, permission, expiration_date, user_uuid) VALUES ($1, $2, $3, $4, $5, $6)"
	_, err = s.db.Exec(ctx, query, t.ID, t.Hash, t.Description, permissionToDatabaseMap[t.Permission], t.ExpirationDate, t.UserID)
	return err
}

func (s PersonalAccessTokenStore) DeleteByID(ctx context.Context, id uuid.UUID) error {
	query := "DELETE FROM personal_access_tokens WHERE uuid = $1"
	_, err := s.db.Exec(ctx, query, id)
	return err
}

func (s PersonalAccessTokenStore) GetUsageLog(ctx context.Context, tokenID uuid.UUID) ([]pat.UsageLogEntry, error) {
	var log []pat.UsageLogEntry

	query := "SELECT token_uuid, source_ip, timestamp FROM personal_access_tokens_usage_log WHERE token_uuid = $1 ORDER BY timestamp DESC"
	rows, err := s.db.Query(ctx, query, tokenID)
	if err != nil {
		return log, err
	}

	for rows.Next() {
		var l pat.UsageLogEntry

		err = rows.Scan(&l.TokenID, &l.SourceIP, &l.Timestamp)
		if err != nil {
			return log, err
		}

		log = append(log, l)
	}

	return log, nil
}

func (s PersonalAccessTokenStore) AddUsageLogEntry(ctx context.Context, e pat.UsageLogEntry) error {
	query := "INSERT INTO personal_access_tokens_usage_log (token_uuid, source_ip, timestamp) VALUES ($1, $2, $3)"
	_, err := s.db.Exec(ctx, query, e.TokenID, e.SourceIP, e.Timestamp)
	return err
}

var permissionFromDatabaseMap = map[string]pat.Permission{
	"read_only":         pat.PermissionReadOnly,
	"read_write":        pat.PermissionReadWrite,
	"read_write_delete": pat.PermissionReadWriteDelete,
}

var permissionToDatabaseMap = map[pat.Permission]string{
	pat.PermissionReadOnly:        "read_only",
	pat.PermissionReadWrite:       "read_write",
	pat.PermissionReadWriteDelete: "read_write_delete",
}
