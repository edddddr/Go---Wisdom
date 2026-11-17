package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/edddddr/taskmanager/internal/domain"
	"github.com/edddddr/taskmanager/internal/service"
)

type fakeUserRepo struct {
	users  map[string]*domain.User
	nextID int
}

func newFakeUserRepo() *fakeUserRepo {
	return &fakeUserRepo{users: make(map[string]*domain.User)}
}

func (f *fakeUserRepo) Create(ctx context.Context, u *domain.User) error {
	if _, exists := f.users[u.Email]; exists {
		return domain.ErrConflict
	}
	f.nextID++
	u.ID = f.nextID
	f.users[u.Email] = u
	return nil
}

func (f *fakeUserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	u, ok := f.users[email]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return u, nil
}

func newAuthService(t *testing.T) *service.AuthService {
	t.Helper() // failures are reported at the caller's line, not here
	svc, err := service.NewAuthService(
		newFakeUserRepo(),
		[]byte("test-secret-test-secret-test-secret!"),
		time.Hour,
	)
	if err != nil {
		t.Fatalf("NewAuthService: %v", err)
	}
	return svc
}

func TestAuthService_RegisterThenLogin(t *testing.T) {
	svc := newAuthService(t)
	ctx := context.Background()

	u, err := svc.Register(ctx, "A@Example.com", "correct-horse-battery")
	if err != nil {
		t.Fatalf("Register: %v", err)
	}

	token, err := svc.Login(ctx, "a@example.com", "correct-horse-battery")
	if err != nil {
		t.Fatalf("Login: %v", err)
	}

	gotID, err := svc.ParseToken(token)
	if err != nil {
		t.Fatalf("ParseToken: %v", err)
	}
	if gotID != u.ID {
		t.Errorf("token user ID = %d, want %d", gotID, u.ID)
	}
}