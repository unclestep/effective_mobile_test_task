package http

import "github.com/gin-gonic/gin"

func RegisterRoutes(rg *gin.RouterGroup, h *Handler) {
	subs := rg.Group("/subscriptions")
	subs.POST("", h.Create)
	subs.GET("", h.List)
	subs.GET("/summary", h.Summary)
	subs.GET("/:id", h.GetByID)
	subs.PATCH("/:id", h.Update)
	subs.DELETE("/:id", h.Delete)
}
