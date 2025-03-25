package postgres

import (
	"context"
	"errors"
	"github.com/evanebb/regauth/token"
	"github.com/evanebb/regauth/util/auth"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PersonalAccessTokenStore struct {
	TransactionStore
}

func NewPersonalAccessTokenStore(db *pgxpool.Pool) PersonalAccessTokenStore {
	return PersonalAccessTokenStore{TransactionStore{db: db}}
}

func (s PersonalAccessTokenStore) GetAllByUser(ctx context.Context, userID uuid.UUID) ([]token.PersonalAccessToken, error) {
	var tokens []token.PersonalAccessToken

	query := "SELECT id, description, permission, expiration_date, user_id FROM personal_access_tokens WHERE user_id = $1"
	rows, err := s.QuerierFromContext(ctx).Query(ctx, query, userID)
	defer rows.Close()
	if err != nil {
		return tokens, err
	}

	for rows.Next() {
		var t token.PersonalAccessToken
		var pt string

		err = rows.Scan(&t.ID, &t.Description, &pt, &t.ExpirationDate, &t.UserID)
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

func (s PersonalAccessTokenStore) GetByID(ctx context.Context, id uuid.UUID) (token.PersonalAccessToken, error) {
	var t token.PersonalAccessToken
	var pt string

	query := "SELECT id, description, permission, expiration_date, user_id FROM personal_access_tokens WHERE id = $1"
	err := s.QuerierFromContext(ctx).QueryRow(ctx, query, id).Scan(&t.ID, &t.Description, &pt, &t.ExpirationDate, &t.UserID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return t, token.ErrNotFound
		}

		return t, err
	}

	t.Permission = permissionFromDatabaseMap[pt]
	return t, t.IsValid()
}

func (s PersonalAccessTokenStore) GetByPlainTextToken(ctx context.Context, plainTextToken string) (token.PersonalAccessToken, error) {
	// Select all tokens of which the stored last eight characters match the plain-text token
	lastEight := plainTextToken[len(plainTextToken)-8:]
	query := "SELECT id, hash, description, permission, expiration_date, user_id FROM personal_access_tokens WHERE last_eight = $1"
	rows, err := s.QuerierFromContext(ctx).Query(ctx, query, lastEight)
	defer rows.Close()
	if err != nil {
		return token.PersonalAccessToken{}, err
	}

	for rows.Next() {
		var t token.PersonalAccessToken
		var pt string
		var hash []byte

		err = rows.Scan(&t.ID, &hash, &t.Description, &pt, &t.ExpirationDate, &t.UserID)
		if err != nil {
			continue
		}

		t.Permission = permissionFromDatabaseMap[pt]

		err = t.IsValid()
		if err != nil {
			continue
		}

		if err := auth.CompareTokenAndHash(plainTextToken, hash); err == nil {
			return t, nil
		}
	}

	return token.PersonalAccessToken{}, token.ErrNotFound
}

func (s PersonalAccessTokenStore) Create(ctx context.Context, t token.PersonalAccessToken, plainTextToken string) error {
	_, err := s.GetByID(ctx, t.ID)
	if err == nil {
		return token.ErrAlreadyExists
	}
	if !errors.Is(err, token.ErrNotFound) {
		return err
	}

	lastEight := plainTextToken[len(plainTextToken)-8:]
	hash := auth.HashTokenWithRandomSalt(plainTextToken)

	query := "INSERT INTO personal_access_tokens (id, hash, last_eight ,description, permission, expiration_date, user_id) VALUES ($1, $2, $3, $4, $5, $6, $7)"
	_, err = s.QuerierFromContext(ctx).Exec(ctx, query, t.ID, hash, lastEight, t.Description, permissionToDatabaseMap[t.Permission], t.ExpirationDate, t.UserID)
	return err
}

func (s PersonalAccessTokenStore) DeleteByID(ctx context.Context, id uuid.UUID) error {
	query := "DELETE FROM personal_access_tokens WHERE id = $1"
	_, err := s.QuerierFromContext(ctx).Exec(ctx, query, id)
	return err
}

func (s PersonalAccessTokenStore) GetUsageLog(ctx context.Context, tokenID uuid.UUID) ([]token.UsageLogEntry, error) {
	var log []token.UsageLogEntry

	query := "SELECT token_id, source_ip, timestamp FROM personal_access_tokens_usage_log WHERE token_id = $1 ORDER BY timestamp DESC"
	rows, err := s.QuerierFromContext(ctx).Query(ctx, query, tokenID)
	defer rows.Close()
	if err != nil {
		return log, err
	}

	for rows.Next() {
		var l token.UsageLogEntry

		err = rows.Scan(&l.TokenID, &l.SourceIP, &l.Timestamp)
		if err != nil {
			return log, err
		}

		log = append(log, l)
	}

	return log, nil
}

func (s PersonalAccessTokenStore) AddUsageLogEntry(ctx context.Context, e token.UsageLogEntry) error {
	query := "INSERT INTO personal_access_tokens_usage_log (token_id, source_ip, timestamp) VALUES ($1, $2, $3)"
	_, err := s.QuerierFromContext(ctx).Exec(ctx, query, e.TokenID, e.SourceIP, e.Timestamp)
	return err
}

var permissionFromDatabaseMap = map[string]token.Permission{
	"read_only":         token.PermissionReadOnly,
	"read_write":        token.PermissionReadWrite,
	"read_write_delete": token.PermissionReadWriteDelete,
}

var permissionToDatabaseMap = map[token.Permission]string{
	token.PermissionReadOnly:        "read_only",
	token.PermissionReadWrite:       "read_write",
	token.PermissionReadWriteDelete: "read_write_delete",
}
