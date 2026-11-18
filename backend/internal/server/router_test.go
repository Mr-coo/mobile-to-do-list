package server_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"taskmanager/internal/server"
	"taskmanager/internal/task"
)

func TestSwaggerDoc_Served(t *testing.T) {
	r := server.New(task.NewHandler(nil), nil)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/swagger/doc.json", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}

	var doc struct {
		Paths map[string]map[string]any `json:"paths"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &doc); err != nil {
		t.Fatal(err)
	}
	want := map[string][]string{
		"/tasks":      {"get", "post"},
		"/tasks/{id}": {"get", "put", "delete"},
	}
	for path, methods := range want {
		for _, m := range methods {
			if _, ok := doc.Paths[path][m]; !ok {
				t.Errorf("missing %s %s", m, path)
			}
		}
	}
}

func TestHTTPServer_HasTimeouts(t *testing.T) {
	s := server.HTTPServer(":0", http.NotFoundHandler())
	if s.ReadHeaderTimeout <= 0 || s.ReadTimeout <= 0 || s.WriteTimeout <= 0 || s.IdleTimeout <= 0 {
		t.Fatalf("timeouts not set: header=%v read=%v write=%v idle=%v",
			s.ReadHeaderTimeout, s.ReadTimeout, s.WriteTimeout, s.IdleTimeout)
	}
}

func TestOversizedBody_Returns413(t *testing.T) {
	r := server.New(task.NewHandler(nil), nil)
	body := `{"title":"` + strings.Repeat("a", 1<<20) + `"}`

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/tasks", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want 413; body %s", w.Code, w.Body.String())
	}
	var got struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.Error.Code != "PAYLOAD_TOO_LARGE" {
		t.Fatalf("code = %q, want PAYLOAD_TOO_LARGE", got.Error.Code)
	}
}
