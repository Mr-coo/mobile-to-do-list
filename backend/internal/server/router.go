package server

import (
	"fmt"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"taskmanager/internal/httpx"
	"taskmanager/internal/task"
)

func New(h *task.Handler, corsOrigins []string) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.CustomRecovery(func(c *gin.Context, recovered any) {
		httpx.WriteError(c, fmt.Errorf("panic: %v", recovered))
	}))
	r.HandleMethodNotAllowed = true
	r.NoRoute(func(c *gin.Context) {
		httpx.Write(c, http.StatusNotFound, "NOT_FOUND", "Route not found", nil)
	})
	r.NoMethod(func(c *gin.Context) {
		httpx.Write(c, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed", nil)
	})

	// cors.New panics on an empty origin list.
	if len(corsOrigins) > 0 {
		r.Use(cors.New(cors.Config{
			AllowOrigins:  corsOrigins,
			AllowMethods:  []string{"GET", "POST", "PUT", "DELETE"},
			AllowHeaders:  []string{"Content-Type"},
			ExposeHeaders: []string{"X-Cache"},
		}))
	}

	h.Register(r.Group("/api"))
	return r
}
