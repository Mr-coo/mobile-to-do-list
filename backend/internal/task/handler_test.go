package task_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"taskmanager/internal/httpx"
	"taskmanager/internal/server"
	"taskmanager/internal/task"
)

func newRouter() *gin.Engine {
	return newRouterWith(task.NewService(newFakeRepo(), nil))
}

func newRouterWith(svc *task.Service) *gin.Engine {
	gin.SetMode(gin.TestMode)
	return server.New(task.NewHandler(svc), nil)
}

func do(r *gin.Engine, method, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func errorOf(t *testing.T, w *httptest.ResponseRecorder) httpx.ErrorDetail {
	t.Helper()
	var body httpx.ErrorBody
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("body is not an error envelope: %s", w.Body)
	}
	return body.Error
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

func TestCreateTask_DuplicateTitle_Returns409(t *testing.T) {
	r := newRouter()
	first := do(r, http.MethodPost, "/api/tasks", `{"title":"Buy milk"}`)
	if first.Code != http.StatusCreated {
		t.Fatalf("first status = %d, body = %s", first.Code, first.Body)
	}

	w := do(r, http.MethodPost, "/api/tasks", `{"title":"buy milk"}`)
	if w.Code != http.StatusConflict {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body)
	}
	if e := errorOf(t, w); e.Code != "DUPLICATE_TITLE" {
		t.Fatalf("code = %q", e.Code)
	}
}

func TestCreateTask_WhitespaceTitle_Returns400(t *testing.T) {
	w := do(newRouter(), http.MethodPost, "/api/tasks", `{"title":"   "}`)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body)
	}
	e := errorOf(t, w)
	if e.Code != "VALIDATION_ERROR" || e.Details["title"] != "is required" {
		t.Fatalf("error = %+v", e)
	}
}

