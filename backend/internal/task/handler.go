package task

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"taskmanager/internal/apperr"
	"taskmanager/internal/httpx"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Register(rg *gin.RouterGroup) {
	rg.POST("/tasks", h.create)
	rg.GET("/tasks", h.list)
	rg.GET("/tasks/:id", h.get)
}

func (h *Handler) create(c *gin.Context) {
	var in CreateInput
	if err := c.ShouldBindJSON(&in); err != nil {
		if verr := apperr.FromBinding(err); verr != nil {
			httpx.WriteError(c, verr)
			return
		}
		httpx.WriteError(c, err)
		return
	}
	if in.Status != "" && !in.Status.Valid() {
		httpx.WriteError(c, &apperr.ValidationError{Details: map[string]string{
			"status": "must be one of todo, in_progress, done",
		}})
		return
	}
	t, err := h.svc.Create(c.Request.Context(), in)
	if err != nil {
		httpx.WriteError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": t})
}

func (h *Handler) list(c *gin.Context) {
	q, err := ParseListQuery(c.Request.URL.Query())
	if err != nil {
		httpx.WriteError(c, err)
		return
	}
	result, err := h.svc.List(c.Request.Context(), q)
	if err != nil {
		httpx.WriteError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *Handler) get(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		httpx.WriteError(c, err)
		return
	}
	t, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		httpx.WriteError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": t})
}

func parseID(c *gin.Context) (uint64, error) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		return 0, &apperr.ValidationError{Details: map[string]string{"id": "must be a positive integer"}}
	}
	return id, nil
}
