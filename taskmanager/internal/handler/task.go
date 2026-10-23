package handler

import (
	"context"
	"net/http"
	"strconv"

	// "strconv"

	"github.com/gin-gonic/gin"

	"github.com/edddddr/taskmanager/internal/domain"
	"github.com/edddddr/taskmanager/internal/service"
)

// TaskService is what the handler NEEDS from the layer below.
type TaskService interface {
	Create(ctx context.Context, in service.CreateTaskInput) (*domain.Task, error)
	List(ctx context.Context, f domain.TaskFilter) ([]domain.Task, error)
}

type createTaskRequest struct {
	Title       string `json:"title" binding:"required,min=1,max=100"`
	Description string `json:"description" binding:"max=500"`
	Priority    *int   `json:"priority" binding:"omitempty,gte=1,lte=5"`
}



type TaskHandler struct {
	svc TaskService
}

func NewTaskHandler(svc TaskService) *TaskHandler {
	return &TaskHandler{svc: svc}
}

func (h *TaskHandler) Create(c *gin.Context) {
	var req createTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	task, err := h.svc.Create(c.Request.Context(), service.CreateTaskInput{
		Title:       req.Title,
		Description: req.Description,
		Priority:    req.Priority,
	})
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, task)
}


func (h *TaskHandler) List(c *gin.Context) {
	var filter domain.TaskFilter
	if raw, ok := c.GetQuery("done"); ok {
		v, err := strconv.ParseBool(raw)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "done must be true or false"})
			return
		}
		filter.Done = &v
	}

	tasks, err := h.svc.List(c.Request.Context(), filter)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, tasks)
}