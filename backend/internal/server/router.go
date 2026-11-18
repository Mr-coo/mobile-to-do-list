package server

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"taskmanager/internal/task"
)

func New(h *task.Handler, corsOrigins []string) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

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
