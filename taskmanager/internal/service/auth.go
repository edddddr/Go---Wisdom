package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/edddddr/taskmanager/internal/domain"
)

type UserRepository interface{
	Create(ctx context.Context, u *domain.User) error
	GetEmail(ctx context.Context, email string) (*domain.User, error)
}


type AuthService struct{
	users UserRepository
	secrete []byte
	ttl     time.Duration
	dummyHash  []byte
}
func NewAuthService(users UserRepository, secrete []byte, ttl time.Duration) (*AuthService, error){
	dummy, err := bcrypt.GenerateFromPassword([]byte("dumy-passowrd"), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("create dummy hash: %w", err)
	}
	
	return &AuthService{users:users, secrete:secrete, ttl:ttl, dummyHash:dummy}, nil

}


func normalizeEmail(e string) string{
	return strings.ToLower(strings.TrimSpace(e))
}

