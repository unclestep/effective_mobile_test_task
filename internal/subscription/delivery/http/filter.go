package http

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"

	"em/internal/subscription"
)

func parseFilter(c *gin.Context) (*subscription.SubscriptionFilter, error) {
	var filter subscription.SubscriptionFilter

	if v := c.Query("user_id"); v != "" {
		filter.UserID = &v
	}
	if v := c.Query("service_name"); v != "" {
		filter.ServiceName = &v
	}
	if v := c.Query("period_from"); v != "" {
		t, err := time.Parse(monthYearLayout, v)
		if err != nil {
			return nil, fmt.Errorf("invalid period_from %q, expected MM-YYYY", v)
		}
		filter.PeriodFrom = &t
	}
	if v := c.Query("period_to"); v != "" {
		t, err := time.Parse(monthYearLayout, v)
		if err != nil {
			return nil, fmt.Errorf("invalid period_to %q, expected MM-YYYY", v)
		}
		filter.PeriodTo = &t
	}

	return &filter, nil
}
