package task

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"taskmanager/internal/apperr"
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if in.Status != "" && !in.Status.Valid() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status"})
		return
	}
	t, err := h.svc.Create(c.Request.Context(), in)
	var verr *apperr.ValidationError
	if errors.As(err, &verr) {
		c.JSON(http.StatusBadRequest, gin.H{"error": verr.Error()})
		return
	}
	if errors.Is(err, apperr.ErrDuplicateTitle) {
		c.JSON(http.StatusConflict, gin.H{"error": "task title already exists"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": t})
}

func (h *Handler) list(c *gin.Context) {
	tasks, err := h.svc.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": tasks})
}

func (h *Handler) get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	t, err := h.svc.Get(c.Request.Context(), id)
	if errors.Is(err, apperr.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": t})
}
