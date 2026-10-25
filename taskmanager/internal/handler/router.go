package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func NewTaskRouter(task *TaskHandler) *gin.Engine {
	r := gin.Default()

	r.GET("/health", func(c *gin.Context){
		c.JSON(http.StatusOK, gin.H{"satus": "ok"})
	})

	v1 := r.Group("/api/v1")
	{
		t := v1.Group("/task")
		t.POST("", task.Create)
		t.GET("", task.List)
		t.GET(":id", task.Get)
	}
		
	return r
}



