// package handler

// import (
// 	"net/http"

// 	"github.com/gin-gonic/gin"
// )

// func NewRouter(tasks *TaskHandler) *gin.Engine {
// 	r := gin.Default()

// 	r.GET("/health", func(c *gin.Context) {
// 		c.JSON(http.StatusOK, gin.H{"status": "ok"})
// 	})

// 	v1 := r.Group("/api/v1")
// 	{
// 		t := v1.Group("/tasks")
// 		t.POST("", tasks.Create)
// 	}
// 	return r
// }