func TestCreateTask_TrimsTitle(t *testing.T) {
	r := newRouter()
	w := do(r, http.MethodPost, "/api/tasks", `{"title":" Buy milk "}`)
	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body)
	}
	var resp struct {
		Data task.Task `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Data.Title != "Buy milk" {
		t.Fatalf("title = %q", resp.Data.Title)
	}

	w = do(r, http.MethodPost, "/api/tasks", `{"title":"Buy milk"}`)
	if w.Code != http.StatusConflict {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body)
	}
}

func TestGetTask_NonNumericID_Returns400(t *testing.T) {
	w := do(newRouter(), http.MethodGet, "/api/tasks/abc", "")
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body)
	}
	e := errorOf(t, w)
	if e.Code != "VALIDATION_ERROR" || e.Details["id"] != "must be a positive integer" {
		t.Fatalf("error = %+v", e)
	}
}

func TestGetTask_NotFound_Body(t *testing.T) {
	w := do(newRouter(), http.MethodGet, "/api/tasks/999", "")
	if e := errorOf(t, w); e.Code != "NOT_FOUND" || e.Message != "Task not found" {
		t.Fatalf("error = %+v", e)
	}
}

func TestCreateTask_MalformedJSON_Returns400(t *testing.T) {
	for _, body := range []string{`{"title":`, ``, `{"title":123}`} {
		w := do(newRouter(), http.MethodPost, "/api/tasks", body)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("body %q: status = %d", body, w.Code)
		}
		if e := errorOf(t, w); e.Code != "VALIDATION_ERROR" || e.Details["body"] != "invalid JSON" {
			t.Fatalf("body %q: error = %+v", body, e)
		}
	}
}

func TestCreateTask_MissingTitle_Details(t *testing.T) {
	w := do(newRouter(), http.MethodPost, "/api/tasks", `{}`)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", w.Code)
	}
	if e := errorOf(t, w); e.Details["title"] != "is required" {
		t.Fatalf("error = %+v", e)
	}
}

func TestCreateTask_InvalidStatus_Details(t *testing.T) {
	w := do(newRouter(), http.MethodPost, "/api/tasks", `{"title":"x","status":"nope"}`)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", w.Code)
	}
	if e := errorOf(t, w); e.Details["status"] != "must be one of todo, in_progress, done" {
		t.Fatalf("error = %+v", e)
	}
}

func TestUnknownRoute_Returns404JSON(t *testing.T) {
	w := do(newRouter(), http.MethodGet, "/api/nope", "")
	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d", w.Code)
	}
	if e := errorOf(t, w); e.Code != "NOT_FOUND" || e.Message != "Route not found" {
		t.Fatalf("error = %+v", e)
	}
}

func TestWrongMethod_Returns405JSON(t *testing.T) {
	w := do(newRouter(), http.MethodDelete, "/api/tasks", "")
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d", w.Code)
	}
	if e := errorOf(t, w); e.Code != "METHOD_NOT_ALLOWED" {
		t.Fatalf("error = %+v", e)
	}
}

func TestListTasks_InvalidStatus_Returns400(t *testing.T) {
	w := do(newRouter(), http.MethodGet, "/api/tasks?status=bogus", "")
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", w.Code)
	}
	e := errorOf(t, w)
	if e.Code != "VALIDATION_ERROR" || e.Details["status"] == "" {
		t.Fatalf("error = %+v", e)
	}
}

func TestListTasks_Meta(t *testing.T) {
	r := newRouter()
	for i := 0; i < 25; i++ {
		do(r, http.MethodPost, "/api/tasks", fmt.Sprintf(`{"title":"Task %d"}`, i))
	}

	w := do(r, http.MethodGet, "/api/tasks?limit=10&page=3", "")
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body)
	}
	var resp task.ListResult
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if len(resp.Data) != 5 || resp.Meta.TotalPages != 3 || resp.Meta.Total != 25 || resp.Meta.Page != 3 || resp.Meta.Limit != 10 {
		t.Fatalf("data = %d, meta = %+v", len(resp.Data), resp.Meta)
	}
	if !strings.Contains(w.Body.String(), `"total_pages":3`) {
		t.Fatalf("body = %s", w.Body)
	}
}

func createTask(t *testing.T, r *gin.Engine, title string) uint64 {
	t.Helper()
	w := do(r, http.MethodPost, "/api/tasks", fmt.Sprintf(`{"title":%q}`, title))
	var resp struct {
		Data task.Task `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	return resp.Data.ID
}

func TestUpdateTask_Returns200(t *testing.T) {
	r := newRouter()
	id := createTask(t, r, "Old")

	w := do(r, http.MethodPut, fmt.Sprintf("/api/tasks/%d", id),
		`{"title":"New","status":"done","assignee":"bob","description":"d"}`)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body)
	}
	var resp struct {
		Data task.Task `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Data.ID != id || resp.Data.Title != "New" || resp.Data.Status != task.StatusDone {
		t.Fatalf("unexpected task: %+v", resp.Data)
	}
}

func TestUpdateTask_MissingStatus_Returns400(t *testing.T) {
	r := newRouter()
	id := createTask(t, r, "Old")

	w := do(r, http.MethodPut, fmt.Sprintf("/api/tasks/%d", id), `{"title":"New"}`)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body)
	}
	e := errorOf(t, w)
	if e.Code != "VALIDATION_ERROR" || e.Details["status"] == "" {
		t.Fatalf("unexpected error: %+v", e)
	}
}

func TestUpdateTask_NotFound_Returns404(t *testing.T) {
	w := do(newRouter(), http.MethodPut, "/api/tasks/99", `{"title":"New","status":"todo"}`)
	if w.Code != http.StatusNotFound || errorOf(t, w).Code != "NOT_FOUND" {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body)
	}
}

func TestUpdateTask_DuplicateTitle_Returns409(t *testing.T) {
	r := newRouter()
	createTask(t, r, "One")
	id := createTask(t, r, "Two")

	w := do(r, http.MethodPut, fmt.Sprintf("/api/tasks/%d", id), `{"title":"one","status":"todo"}`)
	if w.Code != http.StatusConflict || errorOf(t, w).Code != "DUPLICATE_TITLE" {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body)
	}
}

func TestDeleteTask_Returns204(t *testing.T) {
	r := newRouter()
	id := createTask(t, r, "Write docs")
	w := do(r, http.MethodDelete, fmt.Sprintf("/api/tasks/%d", id), "")
	if w.Code != http.StatusNoContent {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body)
	}
	if w.Body.Len() != 0 {
		t.Fatalf("body = %q, want empty", w.Body)
	}
}

func TestDeleteTask_NotFound_Returns404(t *testing.T) {
	w := do(newRouter(), http.MethodDelete, "/api/tasks/999", "")
	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d", w.Code)
	}
	if e := errorOf(t, w); e.Code != "NOT_FOUND" {
		t.Fatalf("code = %s", e.Code)
	}
}

func TestDeleteTask_NonNumericID_Returns400(t *testing.T) {
	w := do(newRouter(), http.MethodDelete, "/api/tasks/abc", "")
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", w.Code)
	}
	if e := errorOf(t, w); e.Code != "VALIDATION_ERROR" {
		t.Fatalf("code = %s", e.Code)
	}
}

func TestDeleteTask_ThenGet_Returns404(t *testing.T) {
	r := newRouter()
	id := createTask(t, r, "Write docs")
	path := fmt.Sprintf("/api/tasks/%d", id)
	if w := do(r, http.MethodDelete, path, ""); w.Code != http.StatusNoContent {
		t.Fatalf("delete status = %d", w.Code)
	}
	w := do(r, http.MethodGet, path, "")
	if w.Code != http.StatusNotFound {
		t.Fatalf("get status = %d", w.Code)
	}
	if e := errorOf(t, w); e.Code != "NOT_FOUND" {
		t.Fatalf("code = %s", e.Code)
	}
}
