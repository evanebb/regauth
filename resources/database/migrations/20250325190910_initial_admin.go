package migrations

import (
	"context"
	"database/sql"
	"errors"
	"github.com/evanebb/regauth/configuration"
	"github.com/evanebb/regauth/user"
	"github.com/google/uuid"
	"github.com/pressly/goose/v3"
	"golang.org/x/crypto/bcrypt"
)

// RegisterInitialAdminMigration is a hacky way to create the initial admin user from the values in the app config by
// hooking into Goose.
func RegisterInitialAdminMigration(conf *configuration.Configuration) {
	goose.AddMigrationContext(upInitialAdmin(conf), downInitialAdmin(conf))
}

func upInitialAdmin(conf *configuration.Configuration) func(ctx context.Context, tx *sql.Tx) error {
	username, password := conf.InitialAdmin.Username, conf.InitialAdmin.Password

	return func(ctx context.Context, tx *sql.Tx) error {
		if username == "" || password == "" {
			return errors.New("no initial admin user specified")
		}

		id, err := uuid.NewV7()
		if err != nil {
			return err
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}

		query := "INSERT INTO users (id, username, role, password_hash) VALUES ($1, $2, $3, $4)"
		if _, err := tx.ExecContext(ctx, query, id, username, user.RoleAdmin, hash); err != nil {
			return err
		}

		nsQuery := "iNSERT INTO namespaces (name, user_id) VALUES ($1, $2)"
		if _, err := tx.ExecContext(ctx, nsQuery, username, id); err != nil {
			return err
		}

		return nil
	}
}

func downInitialAdmin(conf *configuration.Configuration) func(ctx context.Context, tx *sql.Tx) error {
	username, password := conf.InitialAdmin.Username, conf.InitialAdmin.Password

	return func(ctx context.Context, tx *sql.Tx) error {
		if username == "" || password == "" {
			return nil
		}

		// FIXME: this can go completely wrong if the initial admin username has changed between the up and down
		// migration, so maybe we shouldn't do this at all
		query := "DELETE FROM users WHERE username = $1"
		_, err := tx.ExecContext(ctx, query, username)
		return err
	}
}
