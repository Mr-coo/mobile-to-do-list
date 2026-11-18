package httpx

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"taskmanager/internal/apperr"
)

type ErrorBody struct {
	Error ErrorDetail `json:"error"`
}

type ErrorDetail struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Details map[string]string `json:"details,omitempty"`
}

func Write(c *gin.Context, status int, code, message string, details map[string]string) {
	c.AbortWithStatusJSON(status, ErrorBody{Error: ErrorDetail{Code: code, Message: message, Details: details}})
}

// WriteError maps a domain error to the standard error response. Unknown
// errors are logged and hidden behind a generic 500.
func WriteError(c *gin.Context, err error) {
	var verr *apperr.ValidationError
	var tooLarge *http.MaxBytesError
	switch {
	case errors.As(err, &tooLarge):
		Write(c, http.StatusRequestEntityTooLarge, "PAYLOAD_TOO_LARGE", "Request body too large", nil)
	case errors.As(err, &verr):
		Write(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request", verr.Details)
	case errors.Is(err, apperr.ErrNotFound):
		Write(c, http.StatusNotFound, "NOT_FOUND", "Task not found", nil)
	case errors.Is(err, apperr.ErrDuplicateTitle):
		Write(c, http.StatusConflict, "DUPLICATE_TITLE", "A task with this title already exists", nil)
	default:
		slog.Error("internal error", "method", c.Request.Method, "path", c.Request.URL.Path, "err", err)
		Write(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error", nil)
	}
}
