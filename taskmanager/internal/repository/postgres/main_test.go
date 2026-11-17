package postgres_test

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"

	"github.com/edddddr/taskmanager/internal/domain"
	"github.com/edddddr/taskmanager/internal/repository/postgres"
)

// testPool is nil in -short mode; requireDB then skips the test.
var testPool *pgxpool.Pool

func TestMain(m *testing.M) {
	flag.Parse() // testing.Short() panics if flags aren't parsed yet
	if testing.Short() {
		os.Exit(m.Run())
	}
	// os.Exit skips defers, so the real work lives in a function that RETURNS a code.
	os.Exit(runWithDB(m))
}

func runWithDB(m *testing.M) int {
	ctx := context.Background()

	ctr, err := tcpostgres.Run(ctx, "postgres:16-alpine",
		tcpostgres.WithDatabase("teamtasks_test"),
		tcpostgres.WithUsername("test"),
		tcpostgres.WithPassword("test"),
		tcpostgres.BasicWaitStrategies(),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "start postgres container (is Docker running?): %v\n", err)
		return 1
	}
	defer func() { _ = testcontainers.TerminateContainer(ctr) }()

	dsn, err := ctr.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		fmt.Fprintf(os.Stderr, "connection string: %v\n", err)
		return 1
	}

	testPool, err = pgxpool.New(ctx, dsn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "create pool: %v\n", err)
		return 1
	}
	defer testPool.Close()

	if err := applyMigrations(ctx, testPool, "../../../migrations"); err != nil {
		fmt.Fprintf(os.Stderr, "migrate: %v\n", err)
		return 1
	}
	return m.Run()
}

// applyMigrations runs every *.up.sql in order. With no query arguments, pgx
// uses the simple protocol, which allows several statements per Exec.
func applyMigrations(ctx context.Context, pool *pgxpool.Pool, dir string) error {
	files, err := filepath.Glob(filepath.Join(dir, "*.up.sql"))
	if err != nil {
		return err
	}
	sort.Strings(files) // zero-padded numbers sort correctly
	for _, f := range files {
		sql, err := os.ReadFile(f)
		if err != nil {
			return err
		}
		if _, err := pool.Exec(ctx, string(sql)); err != nil {
			return fmt.Errorf("apply %s: %w", f, err)
		}
	}
	return nil
}

// requireDB skips in -short mode, otherwise resets the data so each test starts clean.
func requireDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	if testPool == nil {
		t.Skip("skipping integration test in -short mode")
	}
	_, err := testPool.Exec(context.Background(), `TRUNCATE tasks, users RESTART IDENTITY CASCADE`)
	if err != nil {
		t.Fatalf("reset database: %v", err)
	}
	return testPool
}

func createUser(t *testing.T, pool *pgxpool.Pool, email string) int {
	t.Helper()
	u := &domain.User{Email: email, PasswordHash: "not-a-real-hash"}
	if err := postgres.NewUserRepository(pool).Create(context.Background(), u); err != nil {
		t.Fatalf("create user %s: %v", email, err)
	}
	return u.ID
}

func ptr[T any](v T) *T { return &v }