package handler

import (
	"context"
	// "net/http"

	// "github.com/gin-gonic/gin"

	"github.com/edddddr/taskmanager/internal/domain"
)


type AuthService interface {
	Register(ctx context.Context, email string, password string) (*domain.User, error)
	Login(ctx context.Context, email string, password string) (string, error)
}


type registerRequest struct {
	Email string `json:"email" binding:"required,email,max=254"`
	Password string `json:"password" binding:"required, min=8, max=72"`
}

type loginRequest struct{
	Email string `json:"email" binding:"required,email,max=254"`
	Password string `json:"password" binding:"required, max=72"`
}

type AuthHandler struct{
	svc AuthService
}

func NewAuthHandler(svc AuthService) *AuthHandler{
	return &AuthHandler{svc:svc}
}

