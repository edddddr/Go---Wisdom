package handler

import (
	"context"
	"net/http"
	// "net/http"
	// "strconv"

	"github.com/edddddr/taskmanager/internal/domain"
	"github.com/edddddr/taskmanager/internal/service"
	"github.com/gin-gonic/gin"
)


type TaskRepository interface{
	Create(ctx context.Context, in service.CreateInput) (*domain.Task, error)
}

type CreateInput struct { 
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Priority    *int       `json:"priority"`
}

type TaskHandler struct {
	serv TaskRepository
} 


func NewTaskHandler(serv TaskRepository) *TaskHandler{
	return &TaskHandler{serv:serv}
}


func (h *TaskHandler)  Create(c *gin.Context){
	var req  CreateInput
	if err := c.ShouldBindJSON(&req); err != nil{
		c.JSON(400, gin.H{"error": err.Error()})
		return  
	}

	task, err := h.serv.Create(c.Request.Context(), service.CreateInput{
		Title:       req.Title,
		Description: req.Description,
		Priority:    req.Priority,
	})

	if err != nil {
		respondError(c, err)
		return  
	}

	c.JSON(http.StatusOK, task)




}