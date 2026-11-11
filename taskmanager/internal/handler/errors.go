package handler

import (
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"

	
	"github.com/edddddr/taskmanager/internal/logging"
)

// Make validation errors report JSON field names ("title") instead of Go names ("Title").
func init() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterTagNameFunc(func(f reflect.StructField) string {
			name := strings.SplitN(f.Tag.Get("json"), ",", 2)[0]
			if name == "-" {
				return ""
			}
			return name
		})
	}
}

type fieldError struct {
	Field string `json:"field"`
	Issue string `json:"issue"`
}

type errorBody struct {
	Code      string       `json:"code"`
	Message   string       `json:"message"`
	RequestID string       `json:"request_id,omitempty"`
	Details   []fieldError `json:"details,omitempty"`
}

func writeError(c *gin.Context, status int, code, message string, details ...fieldError) {
	c.AbortWithStatusJSON(status, gin.H{"error": errorBody{
		Code:      code,
		Message:   message,
		RequestID: logging.RequestID(c.Request.Context()),
		Details:   details,
	}})
}
