package subscription

import (
	"time"
)

type Subscription struct {
	SubscriptionID string     `json:"subscription_id"`
	ServiceName    string     `json:"service_name"`
	Price          int        `json:"price"`
	UserID         string     `json:"user_id"`
	StartDate      time.Time  `json:"start_date"`
	EndDate        *time.Time `json:"end_date,omitempty"`
}

func NewSubscription(id, name string, price int, userID string, start time.Time, end *time.Time) *Subscription {
	return &Subscription{
		SubscriptionID: id,
		ServiceName:    name,
		Price:          price,
		UserID:         userID,
		StartDate:      start,
		EndDate:        end,
	}
}
