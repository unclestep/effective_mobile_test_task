package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"em/internal/subscription"
)

type Handler struct {
	uc subscription.UseCase
}

func NewHandler(uc subscription.UseCase) *Handler {
	return &Handler{uc: uc}
}

func (h *Handler) Create(c *gin.Context) {
	var req CreateSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sub, err := h.uc.Create(c.Request.Context(), req.toInput())
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusCreated, toResponse(sub))
}

func (h *Handler) GetByID(c *gin.Context) {
	id := c.Param("id")
	if id == "summary" {
		h.Summary(c)
		return
	}

	sub, err := h.uc.GetByID(c.Request.Context(), id)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, toResponse(sub))
}

func (h *Handler) List(c *gin.Context) {
	filter, err := parseFilter(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	subs, err := h.uc.Get(c.Request.Context(), filter)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, toResponseList(subs))
}

func (h *Handler) Summary(c *gin.Context) {
	filter, err := parseFilter(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	total, err := h.uc.GetTotalCost(c.Request.Context(), filter)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, TotalCostResponse{TotalCost: total})
}

func (h *Handler) Update(c *gin.Context) {
	var req UpdateSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.uc.Update(c.Request.Context(), c.Param("id"), req.toInput()); err != nil {
		respondError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) Delete(c *gin.Context) {
	if err := h.uc.Delete(c.Request.Context(), c.Param("id")); err != nil {
		respondError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func respondError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, subscription.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, subscription.ErrNothingToUpdate):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}
