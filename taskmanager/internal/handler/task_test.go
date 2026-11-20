package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/edddddr/taskmanager/internal/domain"
	"github.com/edddddr/taskmanager/internal/service"
)

const testUserID = 42

var errNotImplemented = errors.New("not implemented in this test")

// ---- fakes and stubs ----

type fakeTokens struct{}

func (fakeTokens) ParseToken(tok string) (int, error) {
	if tok == "good-token" {
		return testUserID, nil
	}
	return 0, domain.ErrUnauthorized
}

type fakePinger struct{}

func (fakePinger) Ping(context.Context) error { return nil }

// fakeAuth satisfies AuthService. Task tests never call it.
type fakeAuth struct{}

func (fakeAuth) Register(context.Context, string, string) (*domain.User, error) {
	return nil, errNotImplemented
}
func (fakeAuth) Login(context.Context, string, string) (string, error) {
	return "", errNotImplemented
}

// fakeTaskService is a stub: each test sets only the behavior it needs.
// calls counts how many times the handler reached the service layer.
type fakeTaskService struct {
	calls    int
	createFn func(ctx context.Context, ownerID int, in service.CreateInput) (*domain.Task, error)
	getFn    func(ctx context.Context, ownerID, id int) (*domain.Task, error)
}

func (f *fakeTaskService) Create(ctx context.Context, ownerID int, in service.CreateInput) (*domain.Task, error) {
	f.calls++
	if f.createFn == nil {
		return nil, errNotImplemented
	}
	return f.createFn(ctx, ownerID, in)
}

func (f *fakeTaskService) Get(ctx context.Context, ownerID, id int) (*domain.Task, error) {
	f.calls++
	if f.getFn == nil {
		return nil, errNotImplemented
	}
	return f.getFn(ctx, ownerID, id)
}

func (f *fakeTaskService) List(context.Context, int, service.ListTasksInput) (*service.TaskPage, error) {
	f.calls++
	return nil, errNotImplemented
}

func (f *fakeTaskService) Update(context.Context, int, int, service.UpdateInput) (*domain.Task, error) {
	f.calls++
	return nil, errNotImplemented
}

func (f *fakeTaskService) Delete(context.Context, int, int) error {
	f.calls++
	return errNotImplemented
}

// ---- helpers ----

// newTestServer builds the REAL router (all middleware included) over fakes.
func newTestServer(t *testing.T, tasks *fakeTaskService) http.Handler {
	t.Helper()
	gin.SetMode(gin.TestMode)

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel) // stops the limiters' background cleanup goroutines

	r, err := NewRouter(Options{
		Auth:           NewAuthHandler(fakeAuth{}),
		Tasks:          NewTaskHandler(tasks),
		Tokens:         fakeTokens{},
		DB:             fakePinger{},
		GlobalLimiter:  NewIPLimiter(ctx, 1000, 1000), // high: limiter isn't under test here
		AuthLimiter:    NewIPLimiter(ctx, 1000, 1000),
		RequestTimeout: 5 * time.Second,
		MaxBodyBytes:   1 << 20,
	})
	if err != nil {
		t.Fatalf("NewRouter: %v", err)
	}
	return r
}

func doRequest(h http.Handler, method, path, body, token string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	return w
}

func errorCode(t *testing.T, w *httptest.ResponseRecorder) string {
	t.Helper()
	var resp struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("response is not the error envelope: %v\nbody: %s", err, w.Body.String())
	}
	return resp.Error.Code
}

// ---- tests ----

func TestTasks_RequireAuth(t *testing.T) {
	headers := map[string]string{
		"no header":    "",
		"wrong scheme": "Basic abc",
		"bad token":    "Bearer nope",
	}
	for name, header := range headers {
		t.Run(name, func(t *testing.T) {
			svc := &fakeTaskService{}
			req := httptest.NewRequest(http.MethodGet, "/api/v1/tasks", nil)
			if header != "" {
				req.Header.Set("Authorization", header)
			}
			w := httptest.NewRecorder()
			newTestServer(t, svc).ServeHTTP(w, req)

			if w.Code != http.StatusUnauthorized {
				t.Errorf("status = %d, want 401", w.Code)
			}
			if got := errorCode(t, w); got != "unauthorized" {
				t.Errorf("code = %q, want unauthorized", got)
			}
			if svc.calls != 0 {
				t.Errorf("service was called %d times; auth must block first", svc.calls)
			}
		})
	}
}

