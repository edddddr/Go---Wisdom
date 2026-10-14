package main

import (
	// "fmt"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type Task struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Done        bool      `json:"done"`
	CreatedAt   time.Time `json:"created_at"`
}


type createTaskRequest struct {
	Title       string `json:"title" binding:"required,min=1,max=100"`
	Description string `json:"description" binding:"max=500"`
}

type updateTaskRequest struct {
	Title       *string `json:"title" binding:"omitempty,min=1,max=100"`
	Description *string `json:"description" binding:"omitempty,max=500"`
	Done        *bool   `json:"done"`
}

type taskHandler struct{
	mu     sync.Mutex
	nextID  int
	tasks  map[int]Task
}


func newTaskHandler() *taskHandler{
	return &taskHandler{tasks: make(map[int]Task)}
}

func (h *taskHandler) create(c *gin.Context){
	var req createTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil{
		c.JSON(http.StatusNotFound, gin.H{"error": err})
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	h.nextID++ 

	t := Task{
		ID : h.nextID,
		Title : req.Title,     
		Description : req.Description,
		CreatedAt  : time.Now().UTC(),
	}

	h.tasks[h.nextID] = t
	c.JSON(http.StatusCreated, t)

}


func (h *taskHandler) list(c *gin.Context){
	h.mu.Lock()
	defer h.mu.Unlock()

	out := make([]Task, 0, len(h.tasks))
	
	for _, t := range h.tasks{
		out = append(out, t)
	}
	
	sort.Slice(out, func(i, j int) bool { return out[i].ID > out[j].ID})

	c.JSON(http.StatusOK, out)

}



func main() {
	r := gin.Default() // includes logger + panic-recovery middleware

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	h := newTaskHandler()




	v1 := r.Group("/api/v1")
	task := v1.Group("/task")
	task.POST("", h.create)
	task.GET("", h.list)
	



		
	r.Run(":8080")
}