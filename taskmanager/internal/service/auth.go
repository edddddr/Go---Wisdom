package service

import (
	"context"
	"errors"
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


func (s *AuthService) Register(ctx context.Context, email string, password string) (*domain.User, error) {
		email = normalizeEmail(email)

		if email == "" || len(password) < 8 || len(password) > 72 {
			return nil, fmt.Errorf("invalid credentials format: %w", domain.ErrInvalidInput)
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

		if err != nil {
			return nil, fmt.Errorf("hash password: %w", err)
		}

		u := &domain.User{Email:email, PasswordHash:string(hash)}

		if err := s.users.Create(ctx, u); err != nil{
			return nil, fmt.Errorf("create user: %w", err)
		}

		return u, nil

}


func (s *AuthService) Login(ctx context.Context, email string, password string) (string, error){
	email = normalizeEmail(email)

	u, err := s.users.GetEmail(ctx, email); 
	
	if errors.Is(err, domain.ErrNotFound) {
		_ = bcrypt.CompareHashAndPassword(s.dummyHash, []byte(password))
		return "", domain.ErrorUnauthorized

	}

	if err != nil {
		return "", fmt.Errorf("get user: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil{
		return "", domain.ErrorUnauthorized
	}
		
	return "u.ID", nil
}