func TestTasks_Create_RejectsBadInput(t *testing.T) {
	tests := []struct {
		name     string
		body     string
		wantCode string
	}{
		{"missing title", `{}`, "validation_failed"},
		{"priority out of range", `{"title":"x","priority":9}`, "validation_failed"},
		{"title too long", `{"title":"` + strings.Repeat("a", 101) + `"}`, "validation_failed"},
		{"malformed json", `{bad`, "invalid_json"},
		{"wrong type", `{"title":5}`, "invalid_json"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := &fakeTaskService{}
			w := doRequest(newTestServer(t, svc), http.MethodPost, "/api/v1/tasks", tc.body, "good-token")

			if w.Code != http.StatusBadRequest {
				t.Errorf("status = %d, want 400", w.Code)
			}
			if got := errorCode(t, w); got != tc.wantCode {
				t.Errorf("code = %q, want %q", got, tc.wantCode)
			}
			if svc.calls != 0 {
				t.Errorf("invalid input reached the service (%d calls)", svc.calls)
			}
		})
	}
}

func TestTasks_Create_UsesOwnerFromToken(t *testing.T) {
	var gotOwner int
	var gotInput service.CreateInput
	svc := &fakeTaskService{
		createFn: func(_ context.Context, ownerID int, in service.CreateInput) (*domain.Task, error) {
			gotOwner, gotInput = ownerID, in
			return &domain.Task{ID: 1, OwnerID: ownerID, Title: in.Title, Priority: 3}, nil
		},
	}

	// A client must not be able to choose the owner by sending owner_id.
	body := `{"title":"write tests","owner_id":999}`
	w := doRequest(newTestServer(t, svc), http.MethodPost, "/api/v1/tasks", body, "good-token")

	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201; body: %s", w.Code, w.Body.String())
	}
	if gotOwner != testUserID {
		t.Errorf("owner = %d, want %d (from the token)", gotOwner, testUserID)
	}
	if gotInput.Priority != nil {
		t.Errorf("priority = %v, want nil so the service applies its default", *gotInput.Priority)
	}
}

func TestTasks_Get_ErrorMapping(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantCode   string
	}{
		{"not found", fmt.Errorf("get task 1: %w", domain.ErrNotFound), 404, "not_found"},
		{"invalid input", &domain.InvalidInputError{Field: "title", Reason: "must not be blank"}, 400, "invalid_input"},
		{"deadline", fmt.Errorf("query: %w", context.DeadlineExceeded), 504, "timeout"},
		{"unexpected", errors.New("dial tcp 10.0.0.5:5432: connection refused"), 500, "internal_error"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := &fakeTaskService{
				getFn: func(context.Context, int, int) (*domain.Task, error) { return nil, tc.err },
			}
			w := doRequest(newTestServer(t, svc), http.MethodGet, "/api/v1/tasks/1", "", "good-token")

			if w.Code != tc.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tc.wantStatus)
			}
			if got := errorCode(t, w); got != tc.wantCode {
				t.Errorf("code = %q, want %q", got, tc.wantCode)
			}
			// Internals must never reach the client.
			if strings.Contains(w.Body.String(), "10.0.0.5") {
				t.Errorf("response leaks internal details: %s", w.Body.String())
			}
		})
	}
}

func TestTasks_Get_InvalidID(t *testing.T) {
	for _, id := range []string{"abc", "0", "-5"} {
		t.Run(id, func(t *testing.T) {
			svc := &fakeTaskService{}
			w := doRequest(newTestServer(t, svc), http.MethodGet, "/api/v1/tasks/"+id, "", "good-token")

			if w.Code != http.StatusBadRequest {
				t.Errorf("status = %d, want 400", w.Code)
			}
			if svc.calls != 0 {
				t.Errorf("service was called for an invalid id")
			}
		})
	}
}