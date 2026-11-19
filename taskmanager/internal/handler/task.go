package handler

import (
	"context"
	"net/http"
	"strconv"

	"github.com/edddddr/taskmanager/internal/domain"
	"github.com/edddddr/taskmanager/internal/service"
	"github.com/gin-gonic/gin"
)


type TaskRepository interface{
	Create(ctx context.Context, ownerID int, in service.CreateInput) (*domain.Task, error)
	List(ctx context.Context, ownerID int,  in service.ListTasksInput) (*service.TaskPage, error)
	Get(ctx context.Context, ownerID int, id int) (*domain.Task, error) 
	Update(ctx context.Context, ownerID int, id int, in service.UpdateInput) (*domain.Task, error)
	Delete(ctx context.Context, ownerID int, id int) error 
}

type CreateTaskRequest struct { 
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Priority    *int      `json:"priority"`
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
	if !bindJSON(c, &req) {
		return
	}


	task, err := h.serv.Create(c.Request.Context(), currentUserID(c), service.CreateInput{
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
	var in service.ListTasksInput

	if done, ok := c.GetQuery("done"); ok {
		v, err := strconv.ParseBool(done)
		if err != nil {
			writeError(c, http.StatusBadRequest, "invalid_query", "invalid query parameter",
				fieldError{Field: "done", Issue: "must be true or false"})
			return
		}
		in.Done = &v
	}

	if raw, ok := c.GetQuery("limit"); ok {
		v, err := strconv.Atoi(raw)
		if err != nil {
			writeError(c, http.StatusBadRequest, "invalid_query", "invalid query parameter",
				fieldError{Field: "limit", Issue: "must be an integer"})
			return
		}
		in.Limit = &v
	}

	if raw, ok := c.GetQuery("cursor"); ok {
		v, err := strconv.Atoi(raw)
		if err != nil {
			writeError(c, http.StatusBadRequest, "invalid_query", "invalid query parameter",
				fieldError{Field: "cursor", Issue: "must be an integer"})
			return
		}
		in.BeforeID = &v
	}

	page, err := h.serv.List(c.Request.Context(), currentUserID(c), in)

	if err != nil{
		respondError(c, err)
		return	
	}

	c.JSON(http.StatusOK, gin.H{"items": page.Items, "next_cursor": page.NextCursor})

}


func (h *TaskHandler) Get(c *gin.Context){
	id, ok := parseID(c)
	
	if !ok {
		return
	}
	task, err := h.serv.Get(c.Request.Context(), currentUserID(c),  id)
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
	if !bindJSON(c, &req) {
		return
	}
	task, err := h.serv.Update(c.Request.Context(), currentUserID(c), id, service.UpdateInput{
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

	err := h.serv.Delete(c.Request.Context(), currentUserID(c),  id)

	if err != nil{
		respondError(c, err)
		return
	}

	c.Status(http.StatusNoContent)

}


func parseID(c *gin.Context) (int, bool) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		writeError(c, http.StatusBadRequest, "invalid_id", "invalid id",
			fieldError{Field: "id", Issue: "must be a positive integer"})
		return 0, false
	}
	return id, true
}

