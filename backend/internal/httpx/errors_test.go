package httpx_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"taskmanager/internal/apperr"
	"taskmanager/internal/httpx"
)

func write(err error) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	httpx.WriteError(c, err)
	return w
}

func TestWriteError_Mappings(t *testing.T) {
	slog.SetDefault(slog.New(slog.DiscardHandler))

	tests := []struct {
		name   string
		err    error
		status int
		code   string
	}{
		{"validation", &apperr.ValidationError{Details: map[string]string{"title": "is required"}}, http.StatusBadRequest, "VALIDATION_ERROR"},
		{"not found", fmt.Errorf("wrapped: %w", apperr.ErrNotFound), http.StatusNotFound, "NOT_FOUND"},
		{"duplicate", apperr.ErrDuplicateTitle, http.StatusConflict, "DUPLICATE_TITLE"},
		{"internal", errors.New("db password is hunter2"), http.StatusInternalServerError, "INTERNAL_ERROR"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := write(tt.err)
			if w.Code != tt.status {
				t.Fatalf("status = %d, want %d", w.Code, tt.status)
			}
			var body httpx.ErrorBody
			if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
				t.Fatal(err)
			}
			if body.Error.Code != tt.code {
				t.Fatalf("code = %q, want %q", body.Error.Code, tt.code)
			}
			if tt.code == "INTERNAL_ERROR" && strings.Contains(w.Body.String(), "hunter2") {
				t.Fatalf("500 body leaks internal error: %s", w.Body)
			}
		})
	}
}

func TestWriteError_OmitsEmptyDetails(t *testing.T) {
	if strings.Contains(write(apperr.ErrNotFound).Body.String(), "details") {
		t.Fatal("details should be omitted when empty")
	}
}
