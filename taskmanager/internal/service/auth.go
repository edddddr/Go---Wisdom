package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/edddddr/taskmanager/internal/domain"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type UserRepository interface{
	Create(ctx context.Context, u *domain.User) error
	GetEmail(ctx context.Context, email string) (*domain.User, error)
}


type AuthService struct{
	users UserRepository
	secret []byte
	ttl     time.Duration
	dummyHash  []byte
}

func NewAuthService(users UserRepository, secret []byte, ttl time.Duration) (*AuthService, error){
	dummy, err := bcrypt.GenerateFromPassword([]byte("dumy-passowrd"), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("create dummy hash: %w", err)
	}
	
	return &AuthService{users:users, secret:secret, ttl:ttl, dummyHash:dummy}, nil

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

		slog.InfoContext(ctx, "user registered", "user_id", u.ID)

		return u, nil

}


func (s *AuthService) Login(ctx context.Context, email string, password string) (string, error){
	email = normalizeEmail(email)

	u, err := s.users.GetEmail(ctx, email); 
	
	if errors.Is(err, domain.ErrNotFound) {
		_ = bcrypt.CompareHashAndPassword(s.dummyHash, []byte(password))
		return "", domain.ErrUnauthorized

	}

	if err != nil {
		return "", fmt.Errorf("get user: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil{
		return "", domain.ErrUnauthorized
	}
		
	return s.issueToken(u.ID)
}


func (s *AuthService) issueToken(userID int) (string, error) {
	now := time.Now()
	claims := jwt.RegisteredClaims{
		Subject:   strconv.Itoa(userID),
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(s.ttl)),
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.secret)
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}
	return token, nil
}


func (s *AuthService) ParseToken(tokenStr string) (int, error) {
	claims := &jwt.RegisteredClaims{}
	_, err := jwt.ParseWithClaims(tokenStr, claims,
		func(t *jwt.Token) (any, error) { return s.secret, nil },
		jwt.WithValidMethods([]string{"HS256"}),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		return 0, domain.ErrUnauthorized
	}
	id, err := strconv.Atoi(claims.Subject)
	if err != nil || id <= 0 {
		return 0, domain.ErrUnauthorized
	}
	return id, nil
}