package task_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"taskmanager/internal/server"
	"taskmanager/internal/task"
)

func newRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return server.New(task.NewHandler(task.NewService(newFakeRepo())), nil)
}

func do(r *gin.Engine, method, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestCreateTask_Returns201(t *testing.T) {
	w := do(newRouter(), http.MethodPost, "/api/tasks", `{"title":"Write docs"}`)
	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body)
	}
	var resp struct {
		Data task.Task `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Data.Title != "Write docs" || resp.Data.Status != task.StatusTodo {
		t.Fatalf("unexpected task: %+v", resp.Data)
	}
}

func TestListTasks_ReturnsData(t *testing.T) {
	r := newRouter()
	do(r, http.MethodPost, "/api/tasks", `{"title":"One"}`)
	do(r, http.MethodPost, "/api/tasks", `{"title":"Two"}`)

	w := do(r, http.MethodGet, "/api/tasks", "")
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	var resp struct {
		Data []task.Task `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if len(resp.Data) != 2 {
		t.Fatalf("len(data) = %d, want 2", len(resp.Data))
	}
}

func TestGetTask_NotFound(t *testing.T) {
	w := do(newRouter(), http.MethodGet, "/api/tasks/999", "")
	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", w.Code)
	}
}
