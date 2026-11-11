package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"

	"github.com/edddddr/taskmanager/internal/domain"
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


func bindJSON(c *gin.Context, dst any) bool {
	if err := c.ShouldBindJSON(dst); err != nil {
		respondBindError(c, err)
		return false
	}
	return true
}


func respondBindError(c *gin.Context, err error) {
	var (
		ve        validator.ValidationErrors
		syntaxErr *json.SyntaxError
		typeErr   *json.UnmarshalTypeError
		tooBig    *http.MaxBytesError
	)
	switch {
	case errors.As(err, &tooBig):
		writeError(c, http.StatusRequestEntityTooLarge, "payload_too_large", "request body too large")
	case errors.As(err, &ve):
		details := make([]fieldError, 0, len(ve))
		for _, fe := range ve {
			issue := fe.Tag()
			if fe.Param() != "" {
				issue += "=" + fe.Param()
			}
			details = append(details, fieldError{Field: fe.Field(), Issue: issue})
		}
		writeError(c, http.StatusBadRequest, "validation_failed", "request validation failed", details...)
	case errors.As(err, &typeErr):
		writeError(c, http.StatusBadRequest, "invalid_json", "a field has the wrong type",
			fieldError{Field: typeErr.Field, Issue: "expected " + typeErr.Type.String()})
	case errors.As(err, &syntaxErr), errors.Is(err, io.EOF), errors.Is(err, io.ErrUnexpectedEOF):
		writeError(c, http.StatusBadRequest, "invalid_json", "request body is not valid JSON")
	default:
		writeError(c, http.StatusBadRequest, "invalid_json", "request body could not be read")
	}
}

