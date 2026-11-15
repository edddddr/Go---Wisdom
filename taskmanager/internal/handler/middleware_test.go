package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func newTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RequestID())
	r.GET("/x", func(c *gin.Context) { c.Status(http.StatusNoContent) })
	return r
}

func TestRequestID_GeneratedWhenMissing(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/x", nil)

	newTestRouter().ServeHTTP(w, req)

	if w.Header().Get("X-Request-ID") == "" {
		t.Error("expected X-Request-ID header to be set")
	}
}