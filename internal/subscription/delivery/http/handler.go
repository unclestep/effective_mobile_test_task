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

// Create godoc
// @Summary Create a subscription
// @Description Creates a new subscription record
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param subscription body CreateSubscriptionRequest true "Subscription data"
// @Success 201 {object} SubscriptionResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /subscriptions [post]
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

// GetByID godoc
// @Summary Get a subscription by id
// @Description Returns a subscription record by ID
// @Tags subscriptions
// @Produce json
// @Param id path string true "Subscription ID"
// @Success 200 {object} SubscriptionResponse
// @Failure 404 {object} map[string]string
// @Router /subscriptions/{id} [get]
func (h *Handler) GetByID(c *gin.Context) {
	sub, err := h.uc.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, toResponse(sub))
}

// List godoc
// @Summary List subscriptions
// @Description Returns a list of subscriptions with the optional filtering
// @Tags subscriptions
// @Produce json
// @Param user_id query string false "Filter by user id"
// @Param service_name query string false "Filter by service name"
// @Param period_from query string false "Period start, MM-YYYY"
// @Param period_to query string false "Period end, MM-YYYY"
// @Success 200 {array} SubscriptionResponse
// @Router /subscriptions [get]
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

// Summary godoc
// @Summary Total cost of subscriptions
// @Description Counts a total cost of subscriptions with the filtering
// @Tags subscriptions
// @Produce json
// @Param user_id query string false "Filter by user id"
// @Param service_name query string false "Filter by service name"
// @Param period_from query string false "Period start, MM-YYYY"
// @Param period_to query string false "Period end, MM-YYYY"
// @Success 200 {object} TotalCostResponse
// @Router /subscriptions/summary [get]
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

// Update godoc
// @Summary Update a subscription
// @Description Updates a subscription record
// @Tags subscriptions
// @Accept json
// @Param id path string true "Subscription ID"
// @Param subscription body UpdateSubscriptionRequest true "Fields to update"
// @Success 204
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /subscriptions/{id} [patch]
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

// Delete godoc
// @Summary Delete a subscription
// @Description Deletes a subscription record
// @Tags subscriptions
// @Param id path string true "Subscription ID"
// @Success 204
// @Failure 404 {object} map[string]string
// @Router /subscriptions/{id} [delete]
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
