package main

import (
	"log"
	"net/http"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// ---- Domain ----

type Task struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Done        bool      `json:"done"`
	CreatedAt   time.Time `json:"created_at"`
}

// ---- Request types (what clients are allowed to send) ----

type createTaskRequest struct {
	Title       string `json:"title" binding:"required,min=1,max=100"`
	Description string `json:"description" binding:"max=500"`
}

type updateTaskRequest struct {
	Title       *string `json:"title" binding:"omitempty,min=1,max=100"`
	Description *string `json:"description" binding:"omitempty,max=500"`
	Done        *bool   `json:"done"`
}

// ---- Handler with temporary in-memory storage ----

type taskHandler struct {
	mu     sync.Mutex
	nextID int
	tasks  map[int]Task
}

func newTaskHandler() *taskHandler {
	return &taskHandler{tasks: make(map[int]Task)}
}

func (h *taskHandler) create(c *gin.Context) {
	var req createTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	h.nextID++
	t := Task{
		ID:          h.nextID,
		Title:       req.Title,
		Description: req.Description,
		CreatedAt:   time.Now().UTC(),
	}
	h.tasks[t.ID] = t

	c.JSON(http.StatusCreated, t)
}

func (h *taskHandler) list(c *gin.Context) {
	h.mu.Lock()
	defer h.mu.Unlock()

	out := make([]Task, 0, len(h.tasks))
	for _, t := range h.tasks {
		out = append(out, t)
	}
	// Map iteration order is random in Go, so sort for stable output.
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })

	c.JSON(http.StatusOK, out)
}

func (h *taskHandler) get(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	t, found := h.tasks[id]
	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
	}
	c.JSON(http.StatusOK, t)
}

func (h *taskHandler) update(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	var req updateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	t, found := h.tasks[id]
	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
	}

	if req.Title != nil {
		t.Title = *req.Title
	}
	if req.Description != nil {
		t.Description = *req.Description
	}
	if req.Done != nil {
		t.Done = *req.Done
	}
	h.tasks[id] = t

	c.JSON(http.StatusOK, t)
}

func (h *taskHandler) remove(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	if _, found := h.tasks[id]; !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
	}
	delete(h.tasks, id)

	c.Status(http.StatusNoContent)
}

// parseID extracts and validates the :id path param.
// It writes the 400 response itself, so callers just return when ok is false.
func parseID(c *gin.Context) (int, bool) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return 0, false
	}
	return id, true
}

func main() {
	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	h := newTaskHandler()

	v1 := r.Group("/api/v1")
	tasks := v1.Group("/tasks")
	{
		tasks.POST("", h.create)
		tasks.GET("", h.list)
		tasks.GET("/:id", h.get)
		tasks.PATCH("/:id", h.update)
		tasks.DELETE("/:id", h.remove)
	}

	if err := r.Run(":8080"); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}