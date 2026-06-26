package http

import (
	"em/internal/subscription"
)

type SubscriptionResponse struct {
	SubscriptionID string     `json:"subscription_id"`
	ServiceName    string     `json:"service_name"`
	Price          int        `json:"price"`
	UserID         string     `json:"user_id"`
	StartDate      MonthYear  `json:"start_date"`
	EndDate        *MonthYear `json:"end_date,omitempty"`
}

func toResponse(sub *subscription.Subscription) SubscriptionResponse {
	resp := SubscriptionResponse{
		SubscriptionID: sub.SubscriptionID,
		ServiceName:    sub.ServiceName,
		Price:          sub.Price,
		UserID:         sub.UserID,
		StartDate:      MonthYear(sub.StartDate),
	}
	if sub.EndDate != nil {
		end := MonthYear(*sub.EndDate)
		resp.EndDate = &end
	}
	return resp
}

func toResponseList(subs []*subscription.Subscription) []SubscriptionResponse {
	out := make([]SubscriptionResponse, 0, len(subs))
	for _, s := range subs {
		out = append(out, toResponse(s))
	}
	return out
}

type TotalCostResponse struct {
	TotalCost int `json:"total_cost"`
}
