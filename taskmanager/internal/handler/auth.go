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


func (h *AuthHandler) Register(c *gin.Context){
	var req  registerRequest

	if err := c.ShouldBindJSON(&req); err != nil{
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	u, err := h.svc.Register(c, req.Email, req.Password)

	if err != nil{
		respondError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": u.ID, "email": u.Email})

}