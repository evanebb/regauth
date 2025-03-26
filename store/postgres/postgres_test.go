package postgres

import (
	"context"
	"github.com/evanebb/regauth/resources/database"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"log"
	"os"
	"testing"
	"time"
)

var container *postgres.PostgresContainer

func TestMain(m *testing.M) {
	os.Exit(testMain(m))
}

func testMain(m *testing.M) int {
	ctx := context.Background()
	var err error

	container, err = postgres.Run(ctx,
		"postgres:17",
		postgres.WithDatabase("regauth"),
		postgres.WithUsername("regauth"),
		postgres.WithPassword("Welkom01!"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second)),
	)
	defer func() {
		if err := testcontainers.TerminateContainer(container); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Fatalf("failed to start container: %s", err)
	}

	db, err := pgxpool.New(ctx, container.MustConnectionString(ctx))
	if err != nil {
		log.Fatalf("failed to open database connection: %s", err)
	}

	goose.SetBaseFS(database.Files)
	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatal(err)
	}

	stdlibDb := stdlib.OpenDBFromPool(db)
	if err := goose.Up(stdlibDb, "migrations"); err != nil {
		log.Fatalf("failed to run migrations: %s", err)
	}

	if err := goose.Up(stdlibDb, "seeds", goose.WithNoVersioning()); err != nil {
		log.Fatalf("failed to run seeds: %s", err)
	}

	// close the database connection before making a snapshot
	db.Close()
	_ = stdlibDb.Close()

	if err := container.Snapshot(ctx); err != nil {
		log.Fatalf("failed to make snapshot: %s", err)
	}

	return m.Run()
}

// getDatabaseConnection will open a database connection to the Postgres testcontainer.
// It will also register a cleanup function to automatically close the database connection and restore the database
// to the last snapshot.
func getDatabaseConnection(t *testing.T) *pgxpool.Pool {
	db, err := pgxpool.New(t.Context(), container.MustConnectionString(t.Context()))
	if err != nil {
		t.Errorf("failed to open database connection: %s", err)
	}

	t.Cleanup(func() {
		db.Close()

		if err := container.Restore(context.Background()); err != nil {
			log.Fatalf("failed to restore snapshot: %s", err)
		}
	})

	return db
}
