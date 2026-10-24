// // package main

// // import (
// // 	// "fmt"
// // 	"net/http"
// // 	"sort"
// // 	"strconv"
// // 	"sync"
// // 	"time"

// // 	"github.com/gin-gonic/gin"
// // )

// // type Task struct {
// // 	ID          int       `json:"id"`
// // 	Title       string    `json:"title"`
// // 	Description string    `json:"description"`
// // 	Priority int          `json:"priority"`
// // 	Done        bool      `json:"done"`
// // 	CreatedAt   time.Time `json:"created_at"`
// // }

// // type createTaskRequest struct {
// // 	Title       string `json:"title" binding:"required,min=1,max=100"`
// // 	Description string `json:"description" binding:"max=500"`
// // 	Priority    *int   `json:"priority" binding:"omitempty,gte=1,lte=5"`
// // }

// // type updateTaskRequest struct {
// // 	Title       *string `json:"title" binding:"omitempty,min=1,max=100"`
// // 	Description *string `json:"description" binding:"omitempty,max=500"`
// // 	Priority    *int   `json:"priority" binding:"omitempty,gte=1,lte=5"`
// // 	Done        *bool   `json:"done"`
// // }

// // type taskHandler struct{
// // 	mu     sync.Mutex
// // 	nextID  int
// // 	tasks  map[int]Task
// // }

// // const defaultPriority = 3

// // func newTaskHandler() *taskHandler{
// // 	return &taskHandler{tasks: make(map[int]Task)}
// // }

// // func (h *taskHandler) create(c *gin.Context){
// // 	var req createTaskRequest
// // 	if err := c.ShouldBindJSON(&req); err != nil{
// // 		c.JSON(http.StatusNotFound, gin.H{"error": err})
// // 		return
// // 	}

// // 	h.mu.Lock()
// // 	defer h.mu.Unlock()

// // 	h.nextID++
// // 	priority := defaultPriority

// // 	if req.Priority != nil {
// // 		priority = *req.Priority
// // 	}

// // 	if req.Priority != nil {
// // 		priority = *req.Priority
// // }

// // 	t := Task{
// // 		ID : h.nextID,
// // 		Title : req.Title,
// // 		Description : req.Description,
// // 		Priority: priority
// // 		CreatedAt  : time.Now().UTC(),
// // 	}

// // 	h.tasks[h.nextID] = t
// // 	c.JSON(http.StatusCreated, t)

// // }

// // func (h *taskHandler) list(c *gin.Context) {
// // 	var doneFilter *bool
// // 	if raw, ok := c.GetQuery("done"); ok {
// // 		v, err := strconv.ParseBool(raw)
// // 		if err != nil {
// // 			c.JSON(http.StatusBadRequest, gin.H{"error": "done must be true or false"})
// // 			return
// // 		}
// // 		doneFilter = &v
// // 	}

// // 	h.mu.Lock()
// // 	out := make([]Task, 0, len(h.tasks))
// // 	for _, t := range h.tasks {
// // 		if doneFilter != nil && t.Done != *doneFilter {
// // 			continue
// // 		}
// // 		out = append(out, t)
// // 	}
// // 	h.mu.Unlock()

// // 	// newest first, as you chose
// // 	sort.Slice(out, func(i, j int) bool { return out[i].ID > out[j].ID })
// // 	c.JSON(http.StatusOK, out)
// // }

// // func (h *taskHandler) get(c *gin.Context){
// // 	id, ok := parseID(c)
// // 	if !ok {
// // 		return
// // 	}

// // 	t, found := h.tasks[id]
// // 	if !found {
// // 		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
// // 		return
// // 	}
// // 	c.JSON(http.StatusOK, t)
// // }

// // func (h *taskHandler) update(c *gin.Context) {
// // 	id, ok := parseID(c)
// // 	if !ok {
// // 		return
// // 	}

// // 	var req updateTaskRequest
// // 	if err := c.ShouldBindJSON(&req); err != nil {
// // 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// // 		return
// // 	}

// // 	h.mu.Lock()
// // 	defer h.mu.Unlock()

// // 	t, found := h.tasks[id]
// // 	if !found {
// // 		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
// // 		return
// // 	}

// // 	if req.Title != nil {
// // 		t.Title = *req.Title
// // 	}
// // 	if req.Description != nil {
// // 		t.Description = *req.Description
// // 	}
// // 	if req.Done != nil {
// // 		t.Done = *req.Done
// // 	}
// // 	if req.Priority != nil {
// // 		t.Priority = *req.Priority
// //     }
// // 	h.tasks[id] = t

// // 	c.JSON(http.StatusOK, t)
// // }

// // func (h *taskHandler) remove(c *gin.Context) {
// // 	id, ok := parseID(c)
// // 	if !ok {
// // 		return
// // 	}
// // l
// // 	h.mu.Lock()
// // 	defer h.mu.Unlock()

// // 	if _, found := h.tasks[id]; !found {
// // 		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
// // 		return
// // 	}
// // 	delete(h.tasks, id)

// // 	c.Status(http.StatusNoContent)
// // }

// // func (h *taskHandler) filter(c *gin.Context){
// // 	done, err := strconv.ParseBool(c.Query("done"))
// // 	if err != nil{
// // 		c.JSON(http.StatusBadRequest, gin.H{"error" : "type error not bool"})
// // 		return
// // 	}
// // 	h.mu.Lock()
// // 	defer h.mu.Unlock()

// // 	out := make([]Task, 0, len(h.tasks))
// // 	for _, t := range h.tasks{
// // 		if t.Done == done{
// // 			out = append(out, t)
// // 		}
// // 	}

// // 	c.JSON(http.StatusOK, out)

// // }

// // func parseID(c *gin.Context) (int, bool){
// // 	id, err := strconv.Atoi(c.Param("id"))
// // 	if err != nil || id <= 0{
// // 		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
// // 		return 0, false
// // 	}else{
// // 		return id, true

// // 	}
// // }

// // func main() {
// // 	r := gin.Default() // includes logger + panic-recovery middleware

// // 	r.GET("/health", func(c *gin.Context) {
// // 		c.JSON(http.StatusOK, gin.H{"status": "ok"})
// // 	})

// // 	h := newTaskHandler()

// // 	v1 := r.Group("/api/v1")
// // 	task := v1.Group("/task")
// // 	task.POST("", h.create)
// // 	task.GET("", h.list)
// // 	task.GET("/:id", h.get)
// // 	task.PATCH("/:id", h.update)
// // 	task.DELETE("/:id", h.remove)

// // 	r.Run(":8080")
// // }

// package main

// import (
// 	"log"

// 	"github.com/edddddr/taskmanager/internal/config"
// 	"github.com/edddddr/taskmanager/internal/handler"
// 	"github.com/edddddr/taskmanager/internal/repository/memory"
// 	"github.com/edddddr/taskmanager/internal/service"
// )

// func main() {
// 	cfg := config.Load()

// 	repo := memory.NewTaskRepository()
// 	svc := service.NewTaskService(repo)
// 	taskHandler := handler.NewTaskHandler(svc)
// 	router := handler.NewRouter(taskHandler)

// 	if err := router.Run(":" + cfg.Port); err != nil {
// 		log.Fatalf("server failed: %v", err)
// 	}
// }


