package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func NewTaskRouter(auth *AuthHandler, task *TaskHandler, tokens TokenParser) *gin.Engine {
	r := gin.Default()

	r.GET("/health", func(c *gin.Context){
		c.JSON(http.StatusOK, gin.H{"satus": "ok"})
	})

	v1 := r.Group("/api/v1")
	{
		a := v1.Group("/auth")
		a.POST("/register", auth.Register)
		a.POST("/login", auth.Login)


		t := v1.Group("/task", RequireAuth(tokens))
		t.POST("", task.Create)
		t.GET("", task.List)
		t.GET(":id", task.Get)
		t.PUT(":id", task.Update)
		t.DELETE(":id", task.Delete)
	}
		
	return r
}



