package http

import (
	"em/internal/subscription"
)

type CreateSubscriptionRequest struct {
	ServiceName string     `json:"service_name" binding:"required"`
	Price       int        `json:"price" binding:"required,gt=0"`
	UserID      string     `json:"user_id" binding:"required,uuid"`
	StartDate   MonthYear  `json:"start_date" binding:"required"`
	EndDate     *MonthYear `json:"end_date,omitempty"`
}

func (r CreateSubscriptionRequest) toInput() *subscription.CreateSubInput {
	input := &subscription.CreateSubInput{
		ServiceName: r.ServiceName,
		Price:       r.Price,
		UserID:      r.UserID,
		StartDate:   r.StartDate.Time(),
	}
	if r.EndDate != nil {
		end := r.EndDate.Time()
		input.EndDate = &end
	}
	return input
}

type UpdateSubscriptionRequest struct {
	ServiceName *string    `json:"service_name,omitempty"`
	Price       *int       `json:"price,omitempty" binding:"omitempty,gt=0"`
	StartDate   *MonthYear `json:"start_date,omitempty"`
	EndDate     *MonthYear `json:"end_date,omitempty"`
}

func (r UpdateSubscriptionRequest) toInput() *subscription.UpdateSubInput {
	input := &subscription.UpdateSubInput{
		ServiceName: r.ServiceName,
		Price:       r.Price,
	}
	if r.StartDate != nil {
		start := r.StartDate.Time()
		input.StartDate = &start
	}
	if r.EndDate != nil {
		end := r.EndDate.Time()
		input.EndDate = &end
	}
	return input
}
