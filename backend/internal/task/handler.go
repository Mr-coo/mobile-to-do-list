package task

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"taskmanager/internal/apperr"
	"taskmanager/internal/httpx"
)

type TaskResponse struct {
	Data Task `json:"data"`
}

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
	rg.PUT("/tasks/:id", h.update)
	rg.DELETE("/tasks/:id", h.delete)
}

type CreateTaskRequest struct {
	Title       string  `json:"title" binding:"required,max=255"`
	Description *string `json:"description" binding:"omitempty,max=10000"`
	Status      Status  `json:"status" binding:"omitempty,oneof=todo in_progress done"`
	Assignee    *string `json:"assignee" binding:"omitempty,max=100"`
}

// @Summary	Create a task
// @Tags		tasks
// @Accept		json
// @Produce	json
// @Param		body	body		CreateTaskRequest	true	"Task to create"
// @Success	201		{object}	TaskResponse
// @Failure	400		{object}	httpx.ErrorBody
// @Failure	409		{object}	httpx.ErrorBody
// @Failure	413		{object}	httpx.ErrorBody
// @Failure	500		{object}	httpx.ErrorBody
// @Router		/tasks [post]
func (h *Handler) create(c *gin.Context) {
	var req CreateTaskRequest
	if !bindJSON(c, &req) {
		return
	}
	t, err := h.svc.Create(c.Request.Context(), CreateInput(req))
	if err != nil {
		httpx.WriteError(c, err)
		return
	}
	c.JSON(http.StatusCreated, TaskResponse{Data: *t})
}

// @Summary	List tasks
// @Tags		tasks
// @Produce	json
// @Param		status		query		string	false	"Filter by status"	Enums(todo, in_progress, done)
// @Param		keyword		query		string	false	"Search in title and description"	maxlength(100)
// @Param		assignee	query		string	false	"Filter by assignee"	maxlength(100)
// @Param		page		query		int		false	"Page number"							default(1)																				minimum(1)
// @Param		limit		query		int		false	"Items per page"						default(10)																				minimum(1)	maximum(100)
// @Param		sort		query		string	false	"Sort field, prefix - for descending"	Enums(created_at, -created_at, updated_at, -updated_at, title, -title, status, -status)	default(-created_at)
// @Success	200			{object}	ListResult
// @Header		200			{string}	X-Cache	"HIT or MISS"
// @Failure	400			{object}	httpx.ErrorBody
// @Failure	500			{object}	httpx.ErrorBody
// @Router		/tasks [get]
func (h *Handler) list(c *gin.Context) {
	q, err := ParseListQuery(c.Request.URL.Query())
	if err != nil {
		httpx.WriteError(c, err)
		return
	}
	result, hit, err := h.svc.List(c.Request.Context(), q)
	if err != nil {
		httpx.WriteError(c, err)
		return
	}
	if hit {
		c.Header("X-Cache", "HIT")
	} else {
		c.Header("X-Cache", "MISS")
	}
	c.JSON(http.StatusOK, result)
}

// @Summary	Get a task
// @Tags		tasks
// @Produce	json
// @Param		id	path		int	true	"Task ID"
// @Success	200	{object}	TaskResponse
// @Failure	400	{object}	httpx.ErrorBody
// @Failure	404	{object}	httpx.ErrorBody
// @Failure	500	{object}	httpx.ErrorBody
// @Router		/tasks/{id} [get]
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
	c.JSON(http.StatusOK, TaskResponse{Data: *t})
}

type UpdateTaskRequest struct {
	Title       string  `json:"title" binding:"required,max=255"`
	Description *string `json:"description" binding:"omitempty,max=10000"`
	Status      Status  `json:"status" binding:"required,oneof=todo in_progress done"`
	Assignee    *string `json:"assignee" binding:"omitempty,max=100"`
}

// @Summary	Update a task
// @Tags		tasks
// @Accept		json
// @Produce	json
// @Param		id		path		int					true	"Task ID"
// @Param		body	body		UpdateTaskRequest	true	"New task values"
// @Success	200		{object}	TaskResponse
// @Failure	400		{object}	httpx.ErrorBody
// @Failure	404		{object}	httpx.ErrorBody
// @Failure	409		{object}	httpx.ErrorBody
// @Failure	413		{object}	httpx.ErrorBody
// @Failure	500		{object}	httpx.ErrorBody
// @Router		/tasks/{id} [put]
func (h *Handler) update(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		httpx.WriteError(c, err)
		return
	}
	var req UpdateTaskRequest
	if !bindJSON(c, &req) {
		return
	}
	t, err := h.svc.Update(c.Request.Context(), id, UpdateInput(req))
	if err != nil {
		httpx.WriteError(c, err)
		return
	}
	c.JSON(http.StatusOK, TaskResponse{Data: *t})
}

// @Summary	Delete a task
// @Tags		tasks
// @Param		id	path	int	true	"Task ID"
// @Success	204
// @Failure	400	{object}	httpx.ErrorBody
// @Failure	404	{object}	httpx.ErrorBody
// @Failure	500	{object}	httpx.ErrorBody
// @Router		/tasks/{id} [delete]
func (h *Handler) delete(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		httpx.WriteError(c, err)
		return
	}
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		httpx.WriteError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func parseID(c *gin.Context) (uint64, error) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		return 0, &apperr.ValidationError{Details: map[string]string{"id": "must be a positive integer"}}
	}
	return id, nil
}

// bindJSON writes the error response and returns false when the body is invalid.
func bindJSON(c *gin.Context, req any) bool {
	err := c.ShouldBindJSON(req)
	if err == nil {
		return true
	}
	if verr := apperr.FromBinding(err); verr != nil {
		httpx.WriteError(c, verr)
	} else {
		httpx.WriteError(c, err)
	}
	return false
}
