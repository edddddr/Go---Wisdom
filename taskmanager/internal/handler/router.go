package handler

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type Pinger interface {
	Ping(ctx context.Context) error
}

type Options struct {
	Auth   *AuthHandler
	Tasks  *TaskHandler
	Tokens TokenParser
	DB     Pinger

	GlobalLimiter *IPLimiter
	AuthLimiter   *IPLimiter

	RequestTimeout time.Duration
	MaxBodyBytes   int64
	TrustedProxies []string // empty means: trust no proxy headers
}



func NewRouter(o Options) (*gin.Engine, error) {
	r := gin.New()

	if err := r.SetTrustedProxies(o.TrustedProxies); err != nil {
		return nil, fmt.Errorf("set trusted proxies: %w", err)
	}

	r.Use(RequestID(), AccessLog(), Recovery(), MaxBody(o.MaxBodyBytes), Timeout(o.RequestTimeout))

	r.HandleMethodNotAllowed = true
	r.NoRoute(func(c *gin.Context) {
		writeError(c, http.StatusNotFound, "not_found", "route not found")
	})
	r.NoMethod(func(c *gin.Context) {
		writeError(c, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
	})

	

	r.GET("/health", func(c *gin.Context){
		c.JSON(http.StatusOK, gin.H{"satus": "ok"})
	})


	r.GET("/readyz", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()
		if err := o.DB.Ping(ctx); err != nil {
			slog.ErrorContext(ctx, "readiness check failed", "error", err)
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unavailable"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	})



	v1 := r.Group("/api/v1",  RateLimit(o.GlobalLimiter))
	{
		a := v1.Group("/auth", RateLimit(o.AuthLimiter))
		a.POST("/register", o.Auth.Register)
		a.POST("/login", o.Auth.Login)


		t := v1.Group("/task", RequireAuth(o.Tokens))
		t.POST("", o.Tasks.Create)
		t.GET("", o.Tasks.List)
		t.GET(":id", o.Tasks.Get)
		t.PUT(":id", o.Tasks.Update)
		t.DELETE(":id", o.Tasks.Delete)
	}
		
	return r, nil
}



