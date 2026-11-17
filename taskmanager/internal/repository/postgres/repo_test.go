package postgres_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/edddddr/taskmanager/internal/domain"
	"github.com/edddddr/taskmanager/internal/repository/postgres"
)

func newTask(t *testing.T, repo *postgres.TaskRepository, owner int, title string) *domain.Task {
	t.Helper()
	task := &domain.Task{OwnerID: owner, Title: title, Priority: 3, CreatedAt: time.Now().UTC()}
	if err := repo.Create(context.Background(), task); err != nil {
		t.Fatalf("create task: %v", err)
	}
	return task
}

func TestTaskRepository_OwnershipIsolation(t *testing.T) {
	pool := requireDB(t)
	ctx := context.Background()
	repo := postgres.NewTaskRepository(pool)
	alice := createUser(t, pool, "alice@example.com")
	bob := createUser(t, pool, "bob@example.com")

	task := newTask(t, repo, alice, "secret")

	if _, err := repo.Get(ctx, bob, task.ID); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("Get as bob: err = %v, want ErrNotFound", err)
	}
	if _, err := repo.Update(ctx, bob, task.ID, domain.TaskPatch{Title: ptr("hacked")}); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("Update as bob: err = %v, want ErrNotFound", err)
	}
	if err := repo.Delete(ctx, bob, task.ID); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("Delete as bob: err = %v, want ErrNotFound", err)
	}
	if list, err := repo.List(ctx, bob, domain.TaskFilter{}); err != nil || len(list) != 0 {
		t.Errorf("List as bob = %v, %v; want empty", list, err)
	}

	// Alice's task must be untouched by all of that.
	got, err := repo.Get(ctx, alice, task.ID)
	if err != nil {
		t.Fatalf("Get as alice: %v", err)
	}
	if got.Title != "secret" {
		t.Errorf("title = %q, want %q", got.Title, "secret")
	}
}

func TestTaskRepository_PartialUpdateKeepsOtherFields(t *testing.T) {
	pool := requireDB(t)
	repo := postgres.NewTaskRepository(pool)
	owner := createUser(t, pool, "a@example.com")
	task := newTask(t, repo, owner, "original")

	got, err := repo.Update(context.Background(), owner, task.ID, domain.TaskPatch{Done: ptr(true)})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if !got.Done {
		t.Error("Done = false, want true")
	}
	if got.Title != "original" || got.Priority != 3 {
		t.Errorf("untouched fields changed: %+v", got)
	}
}

func TestTaskRepository_ListFilterAndOrder(t *testing.T) {
	pool := requireDB(t)
	ctx := context.Background()
	repo := postgres.NewTaskRepository(pool)
	owner := createUser(t, pool, "a@example.com")

	first := newTask(t, repo, owner, "one")
	second := newTask(t, repo, owner, "two")
	third := newTask(t, repo, owner, "three")
	if _, err := repo.Update(ctx, owner, second.ID, domain.TaskPatch{Done: ptr(true)}); err != nil {
		t.Fatal(err)
	}

	all, err := repo.List(ctx, owner, domain.TaskFilter{})
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 3 || all[0].ID != third.ID || all[2].ID != first.ID {
		t.Errorf("want 3 tasks newest-first, got %+v", all)
	}

	done, err := repo.List(ctx, owner, domain.TaskFilter{Done: ptr(true)})
	if err != nil {
		t.Fatal(err)
	}
	if len(done) != 1 || done[0].ID != second.ID {
		t.Errorf("done filter: got %+v, want only task %d", done, second.ID)
	}
}

// This is Lesson 3's lost-update scenario, as an executable test.
func TestTaskRepository_ConcurrentPatchesDontLoseUpdates(t *testing.T) {
	pool := requireDB(t)
	ctx := context.Background()
	repo := postgres.NewTaskRepository(pool)
	owner := createUser(t, pool, "a@example.com")

	for i := 0; i < 20; i++ { // repeat: races are probabilistic
		task := newTask(t, repo, owner, "original")

		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()
			if _, err := repo.Update(ctx, owner, task.ID, domain.TaskPatch{Done: ptr(true)}); err != nil {
				t.Errorf("update done: %v", err) // Errorf, not Fatalf, inside goroutines
			}
		}()
		go func() {
			defer wg.Done()
			if _, err := repo.Update(ctx, owner, task.ID, domain.TaskPatch{Title: ptr("renamed")}); err != nil {
				t.Errorf("update title: %v", err)
			}
		}()
		wg.Wait()

		got, err := repo.Get(ctx, owner, task.ID)
		if err != nil {
			t.Fatal(err)
		}
		if !got.Done || got.Title != "renamed" {
			t.Fatalf("iteration %d lost an update: %+v", i, got)
		}
	}
}

func TestUserRepository(t *testing.T) {
	pool := requireDB(t)
	ctx := context.Background()
	repo := postgres.NewUserRepository(pool)

	t.Run("duplicate email conflicts", func(t *testing.T) {
		createUser(t, pool, "dup@example.com")
		err := repo.Create(ctx, &domain.User{Email: "dup@example.com", PasswordHash: "x"})
		if !errors.Is(err, domain.ErrConflict) {
			t.Errorf("err = %v, want ErrConflict", err)
		}
	})

	t.Run("unknown email is not found", func(t *testing.T) {
		_, err := repo.GetByEmail(ctx, "nobody@example.com")
		if !errors.Is(err, domain.ErrNotFound) {
			t.Errorf("err = %v, want ErrNotFound", err)
		}
	})

	t.Run("database rejects mixed-case email", func(t *testing.T) {
		// Bypass the service's normalization to prove the CHECK constraint is a real backstop.
		assertRejected(t, pool, `INSERT INTO users (email, password_hash) VALUES ('Mixed@Example.com', 'x')`)
	})
}

func assertRejected(t *testing.T, pool *pgxpool.Pool, sql string) {
	t.Helper()
	if _, err := pool.Exec(context.Background(), sql); err == nil {
		t.Errorf("expected the database to reject: %s", sql)
	}
}