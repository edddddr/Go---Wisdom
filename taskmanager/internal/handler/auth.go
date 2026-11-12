package handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/edddddr/taskmanager/internal/domain"
)


type AuthService interface {
	Register(ctx context.Context, email string, password string) (*domain.User, error)
	Login(ctx context.Context, email string, password string) (string, error)
}


type registerRequest struct {
	Email string `json:"email" binding:"required,email,max=254"`
	Password string `json:"password" binding:"required,min=8,max=72"`
}

type loginRequest struct{
	Email string `json:"email" binding:"required,email,max=254"`
	Password string `json:"password" binding:"required,max=72"`
}

type AuthHandler struct{
	svc AuthService
}

func NewAuthHandler(svc AuthService) *AuthHandler{
	return &AuthHandler{svc:svc}
}


func (h *AuthHandler) Register(c *gin.Context){
	var req  registerRequest

	if !bindJSON(c, &req) {
		return
	}

	u, err := h.svc.Register(c, req.Email, req.Password)

	if err != nil{
		respondError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": u.ID, "email": u.Email})

}

func (h *AuthHandler) Login(c *gin.Context){
	var req loginRequest

	if !bindJSON(c, &req) {
		return
	}

	token, err := h.svc.Login(c, req.Email, req.Password)

	if err != nil{
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"access_token": token, "token_type": "Bearer"})
}