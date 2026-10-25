package handler

import (
	"context"
	"net/http"
	"strconv"

	// "net/http"
	// "strconv"

	"github.com/edddddr/taskmanager/internal/domain"
	"github.com/edddddr/taskmanager/internal/service"
	"github.com/gin-gonic/gin"
)


type TaskRepository interface{
	Create(ctx context.Context, in service.CreateInput) (*domain.Task, error)
	List(ctx context.Context, f domain.TaskFilter) ([]domain.Task, error)
	Get(ctx context.Context, id int) (*domain.Task, error) 
	Update(ctx context.Context, id int, in service.UpdateInput) (*domain.Task, error)
	Delete(ctx context.Context, id int) error 
}

type CreateTaskRequest struct { 
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Priority    *int       `json:"priority"`
}


type UpdateTaskRequest struct { 
	Title       *string    `json:"title"`
	Description *string    `json:"description"`
	Priority    *int       `json:"priority"`
	Done        *bool      `json:"done"`
}

type TaskHandler struct {
	serv TaskRepository
} 


func NewTaskHandler(serv TaskRepository) *TaskHandler{
	return &TaskHandler{serv:serv}
}


func (h *TaskHandler)  Create(c *gin.Context){
	var req  CreateTaskRequest
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



func (h *TaskHandler) List(c *gin.Context){
	var filter domain.TaskFilter 

	if done, ok := c.GetQuery("done"); ok {
		v, err := strconv.ParseBool(done)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error" : "The done must be ture of false"})
			return
		}
		filter.Done = &v
	}

	tasks, err := h.serv.List(c.Request.Context(), filter)

	if err != nil{
		respondError(c, err)
		return	
	}

	c.JSON(http.StatusOK, tasks)

}


func (h *TaskHandler) Get(c *gin.Context){
	id, ok := parseID(c)
	
	if !ok {
		return
	}
	task, err := h.serv.Get(c.Request.Context(), id)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, task)

}


func (h *TaskHandler) Update(c *gin.Context){
	id, ok := parseID(c)

	if !ok{
		return
	}

	var req UpdateTaskRequest
	if err := c.ShouldBind(&req); err != nil{
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	task, err := h.serv.Update(c.Request.Context(), id, service.UpdateInput{
		Title: req.Title,
		Description: req.Description,
		Priority: req.Priority,
		Done: req.Done,
	})

	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, task)
}

func (h *TaskHandler) Delete(c *gin.Context){
	id, ok := parseID(c)

	if !ok{
		return
	}

	err := h.serv.Delete(c.Request.Context(), id)

	if err != nil{
		respondError(c, err)
		return
	}

	c.Status(http.StatusNoContent)

}


func parseID(c *gin.Context) (int, bool) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return 0, false
	}
	return id, true
}

