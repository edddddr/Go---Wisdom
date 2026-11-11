package handler

import (
	"net/http"
	"strings"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"runtime/debug"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/edddddr/taskmanager/internal/logging"
)


type TokenParser interface{
	ParseToken(token string) (int, error) 
}

const userIDKey = "userID"

func RequireAuth(p TokenParser) gin.HandlerFunc {
	return func(c *gin.Context) {
		scheme, token, ok := strings.Cut(c.GetHeader("Authorization"), " ")
		if !ok || !strings.EqualFold(scheme, "Bearer") || token == "" {
			abortUnauthorized(c)
			return
		}
		userID, err := p.ParseToken(token)
		if err != nil {
			abortUnauthorized(c)
			return
		}
		c.Set(userIDKey, userID)
		c.Next()
	}
}


func abortUnauthorized(c *gin.Context) {
	c.Header("WWW-Authenticate", "Bearer")
	writeError(c, http.StatusUnauthorized, "unauthorized", "authentication required")
}



func currentUserID(c *gin.Context) int {
	return c.GetInt(userIDKey)
}


func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader("X-Request-ID")
		if !validRequestID(id) {
			id = newRequestID()
		}

		c.Header("X-Request-ID", id)
		c.Request = c.Request.WithContext(logging.WithRequestID(c.Request.Context(), id))
		c.Next()
	}
}


func validRequestID(id string) bool {
	if id == "" || len(id) > 64 {
		return false
	}
	for _, r := range id {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-', r == '_':
		default:
			return false
		}
	}
	return true
}


func newRequestID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b) // crypto/rand.Read never returns an error on supported platforms
	return hex.EncodeToString(b)
}



func AccessLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		status := c.Writer.Status()
		level := slog.LevelInfo
		switch {
		case status >= 500:
			level = slog.LevelError
		case status >= 400:
			level = slog.LevelWarn
		}
		slog.Log(c.Request.Context(), level, "request",
			"method", c.Request.Method,
			"route", c.FullPath(),
			"status", status,
			"duration_ms", time.Since(start).Milliseconds(),
			"bytes", c.Writer.Size(),
			"client_ip", c.ClientIP(),
		)
	}
}


func Recovery() gin.HandlerFunc {
	return gin.CustomRecoveryWithWriter(io.Discard, func(c *gin.Context, recovered any) {
		slog.ErrorContext(c.Request.Context(), "panic recovered",
			"panic", fmt.Sprint(recovered),
			"stack", string(debug.Stack()),
		)
		writeError(c, http.StatusInternalServerError, "internal_error", "internal server error")
	})
}


func MaxBody(n int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, n)
		c.Next()
	}
}
