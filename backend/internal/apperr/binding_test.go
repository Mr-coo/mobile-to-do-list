package apperr_test

import (
	"testing"

	"github.com/gin-gonic/gin/binding"

	"taskmanager/internal/apperr"
)

func TestFromBinding_UsesJSONTagName(t *testing.T) {
	type req struct {
		DueDate string `json:"dueAt" binding:"required"`
		Plain   string `binding:"required"`
	}
	err := binding.Validator.ValidateStruct(&req{})
	verr := apperr.FromBinding(err)
	if verr == nil {
		t.Fatal("expected a ValidationError")
	}
	if _, ok := verr.Details["dueAt"]; !ok {
		t.Fatalf("details = %v, want key dueAt", verr.Details)
	}
	if _, ok := verr.Details["Plain"]; !ok {
		t.Fatalf("details = %v, want key Plain for untagged field", verr.Details)
	}
}
