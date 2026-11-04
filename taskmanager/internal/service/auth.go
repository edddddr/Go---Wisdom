package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
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

