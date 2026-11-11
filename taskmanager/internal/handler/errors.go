package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
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


func respondError(c *gin.Context, err error) {
	var invalid *domain.InvalidInputError

	switch {
	case errors.As(err, &invalid):
		writeError(c, http.StatusBadRequest, "invalid_input", "invalid input",
			fieldError{Field: invalid.Field, Issue: invalid.Reason})
	case errors.Is(err, domain.ErrInvalidInput):
		writeError(c, http.StatusBadRequest, "invalid_input", "invalid input")
	case errors.Is(err, domain.ErrNotFound):
		writeError(c, http.StatusNotFound, "not_found", "resource not found")
	case errors.Is(err, domain.ErrUnauthorized):
		writeError(c, http.StatusUnauthorized, "unauthorized", "invalid credentials")
	case errors.Is(err, domain.ErrConflict):
		writeError(c, http.StatusConflict, "conflict", "resource already exists")
	case errors.Is(err, context.DeadlineExceeded):
		slog.WarnContext(c.Request.Context(), "request deadline exceeded", "error", err)
		writeError(c, http.StatusGatewayTimeout, "timeout", "request took too long")
	case errors.Is(err, context.Canceled):
		// The client disconnected, so nobody is listening. 499 is nginx's convention.
		c.AbortWithStatus(499)
	default:
		// The one place unexpected errors get logged. The client sees nothing internal.
		slog.ErrorContext(c.Request.Context(), "unhandled error", "error", err)
		writeError(c, http.StatusInternalServerError, "internal_error", "internal server error")
	}
}