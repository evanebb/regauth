package postgres

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"reflect"
	"testing"
)

func TestTransactionStore_QuerierFromContext(t *testing.T) {
	db := getDatabaseConnection(t)
	s := TransactionStore{db: db}

	t.Run("no transaction", func(t *testing.T) {

		querier := s.QuerierFromContext(context.Background())
		if _, ok := querier.(*pgxpool.Pool); !ok {
			t.Errorf("expected querier to be *pgxpool.Pool, got %q", reflect.TypeOf(querier).String())
		}
	})

	t.Run("with transaction", func(t *testing.T) {
		err := s.Tx(context.Background(), func(ctx context.Context) error {
			querier := s.QuerierFromContext(ctx)
			if _, ok := querier.(pgx.Tx); !ok {
				// return error here and fail test outside of transaction, to ensure transaction is closed and such
				return fmt.Errorf("expected querier to be pgx.Tx, got %q", reflect.TypeOf(querier).String())
			}

			return nil
		})
		if err != nil {
			t.Error(err)
		}
	})
}

func TestTransactionStore_Tx(t *testing.T) {
	db := getDatabaseConnection(t)
	s := TransactionStore{db: db}

	query := "CREATE TABLE test_table (id bigint PRIMARY KEY GENERATED ALWAYS AS IDENTITY, name VARCHAR(255))"
	if _, err := s.db.Exec(t.Context(), query); err != nil {
		t.Errorf("failed to create test table: %q", err)
	}

	var expectedErr = errors.New("error returned from transaction")

	t.Run("ensure transaction is rolled back on error", func(t *testing.T) {
		err := s.Tx(t.Context(), func(ctx context.Context) error {
			insertQuery := "INSERT INTO test_table (name) VALUES ('foo')"
			if _, err := s.QuerierFromContext(ctx).Exec(ctx, insertQuery); err != nil {
				return err
			}

			// return an error so the transaction is rolled back
			return expectedErr
		})
		if !errors.Is(err, expectedErr) {
			t.Errorf("expected transaction to return %q, got %q", expectedErr, err)
		}

		// check that the transaction was rolled back
		var count int
		if err := s.db.QueryRow(t.Context(), "SELECT count(name) FROM test_table").Scan(&count); err != nil {
			t.Errorf("could not query count, got error %q", err)
		}

		if count != 0 {
			t.Errorf("expected no rows to be present, got %d rows instead", count)
		}
	})
}
