package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
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
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
}



func currentUserID(c *gin.Context) int {
	return c.GetInt(userIDKey)
}
