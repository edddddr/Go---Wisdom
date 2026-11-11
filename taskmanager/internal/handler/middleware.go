package handler

import (
	"net/http"
	"strings"
	"crypto/rand"
	"encoding/hex"

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